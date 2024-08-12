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

// Parse dependencies from a .bean file
func parseDependencies(filePath string) ([]string, error) {
	// Open the .bean file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening .bean file: %v", err)
	}
	defer file.Close()

	// Read dependencies
	var dependencies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			dependencies = append(dependencies, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .bean file: %v", err)
	}

	return dependencies, nil
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

	// Parse and install dependencies
	dependenciesFilePath := filepath.Join("/etc/espresso", packageName)
	dependencies, err := parseDependencies(dependenciesFilePath)
	if err != nil {
		fmt.Printf("error parsing dependencies: %v\n", err)
		return
	}

	for _, dep := range dependencies {

