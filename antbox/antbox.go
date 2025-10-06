package antbox

import "net/http"

type ChatResponse struct {
	Text    string
	History []map[string]any
}

type Antbox interface {
	// Authentication
	Login() error
	SetAuthHeader(req *http.Request)

	// Node operations
	GetNode(uuid string) (*Node, error)
	ListNodes(parent string) ([]Node, error)
	CreateFolder(parent, name string) (*Node, error)
	CreateSmartFolder(parent, name string, filters any) (*Node, error)
	RemoveNode(uuid string) error
	MoveNode(uuid, newParent string) error
	ChangeNodeName(uuid, newName string) error
	CreateFile(filePath string, metadata map[string]any) (*Node, error)
	UpdateFile(uuid, filePath string) (*Node, error)
	FindNodes(filters any, pageSize, pageToken int) (*NodeFilterResult, error)
	EvaluateNode(uuid string) ([]Node, error)
	DownloadNode(uuid, downloadPath string) error
	GetBreadcrumbs(uuid string) ([]Node, error)
	CopyNode(uuid, parent, title string) (*Node, error)
	DuplicateNode(uuid string) (*Node, error)
	ExportNode(uuid string, format string) ([]byte, error)

	// Feature operations
	ListFeatures() ([]Feature, error)
	GetFeature(uuid string) (*Feature, error)
	DeleteFeature(uuid string) error
	ExportFeature(uuid string, exportType string) (string, error)
	ListActionFeatures() ([]Feature, error)
	ListExtensionFeatures() ([]Feature, error)
	RunFeatureAsAction(uuid string, uuids []string) (map[string]any, error)
	RunFeatureAsExtension(uuid string, params map[string]any) (string, error)

	// Action operations
	ListActions() ([]Feature, error)
	RunAction(uuid string, request ActionRunRequest) (map[string]any, error)

	// Extension operations
	ListExtensions() ([]Feature, error)
	RunExtension(uuid string, data map[string]any) (any, error)

	// AI Tool operations
	ListAITools() ([]Feature, error)
	RunAITool(uuid string, params map[string]any) (map[string]any, error)

	// Agent operations
	ListAgents() ([]Agent, error)
	CreateAgent(agent AgentCreate) (*Agent, error)
	GetAgent(uuid string) (*Agent, error)
	DeleteAgent(uuid string) error
	ChatWithAgent(agentUUID string, message string, conversationID string, temperature *float64, maxTokens *int, history []map[string]any) (string, error)
	AnswerFromAgent(agentUUID string, query string, temperature *float64, maxTokens *int) (string, error)
	RagChat(message string, options map[string]any) (ChatResponse, error)

	// API Key operations
	ListAPIKeys() ([]APIKey, error)
	CreateAPIKey(request APIKeyCreate) (*APIKey, error)
	GetAPIKey(uuid string) (*APIKey, error)
	DeleteAPIKey(uuid string) error

	// User operations
	ListUsers() ([]User, error)
	CreateUser(user UserCreate) (*User, error)
	GetUser(email string) (*User, error)
	UpdateUser(email string, user UserUpdate) (*User, error)
	DeleteUser(uuid string) error

	// Group operations
	ListGroups() ([]Group, error)
	CreateGroup(group GroupCreate) (*Group, error)
	GetGroup(uuid string) (*Group, error)
	UpdateGroup(uuid string, group GroupUpdate) (*Group, error)
	DeleteGroup(uuid string) error

	// Template operations
	GetTemplate(uuid string) ([]byte, error)

	// Aspect operations
	ListAspects() ([]Aspect, error)
	CreateAspect(aspect AspectCreate) (*Aspect, error)
	GetAspect(uuid string) (*Aspect, error)
	DeleteAspect(uuid string) error
	ExportAspect(uuid string, format string) (any, error)
}

func NewClient(serverURL, apiKey, root, jwt string, debug bool) Antbox {
	return &client{
		ServerURL: serverURL,
		APIKey:    apiKey,
		Root:      root,
		JWT:       jwt,
		client:    &http.Client{},
		debug:     debug,
	}
}
