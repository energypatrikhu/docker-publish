# Docker Publish (Go Version)

A command-line tool for managing Docker image versioning and publishing, converted from TypeScript to Go.

## Features

- Interactive version management (Major, Minor, Patch) with semantic versioning (stable versions only)
- Automatic Docker image building and publishing
- Git integration for version tracking
- Configuration file management (`.docker-publish`)
- Semantic version validation using [semver.org](https://semver.org/) standards (no pre-release versions)

## Installation

### Option 1: Install with npm from GitHub

```bash
# Install directly from GitHub
npm install -g git+https://github.com/energypatrikhu/docker-publish.git

# Or if published to npm
npm install -g energypatrikhu/docker-publish
```

### Option 2: Build from source

```bash
# Clone the repository
git clone https://github.com/energypatrikhu/docker-publish.git
cd docker-publish

# Build the application
make build
```

## Usage

Run the tool in a directory containing a Dockerfile:

```bash
docker-publish [working-directory]
```

If no directory is specified, it will use the current working directory.

### First Run

On the first run, the tool will prompt you to:
1. Enter the Docker image name (e.g., `username/my-app`)
2. Enter the initial version (e.g., `1.0.0`) - must be a valid stable semantic version

The version must follow [semantic versioning](https://semver.org/) format (e.g., 1.0.0, 2.1.3, 0.1.0). Pre-release versions (alpha, beta, rc) are not supported.

This creates a `.docker-publish` configuration file.

### Subsequent Runs

The tool will:
1. Load the existing configuration
2. Show version update options (Current, Patch, Minor, Major)
3. Ask if you want to build and publish
4. Optionally push changes to Git (if in a Git repository)

## Configuration File

The `.docker-publish` file stores:

```json
{
	"dockerImageName": "username/my-app",
	"version": "1.0.0"
}
```

## Development

### Building

```bash
# Build for current platform
make build

# Clean build artifacts
make clean
```

### Running

```bash
# Run directly with Go
go run .

# Run with arguments
go run . /path/to/project

# Or run the built binary
./bin/docker-publish

# If installed via npm
docker-publish
```

## Requirements

- Go 1.21 or later
- Docker (for building and publishing images)
- Git (optional, for version tracking)

## Differences from TypeScript Version

This Go version provides the same functionality as the original TypeScript/Bun version with the following changes:

- Interactive CLI selection using `gocliselect` library
- Semantic version validation using the Masterminds semver library
- Native Go JSON handling
- Cross-platform binary compilation support
- Better error handling and validation

## License

Same as the original project.
