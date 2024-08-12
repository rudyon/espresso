package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Function to execute a shell command
func executeCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing command: %s\n", err)
	}
	return err
}

// Function to install from a .bean file
func installFromBean(beanPath string) error {
	// Ensure the /etc/espresso directory exists
	err := os.MkdirAll("/etc/espresso", 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Download the .bean file
	response, err := http.Get("https://raw.githubusercontent.com/rudyon/espresso/main/beans/" + beanPath)
	if err != nil {
		return fmt.Errorf("error downloading .bean file: %v", err)
	}
	defer response.Body.Close()

	// Check if the response status is 200 OK
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error: received status code %d", response.StatusCode)
	}

	// Write the .bean file to /etc/espresso
	filePath := filepath.Join("/etc/espresso", beanPath)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating .bean file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("error writing .bean file: %v", err)
	}

	// Make the .bean file executable
	err = os.Chmod(filePath, 0755)
	if err != nil {
		return fmt.Errorf("error setting file permissions: %v", err)
	}

	// Execute the .bean file as a shell script
	return executeCommand("/bin/bash", filePath)
}

// Parse dependencies and commands from a .bean file
func parseBeanFile(filePath string) (dependencies []string, commands []string, err error) {
	// Open the .bean file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening .bean file: %v", err)
	}
	defer file.Close()

	// Read the .bean file
	var parsingDependencies bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "depends=(") {
			// Extract dependencies
			parsingDependencies = true
			line = strings.TrimPrefix(line, "depends=(")
			line = strings.TrimSuffix(line, ")")
			line = strings.ReplaceAll(line, `"`, "") // Remove quotes
			dependencies = strings.Split(line, " ")
		} else if parsingDependencies && strings.HasSuffix(line, ")") {
			// End of dependencies
			parsingDependencies = false
		} else if parsingDependencies {
			// Continue reading dependencies
			line = strings.Trim(line, `"`)
			dependencies = append(dependencies, line)
		} else {
			// Collect commands
			commands = append(commands, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading .bean file: %v", err)
	}

	return dependencies, commands, nil
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root.")
		return
	}

	if len(os.Args) < 3 || os.Args[1] != "brew" {
		fmt.Println("Usage: espresso brew <package>")
		return
	}

	packageName := os.Args[2] + ".bean"

	// Parse dependencies and commands from the main package file
	dependenciesFilePath := filepath.Join("/etc/espresso", packageName)
	dependencies, commands, err := parseBeanFile(dependenciesFilePath)
	if err != nil {
		fmt.Printf("error parsing .bean file: %v\n", err)
		return
	}

	// Install each dependency
	for _, dep := range dependencies {
		fmt.Printf("Checking and installing dependency: %s\n", dep)
		if err := installFromBean(dep + ".bean"); err != nil {
			fmt.Printf("error installing dependency %s: %v\n", dep, err)
			return
		}
	}

	// Create a temporary script file to execute commands
	scriptPath := filepath.Join("/etc/espresso", "temp_script.sh")
	scriptFile, err := os.Create(scriptPath)
	if err != nil {
		fmt.Printf("error creating script file: %v\n", err)
		return
	}
	defer scriptFile.Close()

	for _, command := range commands {
		if _, err := scriptFile.WriteString(command + "\n"); err != nil {
			fmt.Printf("error writing command to script: %v\n", err)
			return
		}
	}

	if err := os.Chmod(scriptPath, 0755); err != nil {
		fmt.Printf("error setting script file permissions: %v\n", err)
		return
	}

	if err := executeCommand("/bin/bash", scriptPath); err != nil {
		fmt.Printf("error executing script: %v\n", err)
		return
	}

	fmt.Println("Installation complete!")
}
