package antbox

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
)

type client struct {
	ServerURL string
	APIKey    string
	Root      string
	JWT       string
	client    *http.Client
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

	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.client.Do(req)
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

	resp, err := c.client.Do(req)
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

	resp, err := c.client.Do(req)
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

	resp, err := c.client.Do(req)
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

func (c *client) SetAuthHeader(req *http.Request) {
	if c.JWT != "" {
		req.Header.Set("Authorization", "Bearer "+c.JWT)
	} else if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
}
