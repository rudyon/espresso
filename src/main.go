package main

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
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
    url := fmt.Sprintf("https://github.com/rudyon/espresso/blob/main/beans/%s", beanFile)
    output := fmt.Sprintf("/etc/espresso/%s", beanFile)
    cmd := exec.Command("wget", url, "-O", output)
    cmd.Env = append(os.Environ(), "PATH=/usr/local/bin:/usr/bin:/bin")
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("error downloading .bean file: %v", err)
    }
    return output, nil
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

    fmt.Printf("Downloaded .bean file to: %s\n", filePath)

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
        // Run the command using a shell
        cmd := exec.Command("bash", "-c", cmdStr)
        cmd.Dir = "/etc/espresso" // Ensure commands are run in the correct directory
        cmd.Env = append(os.Environ(), "PATH=/usr/local/bin:/usr/bin:/bin") // Ensure PATH is set
        output, err := cmd.CombinedOutput()
        if err != nil {
            return fmt.Errorf("error executing command '%s': %v\nOutput: %s", cmdStr, err, output)
        }
        fmt.Printf("Output:\n%s\n", output)
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

