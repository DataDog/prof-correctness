package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty JSON",
			content:     `{}`,
			wantErr:     true,
			errContains: "stacks is required",
		},
		{
			name:        "wrong structure",
			content:     `{"foo": "bar"}`,
			wantErr:     true,
			errContains: "stacks is required",
		},
		{
			name:        "empty stacks without note",
			content:     `{"stacks": []}`,
			wantErr:     true,
			errContains: "stacks",
		},
		{
			name:        "empty stacks with note",
			content:     `{"stacks": [], "note": "This test doesn't check profile content"}`,
			wantErr:     false,
		},
		{
			name:        "missing profile-type",
			content:     `{"stacks": [{"stack-content": [{"regular_expression": "test", "value": 100}]}]}`,
			wantErr:     true,
			errContains: "profile-type",
		},
		{
			name:        "missing stack-content",
			content:     `{"stacks": [{"profile-type": "wall-time"}]}`,
			wantErr:     true,
			errContains: "stack-content",
		},
		{
			name:        "empty stack-content",
			content:     `{"stacks": [{"profile-type": "wall-time", "stack-content": []}]}`,
			wantErr:     true,
			errContains: "stack-content",
		},
		{
			name:        "missing regular_expression",
			content:     `{"stacks": [{"profile-type": "wall-time", "stack-content": [{"value": 100}]}]}`,
			wantErr:     true,
			errContains: "regular_expression",
		},
		{
			name:        "missing value and percent",
			content:     `{"stacks": [{"profile-type": "wall-time", "stack-content": [{"regular_expression": "test"}]}]}`,
			wantErr:     true,
			errContains: "value",
		},
		{
			name: "no value/percent but has value-matching-sum",
			content: `{
				"stacks": [{
					"profile-type": "cpu-time",
					"stack-content": [{"regular_expression": ".*test.*"}],
					"value-matching-sum": 1000000000
				}]
			}`,
			wantErr: false,
		},
		{
			name: "valid JSON with value",
			content: `{
				"test_name": "test",
				"stacks": [{
					"profile-type": "wall-time",
					"stack-content": [{"regular_expression": "^test$", "value": 100}]
				}]
			}`,
			wantErr: false,
		},
		{
			name: "valid JSON with percent",
			content: `{
				"stacks": [{
					"profile-type": "wall-time",
					"stack-content": [{"regular_expression": "^test$", "percent": 50}]
				}]
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "schema_test_*.json")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			_, err = readJSONFile(tmpFile.Name())

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
			}
		})
	}
}

func TestSchemaValidation_AllExistingProfiles(t *testing.T) {
	scenariosDir := "scenarios"
	if _, err := os.Stat(scenariosDir); os.IsNotExist(err) {
		t.Skip("scenarios directory not found")
	}

	var jsonFiles []string
	err := filepath.Walk(scenariosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "expected_profile.json" {
			jsonFiles = append(jsonFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk scenarios directory: %v", err)
	}

	if len(jsonFiles) == 0 {
		t.Skip("No expected_profile.json files found")
	}

	t.Logf("Found %d expected_profile.json files to validate", len(jsonFiles))

	for _, jsonFile := range jsonFiles {
		t.Run(jsonFile, func(t *testing.T) {
			if _, err := readJSONFile(jsonFile); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}
