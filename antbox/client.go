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
			fmt.Printf("==> Request:\n%s\n", string(dump))
		}
	}

	resp, err := c.client.Do(req)

	if c.debug && resp != nil {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println("Error dumping response:", err)
		} else {
			fmt.Printf("<== Response:\n%s\n", string(dump))
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
