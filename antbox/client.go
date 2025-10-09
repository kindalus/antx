package antbox

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

type client struct {
	ServerURL string
	APIKey    string
	Root      string
	JWT       string
	client    *http.Client
	debug     bool
}

func (c *client) roundTrip(req *http.Request) (*http.Response, error) {
	if c.debug {
		contentType := req.Header.Get("Content-Type")
		isMultipart := strings.Contains(contentType, "multipart/form-data") ||
			strings.Contains(contentType, "multipart/mixed")

		if isMultipart {
			// For multipart requests, manually print headers without dumping body
			fmt.Printf("\nRequest >>>>>\n\n%s %s\n", req.Method, req.URL.String())
			for key, values := range req.Header {
				for _, value := range values {
					fmt.Printf("%s: %s\n", key, value)
				}
			}
			fmt.Printf("\n<multipart body>\n")
		} else {
			dump, err := httputil.DumpRequestOut(req, true)
			if err != nil {
				fmt.Println("Error dumping request:", err)
			} else {
				fmt.Printf("\nRequest >>>>>\n\n%s\n", string(dump))
			}
		}
	}

	resp, err := c.client.Do(req)

	if c.debug && resp != nil {
		contentType := resp.Header.Get("Content-Type")
		isMultipart := strings.Contains(contentType, "multipart/form-data") ||
			strings.Contains(contentType, "multipart/mixed")

		if isMultipart {
			// For multipart responses, manually print headers without dumping body
			fmt.Printf(">>>>> Response\n\n%s %s\n", resp.Proto, resp.Status)
			for key, values := range resp.Header {
				for _, value := range values {
					fmt.Printf("%s: %s\n", key, value)
				}
			}
			fmt.Printf("\n<multipart body>\n=====\n\n")
		} else {
			dump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				fmt.Println("Error dumping response:", err)
			} else {
				fmt.Printf(">>>>> Response\n\n%s\n=====\n\n", string(dump))
			}
		}
	}

	return resp, err
}

func (c *client) Login() error {
	if c.Root == "" {
		return fmt.Errorf("root password is not set")
	}

	loginData := fmt.Sprintf("%x", sha256.Sum256([]byte(c.Root)))

	req, err := http.NewRequest("POST", c.ServerURL+"/login/root", bytes.NewBufferString(loginData))
	if err != nil {
		return err
	}

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewHttpErrorWithRequestBody(resp, req, loginData)
	}

	var result struct {
		JWT string `json:"jwt"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	c.JWT = result.JWT
	return nil
}

func (c *client) GetNode(uuid string) (*Node, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/nodes/"+uuid, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}
	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) ListNodes(parent string) ([]Node, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/nodes?parent="+parent, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var nodes []Node
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (c *client) CreateFolder(parent, name string) (*Node, error) {
	newNode := NodeCreate{
		Title:    name,
		Parent:   parent,
		Mimetype: "application/vnd.antbox.folder",
	}

	jsonNode, err := json.Marshal(newNode)
	if err != nil {
		return nil, err
	}

	// Store request body for error reporting
	requestBodyStr := string(jsonNode)

	req, err := http.NewRequest("POST", c.ServerURL+"/nodes", bytes.NewBuffer(jsonNode))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}
	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) CreateSmartFolder(parent, name string, filters NodeFilters) (*Node, error) {
	// Create the request payload with filters - we need a custom struct since NodeCreate doesn't have filters
	payload := struct {
		Title    string      `json:"title"`
		Parent   string      `json:"parent"`
		Mimetype string      `json:"mimetype"`
		Filters  NodeFilters `json:"filters"`
	}{
		Title:    name,
		Parent:   parent,
		Mimetype: "application/vnd.antbox.smartfolder",
		Filters:  filters,
	}

	jsonNode, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Store request body for error reporting
	requestBodyStr := string(jsonNode)

	req, err := http.NewRequest("POST", c.ServerURL+"/nodes", bytes.NewBuffer(jsonNode))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}
	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) RemoveNode(uuid string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/nodes/"+uuid, nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	return nil
}

func (c *client) MoveNode(uuid, newParent string) error {
	updateData := map[string]string{
		"parent": newParent,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("PATCH", c.ServerURL+"/nodes/"+uuid, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	return nil
}

func (c *client) ChangeNodeName(uuid, newName string) error {
	updateData := map[string]string{
		"title": newName,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("PATCH", c.ServerURL+"/nodes/"+uuid, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	return nil
}

func (c *client) uploadMultipartFile(filePath string, metadata any, url string, expectedStatus int) (*bytes.Buffer, *multipart.Writer, error) {
	filePath, err := expandTilde(filePath)
	if err != nil {
		return nil, nil, err
	}

	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filepath.Base(filePath)))
	partHeader.Set("Content-Type", detectMimetype(filePath))
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return nil, nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, nil, err
	}

	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return nil, nil, err
		}

		err = writer.WriteField("metadata", string(metadataJSON))
		if err != nil {
			return nil, nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, nil, err
	}

	return &requestBody, writer, nil
}

func (c *client) CreateFile(path string, metadata NodeCreate) (*Node, error) {
	requestBody, writer, err := c.uploadMultipartFile(path, metadata, c.ServerURL+"/nodes/-/upload", http.StatusCreated)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/nodes/-/upload", requestBody)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, "<multipart body>")
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) CreateNode(node NodeCreate) (*Node, error) {
	jsonData, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/nodes", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	var result Node
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func detectMimetype(path string) string {

	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".ts") {
		return "application/javascript"
	}

	if strings.HasSuffix(path, ".json") {
		return "application/json"
	}

	m, err := mimetype.DetectFile(path)
	if err != nil {
		slog.Error("failed to detect mimetype", "path", path, "error", err)
		return "application/octet-stream"
	}

	return strings.Split(m.String(), ";")[0]
}

func (c *client) UpdateFile(uuid, filePath string) (*Node, error) {
	requestBody, writer, err := c.uploadMultipartFile(filePath, nil, c.ServerURL+"/nodes/"+uuid+"/-/upload", http.StatusOK)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", c.ServerURL+"/nodes/"+uuid+"/-/upload", requestBody)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "<multipart body>")
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) UpdateNode(uuid string, metadata NodeUpdate) (*Node, error) {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", c.ServerURL+"/nodes/"+uuid, bytes.NewBuffer(metadataJSON))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, string(metadataJSON))
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) FindNodes(filters NodeFilters, pageSize, pageToken int) (*NodeFilterResult, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageToken <= 0 {
		pageToken = 1
	}

	requestBody := map[string]any{
		"filters":   filters,
		"pageSize":  pageSize,
		"pageToken": pageToken,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/nodes/-/find", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	var result NodeFilterResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) EvaluateNode(uuid string) ([]Node, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/nodes/"+uuid+"/-/evaluate", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	// The evaluate endpoint returns a generic object, but for smartfolders
	// it should contain a "nodes" array similar to the find result
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Try to extract nodes array from the result
	if nodesInterface, exists := result["nodes"]; exists {
		// Convert the nodes interface to proper Node structs
		nodesBytes, err := json.Marshal(nodesInterface)
		if err != nil {
			return nil, err
		}

		var nodes []Node
		if err := json.Unmarshal(nodesBytes, &nodes); err != nil {
			return nil, err
		}

		return nodes, nil
	}

	// If no nodes array found, return empty slice
	return []Node{}, nil
}

func (c *client) DownloadNode(uuid, downloadPath string) error {
	// Use the export endpoint for downloading node content
	req, err := http.NewRequest("GET", c.ServerURL+"/nodes/"+uuid+"/-/export", nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	// Create the download directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(downloadPath), 0755)
	if err != nil {
		return err
	}

	// Create the file
	file, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	return err
}

func (c *client) SetAuthHeader(req *http.Request) {
	if c.JWT != "" {
		req.Header.Set("Authorization", "Bearer "+c.JWT)
	} else if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
}

func (c *client) GetBreadcrumbs(uuid string) ([]Node, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/nodes/"+uuid+"/-/breadcrumbs", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var breadcrumbs []Node
	if err := json.NewDecoder(resp.Body).Decode(&breadcrumbs); err != nil {
		return nil, err
	}

	return breadcrumbs, nil
}

func (c *client) ChatWithAgent(agentUUID string, message string, conversationID string, temperature *float64, maxTokens *int, history []map[string]any) (ChatHistory, error) {
	options := make(map[string]any)

	if conversationID != "" && len(history) > 0 {
		// Add conversation history if we have a conversationID and history
		options["history"] = history
	}
	if temperature != nil {
		options["temperature"] = *temperature
	}
	if maxTokens != nil {
		options["maxTokens"] = *maxTokens
	}

	payload := map[string]any{
		"text": message,
	}

	if len(options) > 0 {
		payload["options"] = options
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/agents/%s/-/chat", agentUUID)
	req, err := http.NewRequest("POST", c.ServerURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	// Try to decode as ChatHistory format
	var chatHistory ChatHistory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &chatHistory); err == nil {
		return chatHistory, nil
	}

	// Try to decode as array of maps (legacy format)
	var arrayResult []map[string]any
	if err := json.Unmarshal(body, &arrayResult); err == nil {
		// Convert to ChatHistory format
		var history ChatHistory
		for _, msg := range arrayResult {
			chatMsg := ChatMessage{}

			if role, ok := msg["role"].(string); ok {
				chatMsg.Role = ChatMessageRole(role)
			}

			if parts, ok := msg["parts"].([]any); ok {
				for _, partAny := range parts {
					if partMap, ok := partAny.(map[string]any); ok {
						part := ChatMessagePart{}
						if text, ok := partMap["text"].(string); ok {
							part.Text = &text
						}
						if toolCall, ok := partMap["toolCall"].(map[string]any); ok {
							tc := &ToolCall{}
							if name, ok := toolCall["name"].(string); ok {
								tc.Name = name
							}
							if args, ok := toolCall["args"].(map[string]interface{}); ok {
								tc.Args = args
							}
							part.ToolCall = tc
						}
						if toolResponse, ok := partMap["toolResponse"].(map[string]any); ok {
							tr := &ToolResponse{}
							if name, ok := toolResponse["name"].(string); ok {
								tr.Name = name
							}
							if text, ok := toolResponse["text"].(string); ok {
								tr.Text = text
							}
							part.ToolResponse = tr
						}
						chatMsg.Parts = append(chatMsg.Parts, part)
					}
				}
			}
			history = append(history, chatMsg)
		}
		return history, nil
	}

	// Try to decode as object (traditional format)
	var objectResult map[string]any
	if err := json.Unmarshal(body, &objectResult); err == nil {
		// Convert single response to ChatHistory format
		var history ChatHistory
		if response, ok := objectResult["response"]; ok {
			if responseStr, ok := response.(string); ok {
				chatMsg := ChatMessage{
					Role: ChatMessageRoleModel,
					Parts: []ChatMessagePart{
						{Text: &responseStr},
					},
				}
				history = append(history, chatMsg)
				return history, nil
			}
		}
	}

	// If all parsing fails, return empty history
	return ChatHistory{}, nil
}

func (c *client) AnswerFromAgent(agentUUID string, query string, temperature *float64, maxTokens *int) (ChatHistory, error) {
	options := make(map[string]any)

	if temperature != nil {
		options["temperature"] = *temperature
	}
	if maxTokens != nil {
		options["maxTokens"] = *maxTokens
	}

	payload := map[string]any{
		"text": query,
	}

	if len(options) > 0 {
		payload["options"] = options
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/agents/%s/-/answer", agentUUID)
	req, err := http.NewRequest("POST", c.ServerURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	// Try to decode as ChatHistory format
	var chatHistory ChatHistory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &chatHistory); err == nil {
		return chatHistory, nil
	}

	// Try to decode as object (traditional format)
	var objectResult map[string]any
	if err := json.Unmarshal(body, &objectResult); err == nil {
		// Convert single response to ChatHistory format
		var history ChatHistory
		if response, ok := objectResult["response"]; ok {
			if responseStr, ok := response.(string); ok {
				chatMsg := ChatMessage{
					Role: ChatMessageRoleModel,
					Parts: []ChatMessagePart{
						{Text: &responseStr},
					},
				}
				history = append(history, chatMsg)
				return history, nil
			}
		}
	}

	// If all parsing fails, return empty history
	return ChatHistory{}, nil
}

func (c *client) RagChat(message string, options map[string]any) (ChatHistory, error) {

	payload := map[string]any{
		"text": message,
	}

	if len(options) > 0 {
		payload["options"] = options
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/agents/rag/-/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	// Try to decode as ChatHistory format
	var chatHistory ChatHistory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &chatHistory); err == nil {
		return chatHistory, nil
	}

	// Try to decode as array of maps (legacy format)
	var arrayResult []map[string]any
	if err := json.Unmarshal(body, &arrayResult); err == nil {
		// Convert to ChatHistory format
		var history ChatHistory
		for _, msg := range arrayResult {
			chatMsg := ChatMessage{}

			if role, ok := msg["role"].(string); ok {
				chatMsg.Role = ChatMessageRole(role)
			}

			if parts, ok := msg["parts"].([]any); ok {
				for _, partAny := range parts {
					if partMap, ok := partAny.(map[string]any); ok {
						part := ChatMessagePart{}
						if text, ok := partMap["text"].(string); ok {
							part.Text = &text
						}
						if toolCall, ok := partMap["toolCall"].(map[string]any); ok {
							tc := &ToolCall{}
							if name, ok := toolCall["name"].(string); ok {
								tc.Name = name
							}
							if args, ok := toolCall["args"].(map[string]interface{}); ok {
								tc.Args = args
							}
							part.ToolCall = tc
						}
						if toolResponse, ok := partMap["toolResponse"].(map[string]any); ok {
							tr := &ToolResponse{}
							if name, ok := toolResponse["name"].(string); ok {
								tr.Name = name
							}
							if text, ok := toolResponse["text"].(string); ok {
								tr.Text = text
							}
							part.ToolResponse = tr
						}
						chatMsg.Parts = append(chatMsg.Parts, part)
					}
				}
			}
			history = append(history, chatMsg)
		}
		return history, nil
	}

	// Try to decode as object (traditional format)
	var objectResult map[string]any
	if err := json.Unmarshal(body, &objectResult); err == nil {
		// Convert single response to ChatHistory format
		var history ChatHistory
		if response, ok := objectResult["response"]; ok {
			if responseStr, ok := response.(string); ok {
				chatMsg := ChatMessage{
					Role: ChatMessageRoleModel,
					Parts: []ChatMessagePart{
						{Text: &responseStr},
					},
				}
				history = append(history, chatMsg)
				return history, nil
			}
		}
	}

	// If all parsing fails, return empty history
	return ChatHistory{}, nil
}

func (c *client) CopyNode(uuid, parent, title string) (*Node, error) {
	payload := map[string]string{
		"to": parent,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/nodes/"+uuid+"/-/copy", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) DuplicateNode(uuid string) (*Node, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/nodes/"+uuid+"/-/duplicate", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) ExportNode(uuid string, format string) ([]byte, error) {
	url := c.ServerURL + "/nodes/" + uuid + "/-/export"
	if format != "" {
		url += "?format=" + format
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Feature operations
func (c *client) ListFeatures() ([]Feature, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/features", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var features []Feature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, err
	}

	return features, nil
}

func (c *client) GetFeature(uuid string) (*Feature, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/features/"+uuid, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var feature Feature
	if err := json.NewDecoder(resp.Body).Decode(&feature); err != nil {
		return nil, err
	}

	return &feature, nil
}

func (c *client) DeleteFeature(uuid string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/features/"+uuid, nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	return nil
}

func (c *client) ExportFeature(uuid string, exportType string) (string, error) {
	url := c.ServerURL + "/features/" + uuid + "/export"
	if exportType != "" {
		url += "?type=" + exportType
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewHttpErrorWithRequestBody(resp, req, "")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *client) ListActionFeatures() ([]Feature, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/features/-/actions", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var features []Feature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, err
	}

	return features, nil
}

func (c *client) ListExtensionFeatures() ([]Feature, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/features/-/extensions", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var features []Feature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, err
	}

	return features, nil
}

func (c *client) RunFeatureAsAction(uuid string, uuids []string) (map[string]any, error) {
	url := c.ServerURL + "/features/" + uuid + "/-/run-action?uuids=" + strings.Join(uuids, ",")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *client) RunFeatureAsExtension(uuid string, params map[string]any) (string, error) {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/extensions/"+uuid+"/-/exec", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Action operations
func (c *client) ListActions() ([]Feature, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/actions", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var features []Feature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, err
	}

	return features, nil
}

func (c *client) RunAction(uuid string, request ActionRunRequest) (map[string]any, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/actions/"+uuid+"/run", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Extension operations
func (c *client) ListExtensions() ([]Feature, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/extensions", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var features []Feature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, err
	}

	return features, nil
}

func (c *client) RunExtension(uuid string, data map[string]any) (any, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/extensions/"+uuid+"/run", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// AI Tool operations
func (c *client) ListAITools() ([]Feature, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/ai-tools", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var features []Feature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, err
	}

	return features, nil
}

func (c *client) RunAITool(uuid string, params map[string]any) (map[string]any, error) {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/ai-tools/"+uuid+"/run", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Agent operations
func (c *client) ListAgents() ([]Agent, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/agents", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var agents []Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, err
	}

	return agents, nil
}

func (c *client) GetAgent(uuid string) (*Agent, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/agents/"+uuid, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var agent Agent
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *client) DeleteAgent(uuid string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/agents/"+uuid, nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	return nil
}

// API Key operations
func (c *client) ListAPIKeys() ([]APIKey, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/api-keys", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var apiKeys []APIKey
	if err := json.NewDecoder(resp.Body).Decode(&apiKeys); err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (c *client) CreateAPIKey(request APIKeyCreate) (*APIKey, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/api-keys", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result APIKey
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetAPIKey(uuid string) (*APIKey, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/api-keys/"+uuid, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var apiKey APIKey
	if err := json.NewDecoder(resp.Body).Decode(&apiKey); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

func (c *client) DeleteAPIKey(uuid string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/api-keys/"+uuid, nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	return nil
}

// User operations
func (c *client) ListUsers() ([]User, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/users", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *client) CreateUser(user UserCreate) (*User, error) {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/users", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetUser(email string) (*User, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/users/"+email, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *client) UpdateUser(email string, user UserUpdate) (*User, error) {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("PUT", c.ServerURL+"/users/"+email, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) DeleteUser(uuid string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/users/"+uuid, nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	return nil
}

// Group operations
func (c *client) ListGroups() ([]Group, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/groups", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *client) CreateGroup(group GroupCreate) (*Group, error) {
	jsonData, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/groups", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result Group
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetGroup(uuid string) (*Group, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/groups/"+uuid, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var group Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *client) UpdateGroup(uuid string, group GroupUpdate) (*Group, error) {
	jsonData, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("PUT", c.ServerURL+"/groups/"+uuid, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var result Group
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) DeleteGroup(uuid string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/groups/"+uuid, nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	return nil
}

// Template operations
func (c *client) ListTemplates() ([]Template, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/templates", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var templates []Template
	if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (c *client) GetTemplate(uuid string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/templates/"+uuid, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Documentation operations
func (c *client) ListDocs() ([]DocInfo, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/docs", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var docs []DocInfo
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, err
	}

	return docs, nil
}

func (c *client) GetDoc(uuid string) (string, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/docs/"+uuid, nil)
	if err != nil {
		return "", err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewHttpErrorWithRequestBody(resp, req, "")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Aspect operations
func (c *client) ListAspects() ([]Aspect, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/aspects", nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var aspects []Aspect
	if err := json.NewDecoder(resp.Body).Decode(&aspects); err != nil {
		return nil, err
	}

	return aspects, nil
}

func (c *client) GetAspect(uuid string) (*Aspect, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/aspects/"+uuid, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var aspect Aspect
	if err := json.NewDecoder(resp.Body).Decode(&aspect); err != nil {
		return nil, err
	}

	return &aspect, nil
}

func (c *client) DeleteAspect(uuid string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/aspects/"+uuid, nil)
	if err != nil {
		return err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return NewHttpErrorWithRequestBody(resp, req, "")
	}

	return nil
}

func (c *client) ExportAspect(uuid string, format string) (any, error) {
	url := c.ServerURL + "/aspects/" + uuid + "/-/export"
	if format != "" {
		url += "?format=" + format
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *client) UploadAspect(filePath string) (*Aspect, error) {
	requestBody, writer, err := c.uploadMultipartFile(filePath, nil, c.ServerURL+"/aspects/-/upload", http.StatusCreated)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/aspects/-/upload", requestBody)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, "")
	}

	var result Aspect
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) UploadFeature(filePath string) (*Feature, error) {
	requestBody, writer, err := c.uploadMultipartFile(filePath, nil, c.ServerURL+"/features/-/upload", http.StatusCreated)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/features/-/upload", requestBody)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, "<multipart body>")
	}

	var feature Feature
	if err := json.NewDecoder(resp.Body).Decode(&feature); err != nil {
		return nil, err
	}

	return &feature, nil
}

func (c *client) UploadAgent(filePath string) (*Agent, error) {
	requestBody, writer, err := c.uploadMultipartFile(filePath, nil, c.ServerURL+"/agents/-/upload", http.StatusCreated)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/agents/-/upload", requestBody)
	if err != nil {
		return nil, err
	}

	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, NewHttpErrorWithRequestBody(resp, req, "<multipart body>")
	}

	var agent Agent
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return currentUser.HomeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		return strings.Replace(path, "~", currentUser.HomeDir, 1), nil
	}

	return path, nil
}
