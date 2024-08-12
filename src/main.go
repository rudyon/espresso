package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Function to execute a shell command
func executeCommand(cmd string) error {
	// Split the command and arguments
	cmdArgs := strings.Fields(cmd)
	if len(cmdArgs) == 0 {
		return errors.New("no command provided")
	}

	// Check if the command is a script and needs special handling
	if strings.HasSuffix(cmdArgs[0], ".sh") {
		// Ensure the script has execute permissions
		if err := os.Chmod(cmdArgs[0], 0755); err != nil {
			return fmt.Errorf("failed to set execute permissions: %w", err)
		}
		// Use /bin/bash to run the script
		cmdArgs = append([]string{"/bin/bash"}, cmdArgs...)
	}

	// Create the exec.Command with the arguments
	command := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
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
	response, err := http.Get("https://raw.githubusercontent.com/rudyon/espresso/tree/main/beans/" + beanPath)
	if err != nil {
		return fmt.Errorf("error downloading .bean file: %v", err)
	}
	defer response.Body.Close()

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
	cmd := fmt.Sprintf("/bin/bash %s", filePath)
	fmt.Println("Running command:", cmd)
	return executeCommand(cmd)
}

// Parse dependencies from a file
func parseDependencies(filePath string) ([]string, error) {
	// For example purposes, we'll return hardcoded dependencies
	// Replace this with actual file reading logic
	return []string{"ncurses.bean", "libmagic.bean"}, nil
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root.")
		return
	}

	filePath := "dependencies.txt" // Example file path

	// Process dependencies
	dependencies, err := parseDependencies(filePath)
	if err != nil {
		fmt.Printf("error parsing dependencies: %v\n", err)
		return
	}

	for _, dep := range dependencies {
		fmt.Printf("Checking and installing dependency: %s\n", dep)
		if err := installFromBean(dep); err != nil {
			fmt.Printf("error installing dependency %s: %v\n", dep, err)
			return
		}
	}

	// Install the main package
	mainPackage := "nano.bean" // Replace this with the actual package specified by the user
	fmt.Printf("Installing main package: %s\n", mainPackage)
	if err := installFromBean(mainPackage); err != nil {
		fmt.Printf("error installing main package %s: %v\n", mainPackage, err)
	}
}

