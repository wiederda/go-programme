package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

// ContainerInfo represents the structure of Docker container information
type ContainerInfo struct {
	Name       string          `json:"Name" yaml:"Name"`
	Config     ContainerConfig `json:"Config" yaml:"Config"`
	HostConfig HostConfig      `json:"HostConfig" yaml:"HostConfig"`
	Mounts     []Mount         `json:"Mounts" yaml:"Mounts"`
}

// ContainerConfig contains configuration information for the container
type ContainerConfig struct {
	Image  string            `json:"Image" yaml:"Image"`
	Labels map[string]string `json:"Labels" yaml:"Labels"`
	Env    []string          `json:"Env" yaml:"Env"`
}

// HostConfig contains host-specific configurations
type HostConfig struct {
	PortBindings map[string][]PortBinding `json:"PortBindings" yaml:"PortBindings"`
}

// PortBinding describes the port binding between host and container
type PortBinding struct {
	HostPort string `json:"HostPort" yaml:"HostPort"`
}

// Mount describes the volumes used in the container
type Mount struct {
	Source      string `json:"Source" yaml:"Source"`
	Destination string `json:"Destination" yaml:"Destination"`
}

// ComposeFile represents the structure for the Docker Compose file
type ComposeFile struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}

// Service describes a single service in the Docker Compose file
type Service struct {
	Image         string            `yaml:"image"`
	ContainerName string            `yaml:"container_name"`
	Ports         []string          `yaml:"ports"`
	Environment   []string          `yaml:"environment"`
	Volumes       []string          `yaml:"volumes"`
	Labels        map[string]string `yaml:"labels,omitempty"`
}

// convertToCompose converts ContainerInfo to ComposeFile format
func convertToCompose(container ContainerInfo) ComposeFile {
	// Container name without the leading slash
	containerName := strings.TrimPrefix(container.Name, "/")

	// Port mappings
	var portMappings []string
	for containerPort, bindings := range container.HostConfig.PortBindings {
		for _, binding := range bindings {
			portMappings = append(portMappings, fmt.Sprintf("%s:%s", binding.HostPort, containerPort))
		}
	}

	// Volume mappings
	var volumeMappings []string
	for _, mount := range container.Mounts {
		volumeMappings = append(volumeMappings, fmt.Sprintf("%s:%s", mount.Source, mount.Destination))
	}

	// Build Docker Compose structure
	return ComposeFile{
		Services: map[string]Service{
			containerName: {
				Image:         container.Config.Image,
				ContainerName: containerName,
				Ports:         portMappings,
				Environment:   container.Config.Env,
				Volumes:       volumeMappings,
				Labels:        container.Config.Labels,
			},
		},
	}
}

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "", "Pfad zur YAML-Datei mit Containerinformationen (wenn kein direkter Zugriff auf den Docker-Host möglich ist)")
	outputFile := flag.String("output", "docker-compose.yml", "Pfad zur Ausgabe-Docker-Compose-Datei (Standard: `docker-compose.yml`)")
	containerName := flag.String("container", "", "Name des Docker-Containers (wenn keine YAML-Datei mit Containerinformationen angegeben ist und direkter Zugriff auf den Docker-Host möglich ist)")

	// Define and set custom help message
	flag.Usage = printHelp

	// Parse arguments
	flag.Parse()

	// Check for input method: file or Docker inspect
	var containerInfos []ContainerInfo

	if *inputFile != "" {
		// Read container information from input YAML file
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Printf("Error reading input file: %v\n", err)
			os.Exit(1)
		}

		// Unmarshal YAML data
		if err := yaml.Unmarshal(data, &containerInfos); err != nil {
			fmt.Printf("Error parsing input YAML data: %v\n", err)
			os.Exit(1)
		}

	} else if *containerName != "" {
		// Ensure container name starts with "/"
		if (*containerName)[0] != '/' {
			*containerName = "/" + *containerName
		}

		// Check if Docker is installed
		if _, err := exec.LookPath("docker"); err != nil {
			fmt.Println("Error: Docker is not installed or not in PATH.")
			os.Exit(1)
		}

		// Execute Docker Inspect command
		jsonData, err := exec.Command("docker", "inspect", *containerName).Output()
		if err != nil {
			fmt.Printf("Error executing Docker command: %v\n", err)
			os.Exit(1)
		}

		// Unmarshal JSON data into ContainerInfo structure
		if err := json.Unmarshal(jsonData, &containerInfos); err != nil {
			fmt.Printf("Error parsing JSON data: %v\n", err)
			os.Exit(1)
		}

	} else {
		fmt.Println("Error: Either -input file or -container name must be provided.")
		flag.Usage()
		os.Exit(1)
	}

	// Process only the first container (for simplicity)
	container := containerInfos[0]

	// Convert to Docker Compose format
	compose := convertToCompose(container)

	// Generate YAML output
	output, err := yaml.Marshal(compose)
	if err != nil {
		fmt.Printf("Error marshaling YAML data: %v\n", err)
		os.Exit(1)
	}

	// Write to output file
	if err := os.WriteFile(*outputFile, output, 0644); err != nil {
		fmt.Printf("Error writing to output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Docker Compose file successfully written to %s\n", *outputFile)
}

func printHelp() {
	executableName := "docker-compose-converter" // Default name for Linux and macOS
	//if runtime.GOOS == "windows" {
	//	executableName += ".exe" // Append .exe for Windows
	//}

	fmt.Printf("Usage: %s [options]\n", executableName)
	fmt.Println("Options:")
	fmt.Println("  -input string")
	fmt.Println("        Pfad zur YAML-Datei mit Containerinformationen (wenn kein direkter Zugriff auf den Docker-Host möglich ist).")
	fmt.Println("  -container string")
	fmt.Println("        Name des Docker-Containers oder ID (wenn keine YAML-Datei mit Containerinformationen angegeben ist und direkter Zugriff auf den Docker-Host möglich ist).")
	fmt.Println("  -output string")
	fmt.Println("        Pfad zur Ausgabe-Docker-Compose-Datei (Standard: `docker-compose.yml`).")
	fmt.Println("  -help")
	fmt.Println("        Zeigt diese Hilfenachricht an.")
}
