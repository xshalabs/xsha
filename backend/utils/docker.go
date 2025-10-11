package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DockerDetector provides Docker environment detection
type DockerDetector struct{}

// NewDockerDetector creates a new Docker detector
func NewDockerDetector() *DockerDetector {
	return &DockerDetector{}
}

// IsRunningInDocker detects if the application is running inside a Docker container
func (d *DockerDetector) IsRunningInDocker() bool {
	// Method 1: Check for /.dockerenv file (most reliable)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Method 2: Check /proc/1/cgroup for docker/containerd signatures
	if d.checkCgroupForDocker() {
		return true
	}

	// Method 3: Check environment variables commonly set in containers
	if d.checkContainerEnvironment() {
		return true
	}

	return false
}

// checkCgroupForDocker checks /proc/1/cgroup for Docker signatures
func (d *DockerDetector) checkCgroupForDocker() bool {
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}

	content := string(data)
	containerSignatures := []string{
		"docker",
		"containerd",
		"/docker/",
		"/lxc/",
		"/kubepods/",
	}

	for _, signature := range containerSignatures {
		if strings.Contains(content, signature) {
			return true
		}
	}

	return false
}

// checkContainerEnvironment checks for container-specific environment variables
func (d *DockerDetector) checkContainerEnvironment() bool {
	// Check for common container environment variables
	containerEnvVars := []string{
		"KUBERNETES_SERVICE_HOST",
		"DOCKER_CONTAINER",
		"container",
	}

	for _, envVar := range containerEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	// Check if we're in the expected container environment
	if workspaceEnv := os.Getenv("XSHA_WORKSPACE_BASE_DIR"); workspaceEnv == "/app/workspaces" {
		return true
	}

	return false
}

// DockerVolumeResolver provides Docker volume resolution functionality
type DockerVolumeResolver struct{}

// NewDockerVolumeResolver creates a new Docker volume resolver
func NewDockerVolumeResolver() *DockerVolumeResolver {
	return &DockerVolumeResolver{}
}

// VolumeInfo represents Docker volume information
type VolumeInfo struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	Labels     map[string]string `json:"Labels"`
	Scope      string            `json:"Scope"`
	Options    map[string]string `json:"Options"`
}

// GetVolumeRealPath inspects a Docker volume and returns its real mount path
func (r *DockerVolumeResolver) GetVolumeRealPath(volumeName string) (string, error) {
	if volumeName == "" {
		return "", fmt.Errorf("volume name cannot be empty")
	}

	// Check if Docker is available
	if !r.isDockerAvailable() {
		return "", fmt.Errorf("Docker is not available")
	}

	// Execute docker volume inspect command
	cmd := exec.Command("docker", "volume", "inspect", volumeName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to inspect Docker volume %s: %w (stderr: %s)",
			volumeName, err, stderr.String())
	}

	// Parse JSON output
	var volumes []VolumeInfo
	if err := json.Unmarshal(stdout.Bytes(), &volumes); err != nil {
		return "", fmt.Errorf("failed to parse Docker volume inspect output: %w", err)
	}

	// Validate response
	if len(volumes) == 0 {
		return "", fmt.Errorf("volume %s not found", volumeName)
	}

	volume := volumes[0]
	if volume.Mountpoint == "" {
		return "", fmt.Errorf("mountpoint not found for volume %s", volumeName)
	}

	return volume.Mountpoint, nil
}

// GetVolumeInfo returns complete information about a Docker volume
func (r *DockerVolumeResolver) GetVolumeInfo(volumeName string) (*VolumeInfo, error) {
	if volumeName == "" {
		return nil, fmt.Errorf("volume name cannot be empty")
	}

	// Check if Docker is available
	if !r.isDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available")
	}

	// Execute docker volume inspect command
	cmd := exec.Command("docker", "volume", "inspect", volumeName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to inspect Docker volume %s: %w (stderr: %s)",
			volumeName, err, stderr.String())
	}

	// Parse JSON output
	var volumes []VolumeInfo
	if err := json.Unmarshal(stdout.Bytes(), &volumes); err != nil {
		return nil, fmt.Errorf("failed to parse Docker volume inspect output: %w", err)
	}

	// Validate response
	if len(volumes) == 0 {
		return nil, fmt.Errorf("volume %s not found", volumeName)
	}

	return &volumes[0], nil
}

// VolumeExists checks if a Docker volume exists
func (r *DockerVolumeResolver) VolumeExists(volumeName string) bool {
	if volumeName == "" || !r.isDockerAvailable() {
		return false
	}

	cmd := exec.Command("docker", "volume", "ls", "-q")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	volumes := strings.Split(string(output), "\n")
	for _, v := range volumes {
		if strings.TrimSpace(v) == volumeName {
			return true
		}
	}

	return false
}

// isDockerAvailable checks if Docker command is available
func (r *DockerVolumeResolver) isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

// CreateVolume creates a new Docker volume with optional labels
func (r *DockerVolumeResolver) CreateVolume(volumeName string, labels map[string]string) error {
	if volumeName == "" {
		return fmt.Errorf("volume name cannot be empty")
	}

	if !r.isDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	args := []string{"volume", "create", volumeName}

	// Add labels if provided
	for key, value := range labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	cmd := exec.Command("docker", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create Docker volume %s: %w (stderr: %s)",
			volumeName, err, stderr.String())
	}

	return nil
}

// RemoveVolume removes a Docker volume
func (r *DockerVolumeResolver) RemoveVolume(volumeName string, force bool) error {
	if volumeName == "" {
		return fmt.Errorf("volume name cannot be empty")
	}

	if !r.isDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	args := []string{"volume", "rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, volumeName)

	cmd := exec.Command("docker", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove Docker volume %s: %w (stderr: %s)",
			volumeName, err, stderr.String())
	}

	return nil
}
