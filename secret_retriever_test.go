package main

import (
	"context"
	"github.com/fr0stylo/secretary/runtime"
	"os"
	"testing"
	"time"
)

// MockSecretRetrieverClient is a mock implementation of SecretRetrieverClient for testing
type MockSecretRetrieverClient struct {
	secretValues   map[string][]byte
	secretVersions map[string]string
}

// NewMockSecretRetrieverClient creates a new mock client with predefined values
func NewMockSecretRetrieverClient() *MockSecretRetrieverClient {
	return &MockSecretRetrieverClient{
		secretValues:   make(map[string][]byte),
		secretVersions: make(map[string]string),
	}
}

// SetSecretValue sets a secret value for testing
func (m *MockSecretRetrieverClient) SetSecretValue(id string, value []byte) {
	m.secretValues[id] = value
}

// SetSecretVersion sets a secret version for testing
func (m *MockSecretRetrieverClient) SetSecretVersion(id string, version string) {
	m.secretVersions[id] = version
}

// GetSecretValue implements SecretRetrieverClient.GetSecretValue
func (m *MockSecretRetrieverClient) GetSecretValue(ctx context.Context, id string) ([]byte, error) {
	if value, ok := m.secretValues[id]; ok {
		return value, nil
	}
	return nil, nil
}

// GetSecretVersion implements SecretRetrieverClient.GetSecretVersion
func (m *MockSecretRetrieverClient) GetSecretVersion(ctx context.Context, id string) (string, error) {
	if version, ok := m.secretVersions[id]; ok {
		return version, nil
	}
	return "default-version", nil
}

func TestDefaultOpts(t *testing.T) {
	config := runtime.DefaultOptions()

	if config.Frequency != 15*time.Second {
		t.Errorf("Expected default frequency to be 15s, got %v", config.Frequency)
	}

	if config.Timeout != 10*time.Second {
		t.Errorf("Expected default timeout to be 10s, got %v", config.Timeout)
	}
}

func TestWithFrequency(t *testing.T) {
	config := runtime.DefaultOptions()
	opt := runtime.WithFrequency(30 * time.Second)
	opt(config)

	if config.Frequency != 30*time.Second {
		t.Errorf("Expected frequency to be 30s, got %v", config.Frequency)
	}
}

func TestWithTimeout(t *testing.T) {
	config := runtime.DefaultOptions()
	opt := runtime.WithTimeout(20 * time.Second)
	opt(config)

	if config.Timeout != 20*time.Second {
		t.Errorf("Expected timeout to be 20s, got %v", config.Timeout)
	}
}

func TestNewSecretRetriever(t *testing.T) {
	client := NewMockSecretRetrieverClient()
	retriever := runtime.NewRuntime(client, runtime.WithFrequency(30*time.Second))

	if retriever.Client != client {
		t.Error("Expected client to be set correctly")
	}

	if retriever.Config.Frequency != 30*time.Second {
		t.Errorf("Expected frequency to be 30s, got %v", retriever.Config.Frequency)
	}

	if len(retriever.PulledVersions) != 0 {
		t.Errorf("Expected PulledVersions to be empty, got %v", retriever.PulledVersions)
	}
}

func TestCreateSecret(t *testing.T) {
	// Setup
	client := NewMockSecretRetrieverClient()
	client.SetSecretValue("test-secret", []byte("secret-value"))
	client.SetSecretVersion("test-secret", "v1")

	retriever := runtime.NewRuntime(client)

	// Create a temporary file path
	tempPath := "/tmp/test-secret"
	defer os.Remove(tempPath)

	// Test
	secret := &runtime.Secret{
		Identifier: "test-secret",
		Path:       tempPath,
	}

	err := retriever.CreateSecret(context.Background(), secret)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the secret was added to PulledVersions
	if len(retriever.PulledVersions) != 1 {
		t.Errorf("Expected 1 secret in PulledVersions, got %d", len(retriever.PulledVersions))
	}

	if retriever.PulledVersions[0].Version != "v1" {
		t.Errorf("Expected version to be v1, got %s", retriever.PulledVersions[0].Version)
	}

	// Verify the file was created
	content, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(content) != "secret-value" {
		t.Errorf("Expected file content to be 'secret-value', got '%s'", string(content))
	}

	// Verify environment variable was set
	envValue := os.Getenv("test-secret")
	if envValue != tempPath {
		t.Errorf("Expected environment variable to be set to %s, got %s", tempPath, envValue)
	}
}

func TestCreateSecretsFromEnvironment(t *testing.T) {
	// Let's simplify this test and focus on just testing the CreateSecretsFromEnvironment method
	client := NewMockSecretRetrieverClient()
	client.SetSecretValue("aws/secret1", []byte("secret-value-1"))
	client.SetSecretVersion("aws/secret1", "v1")

	retriever := runtime.NewRuntime(client)

	// Clean up any existing files from previous test runs
	os.Remove("/tmp/SECRET1")

	// Create a test environment variable directly
	t.Logf("Setting environment variable SECRETARY_SECRET1=aws/secret1")
	os.Setenv("SECRETARY_SECRET1", "aws/secret1")
	defer os.Unsetenv("SECRETARY_SECRET1")

	// Create a simplified test environment
	testEnv := []string{"SECRETARY_SECRET1=aws/secret1"}

	// Call the method directly with our simplified environment
	t.Logf("Calling CreateSecretsFromEnvironment with: %v", testEnv)
	err := retriever.CreateSecretsFromEnvironment(context.Background(), testEnv)
	if err != nil {
		t.Fatalf("CreateSecretsFromEnvironment failed: %v", err)
	}

	// Verify the secret was added to PulledVersions
	t.Logf("PulledVersions: %+v", retriever.PulledVersions)
	if len(retriever.PulledVersions) != 1 {
		t.Errorf("Expected 1 secret in PulledVersions, got %d", len(retriever.PulledVersions))
	} else {
		t.Logf("Secret added to PulledVersions: %+v", retriever.PulledVersions[0])
	}

	// List all files in /tmp to debug
	files, _ := os.ReadDir("/tmp")
	t.Logf("Files in /tmp:")
	for _, file := range files {
		t.Logf("  %s", file.Name())
	}

	// Verify the file was created
	t.Logf("Checking for file at /tmp/SECRET1")
	if _, err := os.Stat("/tmp/SECRET1"); os.IsNotExist(err) {
		t.Fatalf("File /tmp/SECRET1 does not exist")
	}

	content, err := os.ReadFile("/tmp/SECRET1")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(content) != "secret-value-1" {
		t.Errorf("Expected file content to be 'secret-value-1', got '%s'", string(content))
	}
	defer os.Remove("/tmp/SECRET1")
}

func TestRunAndStop(t *testing.T) {
	// Setup
	client := NewMockSecretRetrieverClient()
	client.SetSecretValue("test-secret", []byte("secret-value"))
	client.SetSecretVersion("test-secret", "v1")

	retriever := runtime.NewRuntime(client, runtime.WithFrequency(100*time.Millisecond))

	// Create a secret
	tempPath := "/tmp/run-test-secret"
	defer os.Remove(tempPath)

	secret := runtime.Secret{
		Identifier: "test-secret",
		Path:       tempPath,
		Version:    "v1",
	}

	// Add the secret to PulledVersions
	retriever.PulledVersions = append(retriever.PulledVersions, &secret)

	// Run the retriever
	ctx := context.Background()
	changeCh := retriever.WatchChanges(ctx)

	// Verify the retriever is running
	if retriever.RunCancel == nil {
		t.Error("Expected runCancel to be set")
	}

	// Stop the retriever
	retriever.StopWatchChanges()

	// Verify no panics when stopping again
	retriever.StopWatchChanges()

	// Verify the channel is closed or at least not receiving updates
	select {
	case _, ok := <-changeCh:
		if ok {
			t.Error("Expected channel to be closed or not receiving updates")
		}
	case <-time.After(200 * time.Millisecond):
		// This is expected, as the channel might not be closed but should not receive updates
	}
}

func TestSecretVersionChange(t *testing.T) {
	// Setup
	client := NewMockSecretRetrieverClient()
	client.SetSecretValue("test-secret", []byte("secret-value"))
	client.SetSecretVersion("test-secret", "v1")

	retriever := runtime.NewRuntime(client, runtime.WithFrequency(100*time.Millisecond))

	// Create a secret
	tempPath := "/tmp/version-test-secret"
	defer os.Remove(tempPath)

	secret := runtime.Secret{
		Identifier: "test-secret",
		Path:       tempPath,
		Version:    "v1",
	}

	// Add the secret to PulledVersions
	retriever.PulledVersions = append(retriever.PulledVersions, &secret)

	// Run the retriever
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changeCh := retriever.WatchChanges(ctx)

	// Wait a bit to ensure the ticker has run at least once
	time.Sleep(150 * time.Millisecond)

	// Change the secret version
	client.SetSecretVersion("test-secret", "v2")

	// Wait for the change notification
	select {
	case <-changeCh:
		// Success - we received a change notification
	case <-time.After(300 * time.Millisecond):
		t.Error("Expected to receive a change notification")
	}

	// Stop the retriever
	retriever.StopWatchChanges()
}
