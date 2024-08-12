package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

	// Read and process the .bean file
	return processBeanFile(filePath)
}

// Function to process the .bean file using regex
func processBeanFile(beanFilePath string) error {
	file, err := os.Open(beanFilePath)
	if err != nil {
		return fmt.Errorf("error opening .bean file: %v", err)
	}
	defer file.Close()

	var dependencies []string
	var commands []string

	// Regex patterns for dependency and command extraction
	depPattern := regexp.MustCompile(`^depends:\s*(.+)$`)
	cmdPattern := regexp.MustCompile(`^[^#].*`)

	scanner := bufio.NewScanner(file)
	inCommandsSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		// Check for dependencies section
		if matches := depPattern.FindStringSubmatch(line); matches != nil {
			depLines := strings.Split(matches[1], " ")
			for _, dep := range depLines {
				dependencies = append(dependencies, dep+".bean")
			}
			continue
		}

		// Check for commands
		if cmdPattern.MatchString(line) {
			if !inCommandsSection {
				inCommandsSection = true
			}
			// Collect commands
			commands = append(commands, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .bean file: %v", err)
	}

	// Install dependencies
	for _, dep := range dependencies {
		fmt.Printf("Installing dependency: %s\n", dep)
		if err := installFromBean(dep); err != nil {
			fmt.Printf("error installing dependency %s: %v\n", dep, err)
			return err
		}
	}

	// Execute commands
	for _, command := range commands {
		fmt.Printf("Executing command: %s\n", command)
		if err := executeCommand("/bin/bash", "-c", command); err != nil {
			fmt.Printf("error executing command %s: %v\n", command, err)
			return err
		}
	}

	return nil
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

	// Install the main package
	fmt.Printf("Installing package: %s\n", packageName)
	if err := installFromBean(packageName); err != nil {
		fmt.Printf("error installing package %s: %v\n", packageName, err)
		return
	}

	fmt.Println("Installation complete!")
}
