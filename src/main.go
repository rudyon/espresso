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

	// Build the URL to download the .bean file
	url := "https://raw.githubusercontent.com/rudyon/espresso/main/beans/" + beanPath
	fmt.Printf("Downloading .bean file from URL: %s\n", url)

	// Download the .bean file
	response, err := http.Get(url)
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

	// Define regex pattern for dependencies line
	dependsPattern := regexp.MustCompile(`^depends=\(([^)]+)\)`)
	
	// Read dependencies
	var dependencies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// Check if line contains dependencies
			if matches := dependsPattern.FindStringSubmatch(line); len(matches) > 1 {
				// Split the dependencies by space and remove quotes
				deps := strings.Fields(matches[1])
				for _, dep := range deps {
					dependency := strings.Trim(dep, `"`)
					dependencies = append(dependencies, dependency + ".bean")
				}
			}
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

	// Parse dependencies from the main package file
	dependenciesFilePath := filepath.Join("/etc/espresso", packageName)
	dependencies, err := parseDependencies(dependenciesFilePath)
	if err != nil {
		fmt.Printf("error parsing dependencies: %v\n", err)
		return
	}

	// Define a function to download a .bean file
	downloadBean := func(beanName string) error {
		url := "https://example.com/beans/" + beanName
		fmt.Printf("Downloading %s...\n", beanName)
		return installFromBean(beanName)
	}

	// Download all .bean files (dependencies and the main package)
	// Ensure main package is downloaded last
	for _, dep := range dependencies {
		if err := downloadBean(dep); err != nil {
			fmt.Printf("error downloading dependency %s: %v\n", dep, err)
			return
		}
	}

	// Download the main package file
	if err := downloadBean(packageName); err != nil {
		fmt.Printf("error downloading package %s: %v\n", packageName, err)
		return
	}

	// Install each .bean file
	fmt.Println("Installing packages...")
	for _, dep := range dependencies {
		fmt.Printf("Installing dependency: %s\n", dep)
		if err := installFromBean(dep); err != nil {
			fmt.Printf("error installing dependency %s: %v\n", dep, err)
			return
		}
	}

	// Install the main package
	fmt.Printf("Installing package: %s\n", packageName)
	if err := installFromBean(packageName); err != nil {
		fmt.Printf("error installing package %s: %v\n", packageName, err)
		return
	}

	fmt.Println("Installation complete!")
}


