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
func executeCommand(command string) error {
	cmd := exec.Command("/bin/bash", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing command: %s\n", err)
	}
	return err
}

// Function to download a .bean file given its URL
func downloadBean(beanName, url string) error {
	// Ensure the /etc/espresso directory exists
	err := os.MkdirAll("/etc/espresso", 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

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
	filePath := filepath.Join("/etc/espresso", beanName)
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

	return nil
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
					dependencies = append(dependencies, dependency+".bean")
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
	baseURL := "https://example.com/beans/" // Replace with the actual base URL

	// Define a function to download a .bean file
	downloadURL := func(beanName string) string {
		return baseURL + beanName
	}

	// Download all .bean files (dependencies and the main package)
	fmt.Println("Downloading dependencies and package...")

	dependenciesFilePath := filepath.Join("/etc/espresso", packageName)
	dependencies, err := parseDependencies(dependenciesFilePath)
	if err != nil {
		fmt.Printf("error parsing dependencies: %v\n", err)
		return
	}

	// Download all .bean files (dependencies and the main package)
	for _, dep := range dependencies {
		fmt.Printf("Downloading %s...\n", dep)
		if err := downloadBean(dep, downloadURL(dep)); err != nil {
			fmt.Printf("error downloading dependency %s: %v\n", dep, err)
			return
		}
	}

	// Download the main package file
	fmt.Printf("Downloading %s...\n", packageName)
	if err := downloadBean(packageName, downloadURL(packageName)); err != nil {
		fmt.Printf("error downloading package %s: %v\n", packageName, err)
		return
	}

	// Install each .bean file
	fmt.Println("Installing packages...")
	for _, dep := range dependencies {
		depFilePath := filepath.Join("/etc/espresso", dep)
		if _, err := os.Stat(depFilePath); err == nil {
			fmt.Printf("Installing dependency: %s\n", dep)
			if err := executeCommand(depFilePath); err != nil {
				fmt.Printf("error installing dependency %s: %v\n", dep, err)
				return
			}
		} else {
			fmt.Printf("error: dependency file %s does not exist\n", depFilePath)
			return
		}
	}

	// Install the main package
	mainPackagePath := filepath.Join("/etc/espresso", packageName)
	if _, err := os.Stat(mainPackagePath); err == nil {
		fmt.Printf("Installing package: %s\n", packageName)
		if err := executeCommand(mainPackagePath); err != nil {
			fmt.Printf("error installing package %s: %v\n", packageName, err)
			return
		}
	} else {
		fmt.Printf("error: main package file %s does not exist\n", mainPackagePath)
		return
	}

	fmt.Println("Installation complete!")
}

