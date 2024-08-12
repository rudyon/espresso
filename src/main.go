package main

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

func ensureEspressoDir() error {
    dir := "/etc/espresso"
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("error creating /etc/espresso directory: %v", err)
        }
    }
    return nil
}

func downloadBean(beanFile string) (string, error) {
    url := fmt.Sprintf("https://raw.githubusercontent.com/rudyon/espresso/main/beans/%s.bean", beanFile)
    output := filepath.Join("/etc/espresso", beanFile)

    fmt.Printf("Downloading %s from %s\n", beanFile, url)

    cmd := exec.Command("wget", url, "-O", output)
    outputBytes, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("error downloading .bean file: %v\nOutput: %s", err, outputBytes)
    }

    fmt.Printf("Downloaded .bean file to: %s\n", output)
    return output, nil
}

func parseDependencies(filePath string) ([]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var dependencies []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "Depends=(") {
            deps := strings.TrimPrefix(line, "Depends=(")
            deps = strings.TrimSuffix(deps, ")")
            dependencies = strings.Split(deps, " ")
            break
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return dependencies, nil
}

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

    fmt.Printf("Running the .bean script: %s\n", filePath)

    // Parse dependencies and install them
    dependencies, err := parseDependencies(filePath)
    if err != nil {
        return fmt.Errorf("error parsing dependencies: %v", err)
    }
    for _, dep := range dependencies {
        fmt.Printf("Checking and installing dependency: %s\n", dep)
        // Recursively install dependencies
        if err := installFromBean(dep + ".bean"); err != nil {
            return fmt.Errorf("error installing dependency %s: %v", dep, err)
        }
    }

    // Execute the .bean script directly
    cmd := exec.Command("bash", filePath)
    cmd.Dir = "/etc/espresso"
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
        return fmt.Errorf("error executing .bean file: %v", err)
    }

    return nil
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: espresso <command> <package>")
        os.Exit(1)
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
            os.Exit(1)
        }
        if err := installFromBean(packageName); err != nil {
            fmt.Printf("Error: %v\n", err)
            os.Exit(1)
        }
    case "drop":
        // Implement removal logic if necessary
        fmt.Println("Remove command not implemented.")
        os.Exit(1)
    case "look":
        // Implement listing logic if necessary
        fmt.Println("Look command not implemented.")
        os.Exit(1)
    default:
        fmt.Println("Error: Unknown command.")
        os.Exit(1)
    }
}

