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

// ... (keep all the existing functions as they are)

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

	// Download the main package file first
	fmt.Printf("Downloading %s...\n", packageName)
	if err := downloadBean(packageName, downloadURL(packageName)); err != nil {
		fmt.Printf("error downloading package %s: %v\n", packageName, err)
		return
	}

	// Parse dependencies after downloading the main package
	dependenciesFilePath := filepath.Join("/etc/espresso", packageName)
	if _, err := os.Stat(dependenciesFilePath); os.IsNotExist(err) {
		fmt.Printf("error: main package file %s does not exist\n", dependenciesFilePath)
		return
	}

	dependencies, err := parseDependencies(dependenciesFilePath)
	if err != nil {
		fmt.Printf("error parsing dependencies: %v\n", err)
		return
	}

	// Download dependencies
	fmt.Println("Downloading dependencies...")
	for _, dep := range dependencies {
		fmt.Printf("Downloading %s...\n", dep)
		if err := downloadBean(dep, downloadURL(dep)); err != nil {
			fmt.Printf("error downloading dependency %s: %v\n", dep, err)
			return
		}
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
