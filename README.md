# Espresso

**Espresso** is a package management tool designed for Coffee Linux. It simplifies the process of downloading, installing, and managing packages from a GitHub repository. This tool supports operations like installing packages, updating itself, removing packages, and searching for available packages.

## Features

- **Install Packages**: Download and install packages along with their dependencies.
- **Update Tool**: Keep `espresso` up-to-date with the latest source code.
- **Remove Packages**: Uninstall packages and clean up associated files.
- **Search Packages**: List or search for available packages.

## Installation

To install and use `espresso`, you need to have Go installed on your system. Follow these steps:

1. **Clone the Repository**:
    ```bash
    git clone https://github.com/rudyon/espresso.git
    cd espresso
    ```

2. **Build the Project**:
    ```bash
    go build -o espresso
    ```

3. **Move the Binary to a Directory in Your PATH** (e.g., `/usr/local/bin`):
    ```bash
    sudo mv espresso /usr/local/bin/
    ```

## Usage

Run `espresso` commands with root privileges to ensure proper functionality.

### Commands

- **brew <package>**: Download and install a package and its dependencies.
    ```bash
    sudo espresso brew <package>
    ```

- **update**: Update `espresso` to the latest version.
    ```bash
    sudo espresso update
    ```

- **remove <package>**: Remove an installed package.
    ```bash
    sudo espresso remove <package>
    ```

- **look [searchTerm]**: Search for packages. If `searchTerm` is provided, only packages containing the term will be listed.
    ```bash
    sudo espresso look [searchTerm]
    ```

## Configuration

`Espresso` stores downloaded packages in `/etc/espresso`. Ensure this directory has the appropriate permissions and enough space.

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the [LICENSE](LICENSE) file for details.

## Warning

`Espresso` disables TLS certificate verification for HTTP requests, which is insecure. Use it in a secure environment and ensure you understand the risks.

## Troubleshooting

- **Error: `TLS certificate verification is disabled.`**
  This warning indicates that certificate verification is skipped for HTTP requests. Use `espresso` only in trusted environments.

- **Error: `This program must be run as root.`**
  Many operations require root privileges. Ensure you use `sudo` for `espresso` commands.

## Contributing

Contributions are welcome! Please submit issues or pull requests on the [GitHub repository](https://github.com/rudyon/espresso).

## Contact

For questions or feedback, contact the [project maintainer](https://github.com/rudyon).

---

Enjoy using `espresso` on Coffee Linux! ☕
