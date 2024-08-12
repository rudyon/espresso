package main

import (
    "bufio"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "strings"
)

const baseURL = "https://github.com/rudyon/espresso/raw/main/beans/"

func downloadBean(beanFile string) (string, error) {
    url := baseURL + beanFile
    response, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("error downloading .bean file: %v", err)
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        return "", fmt.Errorf("error: received status code %d", response.StatusCode)
    }

    tempFile, err := os.CreateTemp("", "espresso_*.bean")
    if err != nil {
        return "", fmt.Errorf("error creating temp file: %v", err)
    }
    defer tempFile.Close()

    _, err = io.Copy(tempFile, response.Body)
    if err != nil {
        return "", fmt.Errorf("error saving .bean file: %v", err)
    }

    return tempFile.Name(), nil
}

func installFromBean(beanFile string) error {
    filePath, err := downloadBean(beanFile)
    if err != nil {
        return err
    }
    defer os.Remove(filePath)

    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("error opening downloaded .bean file: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        fmt.Printf("Running command: %s\n", line)
        cmdParts := strings.Split(line, " ")
        cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
        output, err := cmd.CombinedOutput()
        if err != nil {
            return fmt.Errorf("error executing command '%s': %v\nOutput: %s", line, err, output)
        }
        fmt.Printf("Output:\n%s\n", output)
    }
    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading .bean file: %v", err)
    }
    return nil
}

func install(packageName string) {
    // Simulate reading from a list of installed packages
    if packageName == "nano" {
        if err := installFromBean("nano.bean"); err != nil {
            fmt.Println("Error installing from .bean file:", err)
            return
        }
        fmt.Printf("Installing %s...\n", packageName)
        return
    }

    fmt.Println("Error: Unknown package or package file not found.")
}

func remove(packageName string) {
    fmt.Printf("Removing %s...\n", packageName)
}

func listPackages() {
    fmt.Println("Listing installed packages...")
}

func printUsage() {
    fmt.Println("Usage:")
    fmt.Println("  espresso brew <package>   - Install a package")
    fmt.Println("  espresso drop <package>   - Remove a package")
    fmt.Println("  espresso look             - List installed packages")
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Error: Command not specified.")
        printUsage()
        return
    }

    command := os.Args[1]
    var packageName string

    if len(os.Args) > 2 {
        packageName = os.Args[2]
    }

    switch command {
    case "brew":
        if packageName == "" {
            fmt.Println("Error: No package specified for installation.")
            printUsage()
            return
        }
        install(packageName)
    case "drop":
        if packageName == "" {
            fmt.Println("Error: No package specified for removal.")
            printUsage()
            return
        }
        remove(packageName)
    case "look":
        if len(os.Args) > 2 {
            fmt.Println("Error: The 'look' command does not accept any arguments.")
            printUsage()
            return
        }
        listPackages()
    default:
        fmt.Println("Error: Unknown command.")
        printUsage()
    }
}

