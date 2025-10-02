package utils

import (
	"os"
	"testing"
)

func TestDockerDetector_IsRunningInDocker(t *testing.T) {
	detector := NewDockerDetector()

	// Test detection with .dockerenv file
	t.Run("with .dockerenv file", func(t *testing.T) {
		// In real Docker container, /.dockerenv file would exist
		// We can only test the method doesn't panic
		_ = detector.IsRunningInDocker()
	})

	// Test environment variable detection
	t.Run("with container environment variables", func(t *testing.T) {
		originalEnv := os.Getenv("KUBERNETES_SERVICE_HOST")
		defer os.Setenv("KUBERNETES_SERVICE_HOST", originalEnv)

		os.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
		detector := NewDockerDetector()
		// This might return true if we're actually in a container
		_ = detector.checkContainerEnvironment()

		os.Unsetenv("KUBERNETES_SERVICE_HOST")
		// Should return false when no container env vars are set
		_ = detector.checkContainerEnvironment()
	})
}

func TestDockerVolumeResolver_isDockerAvailable(t *testing.T) {
	resolver := NewDockerVolumeResolver()
	// Just check if the method runs without panic
	// Actual result depends on whether Docker is installed
	_ = resolver.isDockerAvailable()
}

func TestDockerVolumeResolver_VolumeExists(t *testing.T) {
	resolver := NewDockerVolumeResolver()

	// Test with empty volume name
	exists := resolver.VolumeExists("")
	if exists {
		t.Error("Expected false for empty volume name")
	}

	// Test with non-existent volume (assuming this volume doesn't exist)
	exists = resolver.VolumeExists("test-volume-that-does-not-exist-12345")
	if exists {
		t.Error("Expected false for non-existent volume")
	}
}

func TestDockerVolumeResolver_GetVolumeRealPath(t *testing.T) {
	resolver := NewDockerVolumeResolver()

	// Test with empty volume name
	_, err := resolver.GetVolumeRealPath("")
	if err == nil {
		t.Error("Expected error for empty volume name")
	}

	// Test with non-existent volume
	_, err = resolver.GetVolumeRealPath("test-volume-that-does-not-exist-12345")
	// Should return error (either Docker not available or volume not found)
	if err == nil && resolver.isDockerAvailable() {
		t.Error("Expected error for non-existent volume")
	}
}

func TestDockerVolumeResolver_GetVolumeInfo(t *testing.T) {
	resolver := NewDockerVolumeResolver()

	// Test with empty volume name
	_, err := resolver.GetVolumeInfo("")
	if err == nil {
		t.Error("Expected error for empty volume name")
	}

	// Test with non-existent volume
	_, err = resolver.GetVolumeInfo("test-volume-that-does-not-exist-12345")
	// Should return error (either Docker not available or volume not found)
	if err == nil && resolver.isDockerAvailable() {
		t.Error("Expected error for non-existent volume")
	}
}

func TestDockerVolumeResolver_CreateAndRemoveVolume(t *testing.T) {
	resolver := NewDockerVolumeResolver()

	// Skip test if Docker is not available
	if !resolver.isDockerAvailable() {
		t.Skip("Docker is not available, skipping volume create/remove tests")
	}

	volumeName := "test-xsha-config-test-volume"

	// Clean up any existing volume first
	_ = resolver.RemoveVolume(volumeName, true)

	// Create volume with labels
	labels := map[string]string{
		"test":    "true",
		"purpose": "unit-test",
	}

	err := resolver.CreateVolume(volumeName, labels)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}

	// Check if volume exists
	exists := resolver.VolumeExists(volumeName)
	if !exists {
		t.Error("Volume should exist after creation")
	}

	// Get volume info
	info, err := resolver.GetVolumeInfo(volumeName)
	if err != nil {
		t.Errorf("Failed to get volume info: %v", err)
	} else {
		if info.Name != volumeName {
			t.Errorf("Expected volume name %s, got %s", volumeName, info.Name)
		}
		if info.Mountpoint == "" {
			t.Error("Expected mountpoint to be non-empty")
		}
	}

	// Remove volume
	err = resolver.RemoveVolume(volumeName, false)
	if err != nil {
		t.Errorf("Failed to remove volume: %v", err)
	}

	// Check if volume is removed
	exists = resolver.VolumeExists(volumeName)
	if exists {
		t.Error("Volume should not exist after removal")
	}
}
