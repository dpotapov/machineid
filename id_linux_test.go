//go:build linux
// +build linux

package machineid

import (
	"os"
	"reflect"
	"testing"
)

func TestMachineID(t *testing.T) {
	id, err := machineID()
	if err != nil {
		t.Errorf("machineID() err = %v", err)
	}
	if len(id) == 0 {
		t.Errorf("Got empty ID")
	}
}

func TestLookupMachineID(t *testing.T) {
	emptyTempFile := makeTempFile(t, 0600)
	defer os.Remove(emptyTempFile)

	// Test 1: when readFirstFile has bad files in the search list
	paths := []string{"/nonexistent/directory", emptyTempFile, "/nonexistent/directory"}

	_, err := lookupMachineID(paths)
	if err != nil {
		t.Errorf("lookupMachineID() err = %v", err)
	}

	// Test 2: when readFirstFile doesn't return error even if one of the files error'ed
	paths = []string{"/nonexistent/directory", emptyTempFile}
	_, err = lookupMachineID(paths)
	if err != nil {
		t.Errorf("lookupMachineID() err = %v", err)
	}
}

func TestGenerateID(t *testing.T) {
	tempFile := makeTempFile(t, 0600)
	defer os.Remove(tempFile)

	// Test 1: Test when readFile and writeFirstFile succeed
	paths := []string{tempFile}

	id, err := generateID(paths)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(id) == 0 {
		t.Errorf("Got empty ID")
	}
	b, err := readFile(tempFile)
	if err != nil {
		t.Errorf("Unable to read temp file: %v", err)
	}
	if id != trim(string(b)) {
		t.Errorf("Generated ID was not written correctly into a file, want %v, got %v", id, string(b))
	}

	// Test 2: Test when it is unable to write to a file
	tempFile = makeTempFile(t, 0400)
	paths = []string{tempFile}
	_, err = generateID(paths)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Test 3: Test when operating on non-existing file
	paths = []string{"/nonexistent/directory"}

	_, err = generateID(paths)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSearchPaths(t *testing.T) {
	// Save original environment variables
	originalEnvPathname := os.Getenv(ENV_VARNAME)
	originalHome := os.Getenv("HOME")

	defer func() {
		// Restore original environment variables after test
		os.Setenv(ENV_VARNAME, originalEnvPathname)
		os.Setenv("HOME", originalHome)
	}()

	// Test 1: ENV_VARNAME and HOME are not empty
	os.Setenv(ENV_VARNAME, "/test/path")
	os.Setenv("HOME", "/home/test")

	expected := []string{"/test/path", dbusPath, dbusPathEtc, "/home/test/.config/machine-id"}

	result := searchPaths()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test 2: ENV_VARNAME and HOME are empty
	os.Setenv(ENV_VARNAME, "")
	os.Setenv("HOME", "")

	expected = []string{dbusPath, dbusPathEtc}

	result = searchPaths()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func makeTempFile(t *testing.T, mode os.FileMode) string {
	tempFile, err := os.CreateTemp("", "machineid_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_ = tempFile.Close()
	if err := os.Chmod(tempFile.Name(), mode); err != nil {
		t.Fatalf("Unable to set file mode: %v", err)
	}
	return tempFile.Name()
}
