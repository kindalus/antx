package antbox

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login/root" {
			t.Errorf("Expected to request '/login/root', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"jwt":"test-jwt"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "test-password", "")
	if err := client.Login(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/test-uuid" {
			t.Errorf("Expected to request '/nodes/test-uuid', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected 'GET' request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"uuid":"test-uuid","title":"test-title"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	node, err := client.GetNode("test-uuid")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if node.UUID != "test-uuid" {
		t.Errorf("Expected node UUID to be 'test-uuid', got '%s'", node.UUID)
	}

	if node.Title != "test-title" {
		t.Errorf("Expected node title to be 'test-title', got '%s'", node.Title)
	}
}

func TestListNodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes" {
			t.Errorf("Expected to request '/nodes', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected 'GET' request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"uuid":"test-uuid","title":"test-title"}]`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	nodes, err := client.ListNodes("--root--")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	if nodes[0].UUID != "test-uuid" {
		t.Errorf("Expected node UUID to be 'test-uuid', got '%s'", nodes[0].UUID)
	}
}

func TestHttpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `{"error":"Node not found"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	_, err := client.GetNode("non-existent-uuid")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	httpErr, ok := err.(*HttpError)
	if !ok {
		t.Errorf("Expected HttpError, got %T", err)
	}

	if httpErr.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", httpErr.StatusCode)
	}

	if httpErr.Method != "GET" {
		t.Errorf("Expected method GET, got %s", httpErr.Method)
	}

	// Test the new formatted error message
	errorMsg := httpErr.Error()

	// Check for error header format
	expectedHeader := fmt.Sprintf("Error: GET %s/nodes/non-existent-uuid - 404", server.URL)
	if !strings.Contains(errorMsg, expectedHeader) {
		t.Errorf("Expected error header not found.\nExpected: %s\nGot: %s", expectedHeader, errorMsg)
	}

	// Check for request section
	if !strings.Contains(errorMsg, "==> Request") {
		t.Error("Expected '==> Request' section not found")
	}

	// Check for response section
	if !strings.Contains(errorMsg, "Response <==") {
		t.Error("Expected 'Response <==' section not found")
	}

	// Check for pretty-printed JSON in response body
	expectedPrettyJSON := `{
  "error": "Node not found"
}`
	if !strings.Contains(errorMsg, expectedPrettyJSON) {
		t.Errorf("Expected pretty-printed JSON not found.\nExpected:\n%s\nGot:\n%s", expectedPrettyJSON, errorMsg)
	}

	// Check for Content-Type header in response
	if !strings.Contains(errorMsg, "Content-Type: application/json") {
		t.Error("Expected Content-Type header not found in response section")
	}
}

func TestHttpErrorJSONPrettyPrint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "AntboxAPI/1.0")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error":{"message":"Invalid request","code":400,"details":["Missing required field","Invalid format"]}}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	_, err := client.GetNode("invalid-uuid")

	httpErr, ok := err.(*HttpError)
	if !ok {
		t.Fatalf("Expected HttpError, got %T", err)
	}

	errorMsg := httpErr.Error()

	// Check error header
	expectedHeader := fmt.Sprintf("Error: GET %s/nodes/invalid-uuid - 400", server.URL)
	if !strings.Contains(errorMsg, expectedHeader) {
		t.Errorf("Expected error header not found.\nExpected: %s\nGot: %s", expectedHeader, errorMsg)
	}

	// Check for both request and response sections
	if !strings.Contains(errorMsg, "==> Request") {
		t.Error("Expected '==> Request' section not found")
	}

	if !strings.Contains(errorMsg, "Response <==") {
		t.Error("Expected 'Response <==' section not found")
	}

	// Check pretty-printed JSON
	expectedPrettyJSON := `{
  "error": {
    "code": 400,
    "details": [
      "Missing required field",
      "Invalid format"
    ],
    "message": "Invalid request"
  }
}`

	if !strings.Contains(errorMsg, expectedPrettyJSON) {
		t.Errorf("Expected pretty-printed JSON not found in error message.\nExpected:\n%s\nGot:\n%s", expectedPrettyJSON, errorMsg)
	}

	// Check response headers
	if !strings.Contains(errorMsg, "Server: AntboxAPI/1.0") {
		t.Error("Expected Server header not found in response section")
	}
}

func TestHttpErrorWithRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error":"Invalid folder name"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	_, err := client.CreateFolder("parent-uuid", "invalid/name")

	httpErr, ok := err.(*HttpError)
	if !ok {
		t.Fatalf("Expected HttpError, got %T", err)
	}

	errorMsg := httpErr.Error()

	// Check error header
	expectedHeader := fmt.Sprintf("Error: POST %s/nodes - 400", server.URL)
	if !strings.Contains(errorMsg, expectedHeader) {
		t.Errorf("Expected error header not found.\nExpected: %s\nGot: %s", expectedHeader, errorMsg)
	}

	// Check request section with body
	if !strings.Contains(errorMsg, "==> Request") {
		t.Error("Expected '==> Request' section not found")
	}

	// Check that request body contains only the non-empty fields
	expectedRequestBody := `{
  "mimetype": "application/vnd.antbox.folder",
  "parent": "parent-uuid",
  "title": "invalid/name"
}`
	if !strings.Contains(errorMsg, expectedRequestBody) {
		t.Errorf("Expected clean request body not found.\nExpected:\n%s\nGot:\n%s", expectedRequestBody, errorMsg)
	}

	// Check request headers
	if !strings.Contains(errorMsg, "Content-Type: application/json") {
		t.Error("Expected Content-Type: application/json in request headers")
	}

	// Check response section
	if !strings.Contains(errorMsg, "Response <==") {
		t.Error("Expected 'Response <==' section not found")
	}

	// Check response body
	expectedResponseBody := `{
  "error": "Invalid folder name"
}`
	if !strings.Contains(errorMsg, expectedResponseBody) {
		t.Errorf("Expected response body not found.\nExpected:\n%s\nGot:\n%s", expectedResponseBody, errorMsg)
	}
}

func TestHttpErrorNonJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal server error occurred")
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	_, err := client.GetNode("invalid-uuid")

	httpErr, ok := err.(*HttpError)
	if !ok {
		t.Fatalf("Expected HttpError, got %T", err)
	}

	errorMsg := httpErr.Error()

	// Check error header
	expectedHeader := fmt.Sprintf("Error: GET %s/nodes/invalid-uuid - 500", server.URL)
	if !strings.Contains(errorMsg, expectedHeader) {
		t.Errorf("Expected error header not found.\nExpected: %s\nGot: %s", expectedHeader, errorMsg)
	}

	// Check that non-JSON body remains unchanged
	expectedBody := "Internal server error occurred"
	if !strings.Contains(errorMsg, expectedBody) {
		t.Errorf("Expected non-JSON body to remain unchanged, got: %s", errorMsg)
	}

	// Check Content-Type header
	if !strings.Contains(errorMsg, "Content-Type: text/plain") {
		t.Error("Expected Content-Type: text/plain header not found")
	}
}

func TestRemoveNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/test-uuid" {
			t.Errorf("Expected to request '/nodes/test-uuid', got %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Expected 'DELETE' request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	err := client.RemoveNode("test-uuid")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMoveNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/test-uuid" {
			t.Errorf("Expected to request '/nodes/test-uuid', got %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Expected 'PATCH' request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	err := client.MoveNode("test-uuid", "parent-uuid")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestChangeNodeName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/test-uuid" {
			t.Errorf("Expected to request '/nodes/test-uuid', got %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Expected 'PATCH' request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	err := client.ChangeNodeName("test-uuid", "new-name")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestUploadFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test-upload-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "This is a test file content"
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes" {
			t.Errorf("Expected to request '/nodes', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Expected multipart/form-data content type, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, `{"uuid":"uploaded-uuid","title":"test-file.txt"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt")
	node, err := client.CreateFile(tempFile.Name(), "parent-uuid")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if node.UUID != "uploaded-uuid" {
		t.Errorf("Expected uploaded UUID 'uploaded-uuid', got '%s'", node.UUID)
	}
}

func TestDownloadNode(t *testing.T) {
	testContent := "This is downloaded content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/test-uuid/content" {
			t.Errorf("Expected to request '/nodes/test-uuid/content', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected 'GET' request, got '%s'", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, testContent)
	}))
	defer server.Close()

	// Create a temporary directory for download
	tempDir, err := os.MkdirTemp("", "test-download-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	downloadPath := filepath.Join(tempDir, "downloaded-file.txt")

	client := NewClient(server.URL, "", "", "test-jwt")
	err = client.DownloadNode("test-uuid", downloadPath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify file was created and has correct content
	content, err := os.ReadFile(downloadPath)
	if err != nil {
		t.Errorf("Failed to read downloaded file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Expected downloaded content '%s', got '%s'", testContent, string(content))
	}
}
