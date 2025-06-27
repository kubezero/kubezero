package cmd

import (
	"testing"
)

func TestMatchesPodName(t *testing.T) {
	testCases := []struct {
		podName        string
		expectedPrefix string
		expected       bool
	}{
		{"argo-cd-server-123456789-abcde", "argo-cd-server", true},
		{"argo-cd-application-controller-123456789-abcde", "argo-cd-application-controller", true},
		{"argo-cd-repo-server-123456789-abcde", "argo-cd-repo-server", true},
		{"nginx-123456789-abcde", "argo-cd-server", false},
		{"argo-cd", "argo-cd-server", false},
	}

	for _, tc := range testCases {
		t.Run(tc.podName, func(t *testing.T) {
			result := matchesPodName(tc.podName, tc.expectedPrefix)
			if result != tc.expected {
				t.Errorf("matchesPodName(%q, %q) = %v, expected %v",
					tc.podName, tc.expectedPrefix, result, tc.expected)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Test with a file that should exist (this test file)
	if !fileExists("bootstrap_test.go") {
		t.Error("Expected bootstrap_test.go to exist")
	}

	// Test with a file that shouldn't exist
	if fileExists("nonexistent_file.txt") {
		t.Error("Expected nonexistent_file.txt to not exist")
	}
}
