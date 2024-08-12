package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

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

    // Write the .bean file to /etc/espresso
    filePath := filepath.Join("/etc/espresso", beanPath)
    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("error creating .bean file: %v", err)
    }
    defer file.Close()

    _, err = ioutil.ReadAll(file)
    if err != nil {
        return fmt.Errorf("error reading .bean file: %v", err)
    }

    // Execute the .bean file as a shell script
    cmd := fmt.Sprintf("bash /etc/espresso/%s", beanPath)
    fmt.Println("Running command:", cmd)
    return executeCommand(cmd)
}

// Execute a shell command
func executeCommand(cmd string) error {
    // Here you would execute the command
    // Placeholder implementation
    fmt.Println("Executing:", cmd)
    return nil
}

// Parse dependencies from a file
func parseDependencies(filePath string) ([]string, error) {
    // For example purposes, we'll return hardcoded dependencies
    return []string{"ncurses.bean", "libmagic.bean"}, nil
}

func main() {
    filePath := "dependencies.txt" // Example file path

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
}
