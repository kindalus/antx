package antbox

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			fmt.Println("Error dumping request:", err)
		} else {
			fmt.Printf("\nRequest >>>>>\n\n%s\n", string(dump))
		}
	}

	resp, err := c.client.Do(req)

	if c.debug && resp != nil {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println("Error dumping response:", err)
		} else {
			fmt.Printf(">>>>> Response\n\n%s\n=====\n\n", string(dump))
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

	var result struct {
		Nodes []Node `json:"nodes"`
		Total int    `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Nodes, nil
}

func (c *client) CreateFolder(parent, name string) (*Node, error) {
	newNode := Node{
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

func (c *client) CreateSmartFolder(parent, name string, filters any) (*Node, error) {
	// Create the request payload with filters
	payload := map[string]any{
		"title":    name,
		"parent":   parent,
		"mimetype": "application/vnd.antbox.smartfolder",
		"filters":  filters,
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

func (c *client) CreateFile(path, parentUuid string) (*Node, error) {

	path, err := expandTilde(path)
	if err != nil {
		return nil, err
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	// Add the metadata with parent UUID as JSON
	metadata := map[string]string{
		"parent": parentUuid,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("metadata", string(metadataJSON))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	requestBodyStr := requestBody.String()

	req, err := http.NewRequest("POST", c.ServerURL+"/nodes/-/upload", &requestBody)
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
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) UpdateFile(uuid, filePath string) (*Node, error) {
	filePath, err := expandTilde(filePath)
	if err != nil {
		return nil, err
	}

	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	requestBodyStr := requestBody.String()

	req, err := http.NewRequest("PUT", c.ServerURL+"/nodes/"+uuid+"/-/upload", &requestBody)
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
		return nil, NewHttpErrorWithRequestBody(resp, req, requestBodyStr)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (c *client) FindNodes(filters any, pageSize, pageToken int) (*NodeFilterResult, error) {
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
	var result []Node
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
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

func (c *client) ChatWithAgent(agentUUID string, message string, conversationID string, temperature *float64, maxTokens *int) (string, error) {
	options := make(map[string]interface{})

	if conversationID != "" {
		// Add conversation history if we have a conversationID
		options["history"] = []map[string]interface{}{}
	}
	if temperature != nil {
		options["temperature"] = *temperature
	}
	if maxTokens != nil {
		options["maxTokens"] = *maxTokens
	}

	payload := map[string]interface{}{
		"text": message,
	}

	if len(options) > 0 {
		payload["options"] = options
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("/agents/%s/-/chat", agentUUID)
	req, err := http.NewRequest("POST", c.ServerURL+endpoint, bytes.NewBuffer(jsonData))
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
		return "", NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Extract the response message
	if response, ok := result["response"]; ok {
		if responseStr, ok := response.(string); ok {
			return responseStr, nil
		}
	}

	// Fallback to return the whole response as JSON string
	responseJson, _ := json.MarshalIndent(result, "", "  ")
	return string(responseJson), nil
}

func (c *client) AnswerFromAgent(agentUUID string, query string, temperature *float64, maxTokens *int) (string, error) {
	options := make(map[string]interface{})

	if temperature != nil {
		options["temperature"] = *temperature
	}
	if maxTokens != nil {
		options["maxTokens"] = *maxTokens
	}

	payload := map[string]interface{}{
		"text": query,
	}

	if len(options) > 0 {
		payload["options"] = options
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("/agents/%s/-/answer", agentUUID)
	req, err := http.NewRequest("POST", c.ServerURL+endpoint, bytes.NewBuffer(jsonData))
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
		return "", NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Extract the response message
	if response, ok := result["response"]; ok {
		if responseStr, ok := response.(string); ok {
			return responseStr, nil
		}
	}

	// Fallback to return the whole response as JSON string
	responseJson, _ := json.MarshalIndent(result, "", "  ")
	return string(responseJson), nil
}

func (c *client) RagChat(message string, conversationID string, filters map[string]interface{}) (string, error) {
	options := make(map[string]interface{})

	if conversationID != "" {
		// Add conversation history if we have a conversationID
		options["history"] = []map[string]interface{}{}
	}
	if filters != nil {
		// Add parent or other filter options
		for k, v := range filters {
			options[k] = v
		}
	}

	payload := map[string]interface{}{
		"text": message,
	}

	if len(options) > 0 {
		payload["options"] = options
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.ServerURL+"/agents/rag/-/chat", bytes.NewBuffer(jsonData))
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
		return "", NewHttpErrorWithRequestBody(resp, req, string(jsonData))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Extract the response message
	if response, ok := result["response"]; ok {
		if responseStr, ok := response.(string); ok {
			return responseStr, nil
		}
	}

	// Fallback to return the whole response as JSON string
	responseJson, _ := json.MarshalIndent(result, "", "  ")
	return string(responseJson), nil
}

func (c *client) CopyNode(uuid, parent, title string) (*Node, error) {
	payload := map[string]string{
		"parent": parent,
		"title":  title,
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

func (c *client) RunFeatureAsAction(uuid string, uuids []string) (map[string]interface{}, error) {
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *client) RunFeatureAsExtension(uuid string, params map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/features/"+uuid+"/-/run-ext", bytes.NewBuffer(jsonData))
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

func (c *client) RunAction(uuid string, request ActionRunRequest) (map[string]interface{}, error) {
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

	var result map[string]interface{}
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

func (c *client) RunExtension(uuid string, data map[string]interface{}) (interface{}, error) {
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

	var result interface{}
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

func (c *client) RunAITool(uuid string, params map[string]interface{}) (map[string]interface{}, error) {
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

	var result map[string]interface{}
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

func (c *client) CreateAgent(agent AgentCreate) (*Agent, error) {
	jsonData, err := json.Marshal(agent)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/agents", bytes.NewBuffer(jsonData))
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

	var result Agent
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
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

func (c *client) CreateAspect(aspect AspectCreate) (*Aspect, error) {
	jsonData, err := json.Marshal(aspect)
	if err != nil {
		return nil, err
	}

	requestBodyStr := string(jsonData)

	req, err := http.NewRequest("POST", c.ServerURL+"/aspects", bytes.NewBuffer(jsonData))
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

	var result Aspect
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
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

func (c *client) ExportAspect(uuid string, format string) (interface{}, error) {
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

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
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
