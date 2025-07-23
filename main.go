package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nexidian/gocliselect"
)

type DockerPublishConfig struct {
	DockerImageName string `json:"dockerImageName"`
	Version         string `json:"version"`
}

type VersionUpdateValues struct {
	Current string
	Patch   string
	Minor   string
	Major   string
}

func main() {
	workingDirectory := getWorkingDirectory()
	dockerPublishFilePath := filepath.Join(workingDirectory, ".docker-publish")

	// Initialize .docker-publish file if it doesn't exist
	if !fileExists(dockerPublishFilePath) {
		fmt.Println("Initializing .docker-publish file...")

		fmt.Print("Enter the Docker image name (e.g., ghcr.io/energypatrikhu/bandwidth-hero-proxy-go): ")
		dockerImageName := readInput()
		if dockerImageName == "" {
			fmt.Fprintf(os.Stderr, "Docker image name is required.\n")
			os.Exit(1)
		}

		fmt.Print("Enter the initial version (e.g., 1.0.0): ")
		initialVersion := readInput()
		if initialVersion == "" {
			fmt.Fprintf(os.Stderr, "Initial version is required.\n")
			os.Exit(1)
		}

		config := DockerPublishConfig{
			DockerImageName: strings.TrimSpace(dockerImageName),
			Version:         strings.TrimSpace(initialVersion),
		}

		if err := writeConfigFile(dockerPublishFilePath, config); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(".docker-publish file created successfully.")
	}

	// Read existing config
	config, err := readConfigFile(dockerPublishFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		os.Exit(1)
	}

	dockerImageName := config.DockerImageName
	version := config.Version

	// Prepare version update options
	versionUpdateValues := VersionUpdateValues{
		Current: version,
		Patch:   updateVersion(version, "Patch"),
		Minor:   updateVersion(version, "Minor"),
		Major:   updateVersion(version, "Major"),
	}

	// Select version update type
	fmt.Println("Select the version update type:")

	menu := gocliselect.NewMenu("Choose version update")
	menu.AddItem(fmt.Sprintf("Current (%s)", versionUpdateValues.Current), "Current")
	menu.AddItem(fmt.Sprintf("Patch (%s)", versionUpdateValues.Patch), "Patch")
	menu.AddItem(fmt.Sprintf("Minor (%s)", versionUpdateValues.Minor), "Minor")
	menu.AddItem(fmt.Sprintf("Major (%s)", versionUpdateValues.Major), "Major")

	versionType := menu.Display()
	if versionType == "" {
		fmt.Fprintf(os.Stderr, "No selection made\n")
		os.Exit(1)
	}

	var selectedVersion string

	switch versionType {
	case "Current":
		selectedVersion = versionUpdateValues.Current
	case "Patch":
		selectedVersion = versionUpdateValues.Patch
	case "Minor":
		selectedVersion = versionUpdateValues.Minor
	case "Major":
		selectedVersion = versionUpdateValues.Major
	default:
		fmt.Fprintf(os.Stderr, "Invalid choice\n")
		os.Exit(1)
	}

	fmt.Printf("You selected: %s (%s), current version: %s\n", versionType, selectedVersion, version)

	buildMenu := gocliselect.NewMenu("Do you want to build, publish the image and update the version in the .docker-publish file?")
	buildMenu.AddItem("Yes", "yes")
	buildMenu.AddItem("No", "no")
	buildAndPublishConfirm := buildMenu.Display()

	if buildAndPublishConfirm == "" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}

	pushChangesToGit := false
	if fileExists(filepath.Join(workingDirectory, ".git")) && selectedVersion != version {
		gitMenu := gocliselect.NewMenu("Do you want to push the changes to GIT?")
		gitMenu.AddItem("Yes", "yes")
		gitMenu.AddItem("No", "no")
		pushChangesConfirm := gitMenu.Display()
		pushChangesToGit = pushChangesConfirm == "yes"
	}

	if buildAndPublishConfirm == "yes" {
		// Update the version in config
		config.Version = selectedVersion
		if err := writeConfigFile(dockerPublishFilePath, config); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Version updated to: %s\n", selectedVersion)

		// Build the Docker image
		fmt.Println("Building Docker image...")
		if err := runCommand("docker", "build", "-t", fmt.Sprintf("%s:%s", dockerImageName, selectedVersion), "-t", fmt.Sprintf("%s:latest", dockerImageName), workingDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "Error building Docker image: %v\n", err)
			os.Exit(1)
		}

		// Publish the image
		fmt.Println("Publishing Docker image...")
		if err := runCommand("docker", "push", fmt.Sprintf("%s:%s", dockerImageName, selectedVersion)); err != nil {
			fmt.Fprintf(os.Stderr, "Error pushing versioned image: %v\n", err)
			os.Exit(1)
		}

		if err := runCommand("docker", "push", fmt.Sprintf("%s:latest", dockerImageName)); err != nil {
			fmt.Fprintf(os.Stderr, "Error pushing latest image: %v\n", err)
			os.Exit(1)
		}

		if pushChangesToGit {
			// Push changes to GIT
			fmt.Println("Pushing changes to GIT...")
			if err := runCommandInDir(workingDirectory, "git", "add", ".docker-publish"); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding files to git: %v\n", err)
				os.Exit(1)
			}

			if err := runCommandInDir(workingDirectory, "git", "commit", "-m", fmt.Sprintf("chore: update version to %s", selectedVersion)); err != nil {
				fmt.Fprintf(os.Stderr, "Error committing changes: %v\n", err)
				os.Exit(1)
			}

			if err := runCommandInDir(workingDirectory, "git", "push"); err != nil {
				fmt.Fprintf(os.Stderr, "Error pushing to git: %v\n", err)
				os.Exit(1)
			}

			if err := runCommandInDir(workingDirectory, "git", "checkout", "main"); err != nil {
				fmt.Fprintf(os.Stderr, "Error switching to main branch: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Changes pushed to GIT.")
		}

		fmt.Println("Docker image published successfully!")
	} else {
		fmt.Println("Aborted.")
	}
}

// Utility functions

func getWorkingDirectory() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	return cwd
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(input)
}

func readConfigFile(filepath string) (DockerPublishConfig, error) {
	var config DockerPublishConfig
	data, err := os.ReadFile(filepath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func writeConfigFile(filepath string, config DockerPublishConfig) error {
	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

func updateVersion(version string, versionType string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "Invalid version format: %s\n", version)
		os.Exit(1)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch versionType {
	case "Major":
		major++
		minor = 0
		patch = 0
	case "Minor":
		minor++
		patch = 0
	case "Patch":
		patch++
	default:
		fmt.Fprintf(os.Stderr, "Invalid version type: %s\n", versionType)
		os.Exit(1)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandInDir(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
