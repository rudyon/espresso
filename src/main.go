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

// Function to ensure /etc/espresso directory exists
func ensureEspressoDir() error {
    dir := "/etc/espresso"
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        fmt.Printf("Directory %s does not exist. Creating...\n", dir)
        err := os.MkdirAll(dir, 0755) // Create directory with 0755 permissions
        if err != nil {
            return fmt.Errorf("error creating directory %s: %v", dir, err)
        }
    }
    return nil
}

// Function to check if running as root
func checkRoot() {
    if os.Geteuid() != 0 {
        fmt.Println("Error: This program must be run as root. Please use sudo.")
        os.Exit(1)
    }
}

// Download the .bean file from GitHub and return the file path
func downloadBean(beanFile string) (string, error) {
    url := fmt.Sprintf("https://github.com/rudyon/espresso/raw/main/beans/%s.bean", beanFile)
    response, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("error downloading .bean file: %v", err)
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        return "", fmt.Errorf("error: received non-200 response status code %d", response.StatusCode)
    }

    filePath := "/etc/espresso/" + beanFile + ".bean"
    outFile, err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("error creating file %s: %v", filePath, err)
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, response.Body)
    if err != nil {
        return "", fmt.Errorf("error saving .bean file: %v", err)
    }

    return filePath, nil
}

// Install from a .bean file
func installFromBean(beanFile string) error {
    // Ensure /etc/espresso directory exists
    if err := ensureEspressoDir(); err != nil {
        return err
    }

    filePath, err := downloadBean(beanFile)
    if err != nil {
        return err
    }
    defer os.Remove(filePath)

    fmt.Printf("Downloaded .bean file to: %s\n", filePath) // Debug output

    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("error opening downloaded .bean file: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var commands []string

    // Skip the shebang line
    firstLine := true

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        if firstLine {
            // Skip shebang line
            firstLine = false
            continue
        }
        commands = append(commands, line)
    }
    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading .bean file: %v", err)
    }

    for _, cmdStr := range commands {
        fmt.Printf("Running command: %s\n", cmdStr)
        cmdParts := strings.Split(cmdStr, " ")
        cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
        cmd.Dir = "/etc/espresso" // Ensure commands are run in the correct directory
        output, err := cmd.CombinedOutput()
        if err != nil {
            return fmt.Errorf("error executing command '%s': %v\nOutput: %s", cmdStr, err, output)
        }
        fmt.Printf("Output:\n%s\n", output)
    }
    return nil
}

func install(packageName string) {
    if packageName != "" {
        fmt.Printf("Installing %s...\n", packageName)
        if err := installFromBean(packageName); err != nil {
            fmt.Printf("Installation failed: %v\n", err)
        }
    } else {
        fmt.Println("Error: No package specified for installation.")
    }
}

func remove(packageName string) {
    if packageName != "" {
        fmt.Printf("Removing %s...\n", packageName)
    } else {
        fmt.Println("Error: No package specified for removal.")
    }
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
    // Check if the program is run as root
    checkRoot()

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

