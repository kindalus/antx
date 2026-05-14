package antbox

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNodeMarshalOmitEmpty(t *testing.T) {
	// Test with minimal Node (only required fields)
	node := Node{
		Title:    "test-folder",
		Parent:   "parent-uuid",
		Mimetype: "application/vnd.antbox.folder",
	}

	jsonData, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal node: %v", err)
	}

	jsonStr := string(jsonData)

	// Check that only non-empty fields are present
	expectedFields := []string{
		`"title":"test-folder"`,
		`"parent":"parent-uuid"`,
		`"mimetype":"application/vnd.antbox.folder"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected field %s not found in JSON: %s", field, jsonStr)
		}
	}

	// Check that empty fields are omitted
	omittedFields := []string{
		`"uuid"`,
		`"fid"`,
		`"owner"`,
		`"group"`,
		`"permissions"`,
		`"size"`,
		`"createdTime"`,
		`"modifiedTime"`,
	}

	for _, field := range omittedFields {
		if strings.Contains(jsonStr, field) {
			t.Errorf("Empty field %s should be omitted but found in JSON: %s", field, jsonStr)
		}
	}
}

func TestNodeMarshalWithAllFields(t *testing.T) {
	// Test with all fields populated
	permissions := &Permissions{
		Group:         []Permission{PermissionRead},
		Authenticated: []Permission{PermissionRead, PermissionWrite},
		Anonymous:     []Permission{PermissionRead},
		Advanced:      map[string][]Permission{"custom": {PermissionExport}},
	}

	node := Node{
		UUID:        "test-uuid",
		Fid:         "test-fid",
		Title:       "test-folder",
		Mimetype:    "application/vnd.antbox.folder",
		Parent:      "parent-uuid",
		Owner:       "test-owner",
		Group:       "test-group",
		Permissions: permissions,
		Size:        1024,
		CreatedAt:   "2024-01-01T00:00:00Z",
		ModifiedAt:  "2024-01-01T00:00:00Z",
	}

	jsonData, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal node: %v", err)
	}

	jsonStr := string(jsonData)

	// Check that all fields are present when populated
	expectedFields := []string{
		`"uuid":"test-uuid"`,
		`"fid":"test-fid"`,
		`"title":"test-folder"`,
		`"mimetype":"application/vnd.antbox.folder"`,
		`"parent":"parent-uuid"`,
		`"owner":"test-owner"`,
		`"group":"test-group"`,
		`"permissions":`,
		`"size":1024`,
		`"createdTime":"2024-01-01T00:00:00Z"`,
		`"modifiedTime":"2024-01-01T00:00:00Z"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected field %s not found in JSON: %s", field, jsonStr)
		}
	}
}

func TestPermissionsOmitEmpty(t *testing.T) {
	// Test with empty permissions
	permissions := &Permissions{}

	jsonData, err := json.Marshal(permissions)
	if err != nil {
		t.Fatalf("Failed to marshal permissions: %v", err)
	}

	jsonStr := string(jsonData)

	// Should be empty object since all fields are empty
	expected := "{}"
	if jsonStr != expected {
		t.Errorf("Expected empty permissions to marshal to %s, got %s", expected, jsonStr)
	}

	// Test with some fields populated
	permissions = &Permissions{
		Group:         []Permission{PermissionRead},
		Authenticated: []Permission{PermissionWrite},
	}

	jsonData, err = json.Marshal(permissions)
	if err != nil {
		t.Fatalf("Failed to marshal permissions: %v", err)
	}

	jsonStr = string(jsonData)

	// Check that only populated fields are present
	if !strings.Contains(jsonStr, `"group":["Read"]`) {
		t.Errorf("Expected group field not found in JSON: %s", jsonStr)
	}

	if !strings.Contains(jsonStr, `"authenticated":["Write"]`) {
		t.Errorf("Expected authenticated field not found in JSON: %s", jsonStr)
	}

	// Check that empty fields are omitted
	if strings.Contains(jsonStr, `"anonymous"`) {
		t.Errorf("Empty anonymous field should be omitted from JSON: %s", jsonStr)
	}

	if strings.Contains(jsonStr, `"advanced"`) {
		t.Errorf("Empty advanced field should be omitted from JSON: %s", jsonStr)
	}
}
