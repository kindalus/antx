package antbox

import (
	"fmt"
	"io"
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

	client := NewClient(server.URL, "", "test-password", "", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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
		if r.URL.Path != "/nodes/-/upload" {
			t.Errorf("Expected to request '/nodes/-/upload', got %s", r.URL.Path)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
	node, err := client.CreateFile(tempFile.Name(), NodeCreate{Parent: "parent-uuid"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if node.UUID != "uploaded-uuid" {
		t.Errorf("Expected uploaded UUID 'uploaded-uuid', got '%s'", node.UUID)
	}
}

func TestUpdateFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-update-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString("updated content"); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/file-uuid/-/upload" {
			t.Errorf("Expected to request '/nodes/file-uuid/-/upload', got %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("Expected 'PUT' request, got '%s'", r.Method)
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Expected multipart/form-data content type, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"uuid":"file-uuid","title":"updated.txt"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	node, err := client.UpdateFile("file-uuid", tempFile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if node.UUID != "file-uuid" {
		t.Fatalf("Expected updated file node, got %#v", node)
	}
}

func TestFindNodesOpenAPIArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/-/find" {
			t.Errorf("Expected to request '/nodes/-/find', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		if !strings.Contains(string(body), `"filters":[["title","match","doc"]]`) {
			t.Errorf("Expected typed filters payload, got %s", string(body))
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"uuid":"node-uuid","title":"Document"}]`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	result, err := client.FindNodes(NodeFilters1D{{"title", FilterOperatorMatch, "doc"}}, 20, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result.Nodes) != 1 || result.Nodes[0].UUID != "node-uuid" {
		t.Fatalf("Expected one node from array response, got %#v", result)
	}
}

func TestAnswerFromAgentOpenAPIChatMessageResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/agent-uuid/-/answer" {
			t.Errorf("Expected to request '/agents/agent-uuid/-/answer', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"role":"model","parts":[{"text":"Answer text"}]}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	history, err := client.AnswerFromAgent("agent-uuid", "question", nil, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(history) != 1 || history[0].Role != ChatMessageRoleModel || history[0].Parts[0].Text == nil || *history[0].Parts[0].Text != "Answer text" {
		t.Fatalf("Expected single ChatMessage response converted to history, got %#v", history)
	}
}

func TestUploadAgentUsesJSONPayload(t *testing.T) {
	tempFile, err := os.CreateTemp("", "agent-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(`{"name":"Support Agent","model":"default"}`); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/-/upload" {
			t.Errorf("Expected to request '/agents/-/upload', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		if !strings.Contains(string(body), `"name":"Support Agent"`) {
			t.Errorf("Expected agent JSON payload, got %s", string(body))
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, `{"uuid":"agent-uuid","name":"Support Agent","exposedToUsers":true,"createdTime":"2026-01-01T00:00:00Z","modifiedTime":"2026-01-01T00:00:00Z"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	agent, err := client.UploadAgent(tempFile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if agent.UUID != "agent-uuid" || agent.Name != "Support Agent" {
		t.Fatalf("Expected uploaded agent response, got %#v", agent)
	}
}

func TestEvaluateNodeOpenAPIArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/smart-uuid/-/evaluate" {
			t.Errorf("Expected to request '/nodes/smart-uuid/-/evaluate', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected 'GET' request, got '%s'", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"uuid":"node-uuid","title":"Evaluated Node"}]`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	nodes, err := client.EvaluateNode("smart-uuid")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].UUID != "node-uuid" {
		t.Fatalf("Expected one evaluated node, got %#v", nodes)
	}
}

func TestUpdateUserUsesPatchAndTitlePayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/test@example.com" {
			t.Errorf("Expected to request '/users/test@example.com', got %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Expected 'PATCH' request, got '%s'", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		if !strings.Contains(string(body), `"title":"Updated User"`) {
			t.Errorf("Expected title in update payload, got %s", string(body))
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"email":"test@example.com","title":"Updated User","group":"group-uuid","groups":[],"hasWhatsapp":false,"active":true,"createdTime":"2026-01-01T00:00:00Z","modifiedTime":"2026-01-01T00:00:00Z"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	user, err := client.UpdateUser("test@example.com", UserUpdate{Name: "Updated User"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if user.Email != "test@example.com" || user.Title != "Updated User" {
		t.Fatalf("Expected updated user response, got %#v", user)
	}
}

func TestDisplayNameFallbacks(t *testing.T) {
	agent := Agent{Name: "New Name", Title: "Legacy Title"}
	if agent.DisplayName() != "New Name" {
		t.Fatalf("Expected agent name preference, got %q", agent.DisplayName())
	}

	agent = Agent{Title: "Legacy Title"}
	if agent.DisplayName() != "Legacy Title" {
		t.Fatalf("Expected agent title fallback, got %q", agent.DisplayName())
	}

	user := User{Title: "New User", Name: "Legacy User"}
	if user.DisplayName() != "New User" {
		t.Fatalf("Expected user title preference, got %q", user.DisplayName())
	}

	user = User{Name: "Legacy User"}
	if user.DisplayName() != "Legacy User" {
		t.Fatalf("Expected user name fallback, got %q", user.DisplayName())
	}
}

func TestDownloadNode(t *testing.T) {
	testContent := "This is downloaded content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/test-uuid/-/export" {
			t.Errorf("Expected to request '/nodes/test-uuid/-/export', got %s", r.URL.Path)
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

	client := NewClient(server.URL, "", "", "test-jwt", false)
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

func TestGetAuditLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audit/node-uuid" {
			t.Errorf("Expected to request '/audit/node-uuid', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected 'GET' request, got '%s'", r.Method)
		}
		if got := r.URL.Query().Get("mimetype"); got != "application/pdf" {
			t.Errorf("Expected mimetype query 'application/pdf', got %q", got)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"streamId":"node-uuid","eventId":"event-uuid","eventType":"NodeUpdatedEvent","occurredOn":"2026-05-14T10:00:00Z","userEmail":"user@example.com","tenant":"default","payload":{"title":{"old":"A","new":"B"}},"sequence":2}]`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	events, err := client.GetAuditLog("node-uuid", "application/pdf")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "NodeUpdatedEvent" || events[0].Sequence != 2 {
		t.Fatalf("Expected decoded audit event, got %#v", events)
	}
	if events[0].Payload["title"] == nil {
		t.Fatalf("Expected payload to be decoded, got %#v", events[0].Payload)
	}
}

func TestSupportedClientEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		switch key {
		case "GET /login/me":
			fmt.Fprintln(w, `{"email":"me@example.com","name":"Me","groups":["users"]}`)
		case "POST /nodes":
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintln(w, `{"uuid":"node-created","title":"Created","mimetype":"application/vnd.antbox.folder"}`)
		case "PATCH /nodes/node-uuid":
			fmt.Fprintln(w, `{"uuid":"node-uuid","title":"Updated"}`)
		case "GET /nodes/node-uuid/-/breadcrumbs":
			fmt.Fprintln(w, `[{"uuid":"--root--","title":"root"},{"uuid":"node-uuid","title":"Node"}]`)
		case "POST /agents/agent-uuid/-/chat":
			fmt.Fprintln(w, `[{"role":"model","parts":[{"text":"chat response"}]}]`)
		case "POST /nodes/node-uuid/-/copy":
			fmt.Fprintln(w, `{"uuid":"copied-uuid","title":"Copied"}`)
		case "GET /nodes/node-uuid/-/duplicate":
			fmt.Fprintln(w, `{"uuid":"cloned-uuid","title":"Cloned"}`)
		case "GET /nodes/node-uuid/-/export":
			fmt.Fprint(w, "exported")
		case "GET /agents":
			fmt.Fprintln(w, `[{"uuid":"agent-uuid","name":"Agent","exposedToUsers":true,"createdTime":"2026-01-01T00:00:00Z","modifiedTime":"2026-01-01T00:00:00Z"}]`)
		case "GET /agents/agent-uuid":
			fmt.Fprintln(w, `{"uuid":"agent-uuid","name":"Agent","exposedToUsers":true,"createdTime":"2026-01-01T00:00:00Z","modifiedTime":"2026-01-01T00:00:00Z"}`)
		case "DELETE /agents/agent-uuid", "DELETE /api-keys/key-uuid", "DELETE /users/user-uuid", "DELETE /groups/group-uuid":
			w.WriteHeader(http.StatusOK)
		case "GET /api-keys":
			fmt.Fprintln(w, `[{"uuid":"key-uuid","title":"Key","secret":"secret","group":"group-uuid","active":true,"createdTime":"2026-01-01T00:00:00Z"}]`)
		case "POST /api-keys":
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintln(w, `{"uuid":"key-uuid","title":"Key","secret":"secret","group":"group-uuid","active":true,"createdTime":"2026-01-01T00:00:00Z"}`)
		case "GET /api-keys/key-uuid":
			fmt.Fprintln(w, `{"uuid":"key-uuid","title":"Key","secret":"secret","group":"group-uuid","active":true,"createdTime":"2026-01-01T00:00:00Z"}`)
		case "GET /users":
			fmt.Fprintln(w, `[{"email":"user@example.com","title":"User","group":"group-uuid","groups":[],"hasWhatsapp":false,"active":true,"createdTime":"2026-01-01T00:00:00Z","modifiedTime":"2026-01-01T00:00:00Z"}]`)
		case "POST /users":
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintln(w, `{"email":"new@example.com","title":"New User","group":"group-uuid","groups":[],"hasWhatsapp":false,"active":true,"createdTime":"2026-01-01T00:00:00Z","modifiedTime":"2026-01-01T00:00:00Z"}`)
		case "GET /users/user@example.com":
			fmt.Fprintln(w, `{"email":"user@example.com","title":"User","group":"group-uuid","groups":[],"hasWhatsapp":false,"active":true,"createdTime":"2026-01-01T00:00:00Z","modifiedTime":"2026-01-01T00:00:00Z"}`)
		case "GET /groups":
			fmt.Fprintln(w, `[{"uuid":"group-uuid","title":"Group","createdTime":"2026-01-01T00:00:00Z"}]`)
		case "POST /groups":
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintln(w, `{"uuid":"group-uuid","title":"Group","createdTime":"2026-01-01T00:00:00Z"}`)
		case "GET /groups/group-uuid":
			fmt.Fprintln(w, `{"uuid":"group-uuid","title":"Group","createdTime":"2026-01-01T00:00:00Z"}`)
		case "GET /docs":
			fmt.Fprintln(w, `[{"uuid":"doc-uuid","description":"Doc"}]`)
		case "GET /docs/doc-uuid":
			fmt.Fprint(w, "# Doc")
		default:
			t.Fatalf("Unexpected request %s", key)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "", "test-jwt", false)
	if user, err := client.GetCurrentUser(); err != nil || user.Email != "me@example.com" {
		t.Fatalf("GetCurrentUser = %#v, %v", user, err)
	}
	if node, err := client.CreateSmartFolder("--root--", "Smart", NodeFilters1D{{"title", FilterOperatorMatch, "x"}}); err != nil || node.UUID == "" {
		t.Fatalf("CreateSmartFolder = %#v, %v", node, err)
	}
	if node, err := client.CreateNode(NodeCreate{Title: "Created", Mimetype: "application/vnd.antbox.folder"}); err != nil || node.UUID == "" {
		t.Fatalf("CreateNode = %#v, %v", node, err)
	}
	if node, err := client.UpdateNode("node-uuid", NodeUpdate{Title: "Updated"}); err != nil || node.Title != "Updated" {
		t.Fatalf("UpdateNode = %#v, %v", node, err)
	}
	if breadcrumbs, err := client.GetBreadcrumbs("node-uuid"); err != nil || len(breadcrumbs) != 2 {
		t.Fatalf("GetBreadcrumbs = %#v, %v", breadcrumbs, err)
	}
	if history, err := client.ChatWithAgent("agent-uuid", "hello", "", nil, nil, nil); err != nil || len(history) != 1 {
		t.Fatalf("ChatWithAgent = %#v, %v", history, err)
	}
	if node, err := client.CopyNode("node-uuid", "--root--", "Copied"); err != nil || node.UUID != "copied-uuid" {
		t.Fatalf("CopyNode = %#v, %v", node, err)
	}
	if node, err := client.CloneNode("node-uuid"); err != nil || node.UUID != "cloned-uuid" {
		t.Fatalf("CloneNode = %#v, %v", node, err)
	}
	if data, err := client.ExportNode("node-uuid", ""); err != nil || string(data) != "exported" {
		t.Fatalf("ExportNode = %q, %v", string(data), err)
	}
	if agents, err := client.ListAgents(); err != nil || len(agents) != 1 {
		t.Fatalf("ListAgents = %#v, %v", agents, err)
	}
	if agent, err := client.GetAgent("agent-uuid"); err != nil || agent.UUID != "agent-uuid" {
		t.Fatalf("GetAgent = %#v, %v", agent, err)
	}
	if err := client.DeleteAgent("agent-uuid"); err != nil {
		t.Fatalf("DeleteAgent = %v", err)
	}
	if keys, err := client.ListAPIKeys(); err != nil || len(keys) != 1 {
		t.Fatalf("ListAPIKeys = %#v, %v", keys, err)
	}
	if key, err := client.CreateAPIKey(APIKeyCreate{Group: "group-uuid"}); err != nil || key.UUID != "key-uuid" {
		t.Fatalf("CreateAPIKey = %#v, %v", key, err)
	}
	if key, err := client.GetAPIKey("key-uuid"); err != nil || key.UUID != "key-uuid" {
		t.Fatalf("GetAPIKey = %#v, %v", key, err)
	}
	if err := client.DeleteAPIKey("key-uuid"); err != nil {
		t.Fatalf("DeleteAPIKey = %v", err)
	}
	if users, err := client.ListUsers(); err != nil || len(users) != 1 {
		t.Fatalf("ListUsers = %#v, %v", users, err)
	}
	if user, err := client.CreateUser(UserCreate{Email: "new@example.com", Name: "New User", Group: "group-uuid"}); err != nil || user.Email != "new@example.com" {
		t.Fatalf("CreateUser = %#v, %v", user, err)
	}
	if user, err := client.GetUser("user@example.com"); err != nil || user.Email != "user@example.com" {
		t.Fatalf("GetUser = %#v, %v", user, err)
	}
	if err := client.DeleteUser("user-uuid"); err != nil {
		t.Fatalf("DeleteUser = %v", err)
	}
	if groups, err := client.ListGroups(); err != nil || len(groups) != 1 {
		t.Fatalf("ListGroups = %#v, %v", groups, err)
	}
	if group, err := client.CreateGroup(GroupCreate{Title: "Group"}); err != nil || group.UUID != "group-uuid" {
		t.Fatalf("CreateGroup = %#v, %v", group, err)
	}
	if group, err := client.GetGroup("group-uuid"); err != nil || group.UUID != "group-uuid" {
		t.Fatalf("GetGroup = %#v, %v", group, err)
	}
	if err := client.DeleteGroup("group-uuid"); err != nil {
		t.Fatalf("DeleteGroup = %v", err)
	}
	if docs, err := client.ListDocs(); err != nil || len(docs) != 1 {
		t.Fatalf("ListDocs = %#v, %v", docs, err)
	}
	if doc, err := client.GetDoc("doc-uuid"); err != nil || doc != "# Doc" {
		t.Fatalf("GetDoc = %q, %v", doc, err)
	}
}
