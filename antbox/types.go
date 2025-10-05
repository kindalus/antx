package antbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Node struct {
	UUID        string       `json:"uuid,omitempty"`
	Fid         string       `json:"fid,omitempty"`
	Title       string       `json:"title,omitempty"`
	Mimetype    string       `json:"mimetype,omitempty"`
	Parent      string       `json:"parent,omitempty"`
	Owner       string       `json:"owner,omitempty"`
	Group       string       `json:"group,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
	Size        int          `json:"size,omitempty"`
	CreatedAt   string       `json:"createdAt,omitempty"`
	ModifiedAt  string       `json:"modifiedAt,omitempty"`
}

// HumanReadableSize returns a human-readable representation of the node's size
func (n *Node) HumanReadableSize() string {
	if n.Size == 0 {
		return "0B"
	}

	const unit = 1024
	units := []string{"B", "K", "M", "G", "T", "P"}

	size := float64(n.Size)
	unitIndex := 0

	for size >= unit && unitIndex < len(units)-1 {
		size /= unit
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%.0f%s", size, units[unitIndex])
	}

	return fmt.Sprintf("%.1f%s", size, units[unitIndex])
}

type Permissions struct {
	Group         []string       `json:"group,omitempty"`
	Authenticated []string       `json:"authenticated,omitempty"`
	Anonymous     []string       `json:"anonymous,omitempty"`
	Advanced      map[string]any `json:"advanced,omitempty"`
}

type NodeFilterResult struct {
	Nodes     []Node `json:"nodes"`
	PageSize  int    `json:"pageSize"`
	PageToken int    `json:"pageToken"`
}

type HttpError struct {
	StatusCode      int
	Status          string
	Body            string
	URL             string
	Method          string
	RequestHeaders  http.Header
	RequestBody     string
	ResponseHeaders http.Header
}

func (e *HttpError) Error() string {
	var result strings.Builder

	// Error header
	result.WriteString(fmt.Sprintf("Error: %s %s - %d\n\n", e.Method, e.URL, e.StatusCode))

	// Request section
	result.WriteString("==> Request\n")
	e.writeHeaders(&result, e.RequestHeaders)
	if e.RequestBody != "" {
		result.WriteString("Body: ")
		result.WriteString(e.formatJSON(e.RequestBody))
		result.WriteString("\n")
	}
	result.WriteString("\n")

	// Response section
	result.WriteString("Response <==\n")
	e.writeHeaders(&result, e.ResponseHeaders)
	if e.Body != "" {
		result.WriteString("Body: ")
		result.WriteString(e.formatJSON(e.Body))
		result.WriteString("\n")
	}

	return result.String()
}

func (e *HttpError) writeHeaders(result *strings.Builder, headers http.Header) {
	for name, values := range headers {
		for _, value := range values {
			fmt.Fprintf(result, "%s: %s\n", name, value)
		}
	}
}

func (e *HttpError) formatJSON(body string) string {
	if body == "" {
		return ""
	}

	// Try to detect and pretty print JSON
	trimmed := strings.TrimSpace(body)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {

		var jsonData interface{}
		if err := json.Unmarshal([]byte(trimmed), &jsonData); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				return string(prettyJSON)
			}
		}
	}

	// Return original body if not JSON or if pretty printing fails
	return body
}

func readResponseBody(resp *http.Response) string {
	if resp.Body == nil {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func NewHttpError(resp *http.Response, method, url string) *HttpError {
	bodyStr := readResponseBody(resp)

	return &HttpError{
		StatusCode:      resp.StatusCode,
		Status:          resp.Status,
		Body:            bodyStr,
		URL:             url,
		Method:          method,
		RequestHeaders:  make(http.Header),
		RequestBody:     "",
		ResponseHeaders: resp.Header,
	}
}

func NewHttpErrorWithRequest(resp *http.Response, req *http.Request) *HttpError {
	bodyStr := readResponseBody(resp)
	requestBodyStr := ""

	// Try to read request body if available
	if req.Body != nil {
		if bodyBytes, err := io.ReadAll(req.Body); err == nil {
			requestBodyStr = string(bodyBytes)
		}
	}

	return &HttpError{
		StatusCode:      resp.StatusCode,
		Status:          resp.Status,
		Body:            bodyStr,
		URL:             req.URL.String(),
		Method:          req.Method,
		RequestHeaders:  req.Header,
		RequestBody:     requestBodyStr,
		ResponseHeaders: resp.Header,
	}
}

func NewHttpErrorWithRequestBody(resp *http.Response, req *http.Request, requestBody string) *HttpError {
	bodyStr := readResponseBody(resp)

	return &HttpError{
		StatusCode:      resp.StatusCode,
		Status:          resp.Status,
		Body:            bodyStr,
		URL:             req.URL.String(),
		Method:          req.Method,
		RequestHeaders:  req.Header,
		RequestBody:     requestBody,
		ResponseHeaders: resp.Header,
	}
}
