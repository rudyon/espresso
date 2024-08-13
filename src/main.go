package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	baseURL = "https://raw.githubusercontent.com/rudyon/espresso/main/"
	beansURL = baseURL + "beans/"
	espressoSourceURL = baseURL + "main.go"
)

// Function to execute a shell command
func executeCommand(command string) error {
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing command: %s\n", err)
	}
	return err
}

// Function to download a file given its URL
func downloadFile(filePath, url string) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading file: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error: received status code %d", response.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// Function to download a .bean file given its name
func downloadBean(beanName string) error {
	err := os.MkdirAll("/etc/espresso", 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	filePath := filepath.Join("/etc/espresso", beanName)
	err = downloadFile(filePath, beansURL+beanName)
	if err != nil {
		return err
	}

	err = os.Chmod(filePath, 0755)
	if err != nil {
		return fmt.Errorf("error setting file permissions: %v", err)
	}

	return nil
}

// Parse dependencies from a .bean file
func parseDependencies(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening .bean file: %v", err)
	}
	defer file.Close()

	dependsPattern := regexp.MustCompile(`^depends=\(([^)]+)\)`)

	var dependencies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			if matches := dependsPattern.FindStringSubmatch(line); len(matches) > 1 {
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

// Function to update espresso
func updateEspresso() error {
	fmt.Println("Updating espresso...")
	
	// Download the latest source code
	tempFile := "/tmp/espresso_new.go"
	err := downloadFile(tempFile, espressoSourceURL)
	if err != nil {
		return fmt.Errorf("error downloading new espresso source: %v", err)
	}

	// Compile the new source
	err = executeCommand(fmt.Sprintf("go build -o /tmp/espresso_new %s", tempFile))
	if err != nil {
		return fmt.Errorf("error compiling new espresso source: %v", err)
	}

	// Replace the current binary
	err = executeCommand("mv /tmp/espresso_new /usr/local/bin/espresso")
	if err != nil {
		return fmt.Errorf("error replacing espresso binary: %v", err)
	}

	fmt.Println("espresso has been updated successfully!")
	return nil
}

// Function to remove a package
func removePackage(packageName string) error {
	fmt.Printf("Removing package %s...\n", packageName)
	
	beanFile := filepath.Join("/etc/espresso", packageName+".bean")
	
	// Check if the .bean file exists
	if _, err := os.Stat(beanFile); os.IsNotExist(err) {
		return fmt.Errorf("package %s is not installed", packageName)
	}

	// Execute the .bean file with the "remove" argument
	err := executeCommand(beanFile + " remove")
	if err != nil {
		return fmt.Errorf("error removing package %s: %v", packageName, err)
	}

	// Remove the .bean file
	err = os.Remove(beanFile)
	if err != nil {
		return fmt.Errorf("error removing .bean file for %s: %v", packageName, err)
	}

	fmt.Printf("Package %s has been removed successfully!\n", packageName)
	return nil
}

// Function to list or search for packages
func lookPackages(searchTerm string) error {
	fmt.Println("Searching for packages...")

	// Fetch the list of .bean files from the GitHub repository
	response, err := http.Get(beansURL)
	if err != nil {
		return fmt.Errorf("error fetching package list: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading package list: %v", err)
	}

	// Parse the HTML content to extract .bean file names
	re := regexp.MustCompile(`href="([^"]+\.bean)"`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	if len(matches) == 0 {
		fmt.Println("No packages found.")
		return nil
	}

	fmt.Println("Available packages:")
	for _, match := range matches {
		packageName := match[1]
		if searchTerm == "" || strings.Contains(packageName, searchTerm) {
			fmt.Println(packageName)
		}
	}

	return nil
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root.")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: espresso <command> [arguments]")
		fmt.Println("Commands: brew, update, remove, look")
		return
	}

	command := os.Args[1]

	switch command {
	case "brew":
		if len(os.Args) < 3 {
			fmt.Println("Usage: espresso brew <package>")
			return
		}
		packageName := os.Args[2] + ".bean"
		
		fmt.Printf("Downloading %s...\n", packageName)
		if err := downloadBean(packageName); err != nil {
			fmt.Printf("Error downloading package %s: %v\n", packageName, err)
			return
		}

		// Parse dependencies
		dependenciesFilePath := filepath.Join("/etc/espresso", packageName)
		dependencies, err := parseDependencies(dependenciesFilePath)
		if err != nil {
			fmt.Printf("Error parsing dependencies: %v\n", err)
			return
		}

		// Download and install dependencies
		fmt.Println("Downloading and installing dependencies...")
		for _, dep := range dependencies {
			fmt.Printf("Downloading %s...\n", dep)
			if err := downloadBean(dep); err != nil {
				fmt.Printf("Error downloading dependency %s: %v\n", dep, err)
				return
			}
			
			depFilePath := filepath.Join("/etc/espresso", dep)
			fmt.Printf("Installing dependency: %s\n", dep)
			if err := executeCommand(depFilePath); err != nil {
				fmt.Printf("Error installing dependency %s: %v\n", dep, err)
				return
			}
		}

		// Install the main package
		fmt.Printf("Installing package: %s\n", packageName)
		if err := executeCommand(dependenciesFilePath); err != nil {
			fmt.Printf("Error installing package %s: %v\n", packageName, err)
			return
		}

		fmt.Println("Installation complete!")

	case "update":
		err := updateEspresso()
		if err != nil {
			fmt.Printf("Error updating espresso: %v\n", err)
		}

	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: espresso remove <package>")
			return
		}
		packageName := os.Args[2]
		err := removePackage(packageName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "look":
		searchTerm := ""
		if len(os.Args) >= 3 {
			searchTerm = os.Args[2]
		}
		err := lookPackages(searchTerm)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: espresso <command> [arguments]")
		fmt.Println("Commands: brew, update, remove, look")
	}
}


