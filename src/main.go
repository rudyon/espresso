package main

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

func readPackages() []string {
    file, err := os.Open("packages.txt")
    if err != nil {
        if os.IsNotExist(err) {
            return []string{}
        }
        fmt.Println("Error reading packages file:", err)
        return nil
    }
    defer file.Close()

    var packages []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        packages = append(packages, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading file:", err)
    }
    return packages
}

func writePackages(packages []string) {
    file, err := os.Create("packages.txt")
    if err != nil {
        fmt.Println("Error creating packages file:", err)
        return
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
    for _, pkg := range packages {
        writer.WriteString(pkg + "\n")
    }
    writer.Flush()
}

func install(packageName string) {
    packages := readPackages()
    for _, pkg := range packages {
        if pkg == packageName {
            fmt.Println("Package already installed:", packageName)
            return
        }
    }

    if packageName == "nano" {
        if err := installFromBean("nano.bean"); err != nil {
            fmt.Println("Error installing from .bean file:", err)
            return
        }
        packages = append(packages, packageName)
        writePackages(packages)
        fmt.Printf("Installing %s...\n", packageName)
        return
    }

    fmt.Println("Error: Unknown package or package file not found.")
}

func remove(packageName string) {
    packages := readPackages()
    var newPackages []string
    found := false
    for _, pkg := range packages {
        if pkg != packageName {
            newPackages = append(newPackages, pkg)
        } else {
            found = true
        }
    }
    if !found {
        fmt.Println("Package not found:", packageName)
        return
    }
    writePackages(newPackages)
    fmt.Printf("Removing %s...\n", packageName)
}

func listPackages() {
    packages := readPackages()
    if len(packages) == 0 {
        fmt.Println("No packages installed.")
        return
    }
    fmt.Println("Installed packages:")
    for _, pkg := range packages {
        fmt.Println(pkg)
    }
}

func installFromBean(beanFile string) error {
    file, err := os.Open(beanFile)
    if err != nil {
        return fmt.Errorf("error reading .bean file: %v", err)
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

