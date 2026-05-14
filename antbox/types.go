package antbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ChatMessageRole represents the role in a chat message
type ChatMessageRole string

const (
	ChatMessageRoleUser     ChatMessageRole = "user"
	ChatMessageRoleModel    ChatMessageRole = "model"
	ChatMessageRoleSystem   ChatMessageRole = "system"
	ChatMessageRoleFunction ChatMessageRole = "function"
	ChatMessageRoleTool     ChatMessageRole = "tool"
)

// ToolCall represents a tool call in a chat message
type ToolCall struct {
	ID   string         `json:"id,omitempty"`
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

// ToolResponse represents a tool response in a chat message
type ToolResponse struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Text string `json:"text"`
}

// ChatMessagePart represents a part of a chat message
type ChatMessagePart struct {
	Text         *string       `json:"text,omitempty"`
	ToolCall     *ToolCall     `json:"toolCall,omitempty"`
	ToolResponse *ToolResponse `json:"toolResponse,omitempty"`
}

// ChatMessage represents a message in a chat history
type ChatMessage struct {
	Role  ChatMessageRole   `json:"role"`
	Parts []ChatMessagePart `json:"parts"`
}

// ChatHistory represents a conversation history
type ChatHistory []ChatMessage

// FilterOperator represents the operators available for node filtering
type FilterOperator string

const (
	FilterOperatorEqual        FilterOperator = "=="
	FilterOperatorLessEqual    FilterOperator = "<="
	FilterOperatorGreaterEqual FilterOperator = ">="
	FilterOperatorLess         FilterOperator = "<"
	FilterOperatorGreater      FilterOperator = ">"
	FilterOperatorNotEqual     FilterOperator = "!="
	FilterOperatorMatch        FilterOperator = "match"
	FilterOperatorIn           FilterOperator = "in"
	FilterOperatorNotIn        FilterOperator = "not-in"
	FilterOperatorContains     FilterOperator = "contains"
	FilterOperatorContainsAll  FilterOperator = "contains-all"
	FilterOperatorContainsAny  FilterOperator = "contains-any"
	FilterOperatorNotContains  FilterOperator = "not-contains"
	FilterOperatorContainsNone FilterOperator = "contains-none"
)

// NodeFilter represents a single filter with field, operator, and value
type NodeFilter [3]interface{} // [field string, operator FilterOperator, value interface{}]

// NodeFilters1D represents a 1D array of node filters (AND logic)
type NodeFilters1D []NodeFilter

// NodeFilters2D represents a 2D array of node filters (OR of ANDs)
type NodeFilters2D []NodeFilters1D

// NodeFilters represents either 1D or 2D node filters
type NodeFilters interface{}

// NewNodeFilter creates a new NodeFilter with the given field, operator, and value
func NewNodeFilter(field string, operator FilterOperator, value interface{}) NodeFilter {
	return NodeFilter{field, operator, value}
}

// NewNodeFilters1D creates a new 1D NodeFilters from a slice of NodeFilter
func NewNodeFilters1D(filters ...NodeFilter) NodeFilters1D {
	return NodeFilters1D(filters)
}

// NewNodeFilters2D creates a new 2D NodeFilters from a slice of NodeFilters1D
func NewNodeFilters2D(filterGroups ...NodeFilters1D) NodeFilters2D {
	return NodeFilters2D(filterGroups)
}

// Equal creates a filter for equality (==)
func Equal(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorEqual, value)
}

// NotEqual creates a filter for inequality (!=)
func NotEqual(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorNotEqual, value)
}

// Match creates a filter for pattern matching.
func Match(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorMatch, value)
}

// Contains creates a filter for contains check
func Contains(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorContains, value)
}

// In creates a filter for "in" operator
func In(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorIn, value)
}

// GreaterThan creates a filter for greater than (>)
func GreaterThan(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorGreater, value)
}

// LessThan creates a filter for less than (<)
func LessThan(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorLess, value)
}

// GreaterEqual creates a filter for greater than or equal (>=)
func GreaterEqual(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorGreaterEqual, value)
}

// LessEqual creates a filter for less than or equal (<=)
func LessEqual(field string, value interface{}) NodeFilter {
	return NewNodeFilter(field, FilterOperatorLessEqual, value)
}

type Node struct {
	UUID                   string         `json:"uuid,omitempty"`
	Fid                    string         `json:"fid,omitempty"`
	Title                  string         `json:"title,omitempty"`
	Name                   string         `json:"name,omitempty"`
	Description            string         `json:"description,omitempty"`
	Mimetype               string         `json:"mimetype,omitempty"`
	Parent                 string         `json:"parent,omitempty"`
	Owner                  string         `json:"owner,omitempty"`
	Group                  string         `json:"group,omitempty"`
	Groups                 []string       `json:"groups,omitempty"`
	Permissions            *Permissions   `json:"permissions,omitempty"`
	Size                   int            `json:"size,omitempty"`
	CreatedAt              string         `json:"createdTime,omitempty"`
	ModifiedAt             string         `json:"modifiedTime,omitempty"`
	Fulltext               string         `json:"fulltext,omitempty"`
	Locked                 bool           `json:"locked,omitempty"`
	LockedBy               string         `json:"lockedBy,omitempty"`
	UnlockAuthorizedGroups []string       `json:"unlockAuthorizedGroups,omitempty"`
	Aspects                []string       `json:"aspects,omitempty"`
	Properties             map[string]any `json:"properties,omitempty"`
	Tags                   []string       `json:"tags,omitempty"`
	Related                []string       `json:"related,omitempty"`
	Filters                NodeFilters    `json:"filters,omitempty"`
	Email                  string         `json:"email,omitempty"`
	Phone                  string         `json:"phone,omitempty"`
	HasWhatsapp            bool           `json:"hasWhatsapp,omitempty"`
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

	// Show decimal if integer part < 10, otherwise truncate
	if int(size) < 10 && unitIndex > 0 {
		return fmt.Sprintf("%.1f%s", size, units[unitIndex])
	}
	return fmt.Sprintf("%d%s", int(size), units[unitIndex])
}

type Permission string

const (
	PermissionRead   Permission = "Read"
	PermissionWrite  Permission = "Write"
	PermissionExport Permission = "Export"
)

type Permissions struct {
	Group         []Permission            `json:"group,omitempty"`
	Authenticated []Permission            `json:"authenticated,omitempty"`
	Anonymous     []Permission            `json:"anonymous,omitempty"`
	Advanced      map[string][]Permission `json:"advanced,omitempty"`
}

// NodeCreate represents the request to create a node
type NodeCreate struct {
	Title       string       `json:"title"`
	Mimetype    string       `json:"mimetype"`
	Parent      string       `json:"parent,omitempty"`
	Content     string       `json:"content,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

// NodeUpdate represents the request to update a node
type NodeUpdate struct {
	Title       string       `json:"title,omitempty"`
	Mimetype    string       `json:"mimetype,omitempty"`
	Parent      string       `json:"parent,omitempty"`
	Content     string       `json:"content,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

type NodeFilterResult struct {
	Nodes     []Node `json:"nodes"`
	PageSize  int    `json:"pageSize"`
	PageToken int    `json:"pageToken"`
}

// Agent represents an AI agent.
type Agent struct {
	UUID           string `json:"uuid,omitempty"`
	Name           string `json:"name,omitempty"`
	Title          string `json:"title,omitempty"` // legacy field accepted for older servers/tests
	Description    string `json:"description,omitempty"`
	ExposedToUsers bool   `json:"exposedToUsers,omitempty"`
	Model          string `json:"model,omitempty"`
	Tools          any    `json:"tools,omitempty"`
	SystemPrompt   string `json:"systemPrompt,omitempty"`
	MaxLlmCalls    int    `json:"maxLlmCalls,omitempty"`
	CreatedAt      string `json:"createdTime,omitempty"`
	ModifiedAt     string `json:"modifiedTime,omitempty"`
}

func (a Agent) DisplayName() string {
	if a.Name != "" {
		return a.Name
	}
	return a.Title
}

// CreateAgentRequest represents the request to create or replace an agent.
type CreateAgentRequest struct {
	UUID           string `json:"uuid,omitempty"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	ExposedToUsers *bool  `json:"exposedToUsers,omitempty"`
	Model          string `json:"model,omitempty"`
	Tools          any    `json:"tools,omitempty"`
	SystemPrompt   string `json:"systemPrompt,omitempty"`
	MaxLlmCalls    int    `json:"maxLlmCalls,omitempty"`
}

type AgentCreate = CreateAgentRequest

type ChatHistoryEntry struct {
	Role      string `json:"role,omitempty"`
	Text      string `json:"text,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

type ChatOptions struct {
	History      []ChatHistoryEntry `json:"history,omitempty"`
	Files        []string           `json:"files,omitempty"`
	Temperature  *float64           `json:"temperature,omitempty"`
	MaxTokens    *int               `json:"maxTokens,omitempty"`
	Instructions string             `json:"instructions,omitempty"`
}

type AnswerOptions struct {
	Files        []string `json:"files,omitempty"`
	Temperature  *float64 `json:"temperature,omitempty"`
	MaxTokens    *int     `json:"maxTokens,omitempty"`
	Instructions string   `json:"instructions,omitempty"`
}

// AgentChatRequest represents a chat request to an agent.
type AgentChatRequest struct {
	Text    string       `json:"text"`
	Options *ChatOptions `json:"options,omitempty"`
}

// AgentAnswerRequest represents an answer request to an agent.
type AgentAnswerRequest struct {
	Text    string         `json:"text"`
	Options *AnswerOptions `json:"options,omitempty"`
}

// APIKey represents an API key.
type APIKey struct {
	UUID        string `json:"uuid,omitempty"`
	Title       string `json:"title,omitempty"`
	Secret      string `json:"secret,omitempty"`
	Group       string `json:"group,omitempty"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active,omitempty"`
	CreatedAt   string `json:"createdTime,omitempty"`
	Owner       string `json:"owner,omitempty"` // legacy field accepted for older servers/tests
}

// APIKeyCreate represents the request to create an API key.
type APIKeyCreate struct {
	Group       string `json:"group"`
	Description string `json:"description,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

type CreateAPIKeyRequest = APIKeyCreate

// User represents a user account.
type User struct {
	UUID        string   `json:"uuid,omitempty"` // legacy field accepted for older servers/tests
	Email       string   `json:"email,omitempty"`
	Title       string   `json:"title,omitempty"`
	Name        string   `json:"name,omitempty"` // returned by /login/me
	Group       string   `json:"group,omitempty"`
	Groups      []string `json:"groups,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	HasWhatsapp bool     `json:"hasWhatsapp,omitempty"`
	Active      bool     `json:"active,omitempty"`
	CreatedAt   string   `json:"createdTime,omitempty"`
	ModifiedAt  string   `json:"modifiedTime,omitempty"`
}

func (u User) DisplayName() string {
	if u.Title != "" {
		return u.Title
	}
	return u.Name
}

// UserCreate represents the request to create a user.
type UserCreate struct {
	Email       string   `json:"email"`
	Title       string   `json:"title"`
	Name        string   `json:"name,omitempty"` // legacy alias accepted by older callers
	Group       string   `json:"group"`
	Groups      []string `json:"groups,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	HasWhatsapp bool     `json:"hasWhatsapp,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

type CreateUserRequest = UserCreate

// UserUpdate represents the request to update a user.
type UserUpdate struct {
	Title       string   `json:"title,omitempty"`
	Name        string   `json:"name,omitempty"` // legacy alias accepted by older callers
	Group       string   `json:"group,omitempty"`
	Groups      []string `json:"groups,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	HasWhatsapp bool     `json:"hasWhatsapp,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

type UpdateUserRequest = UserUpdate

// Group represents a group.
type Group struct {
	UUID        string `json:"uuid,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdTime,omitempty"`
}

// GroupCreate represents the request to create a group.
type GroupCreate struct {
	UUID        string `json:"uuid,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type CreateGroupRequest = GroupCreate

type Breadcrumb struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
}

// DocInfo represents documentation information
type DocInfo struct {
	UUID        string `json:"uuid,omitempty"`
	Description string `json:"description,omitempty"`
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

		var jsonData any
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
