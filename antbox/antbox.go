package antbox

import "net/http"

type Antbox interface {
	Login() error
	GetNode(uuid string) (*Node, error)
	ListNodes(parent string) ([]Node, error)
	SetAuthHeader(req *http.Request)
	CreateFolder(parent, name string) (*Node, error)
	RemoveNode(uuid string) error
	MoveNode(uuid, newParent string) error
	ChangeNodeName(uuid, newName string) error
	CreateFile(filePath, parentUuid string) (*Node, error)
	UpdateFile(uuid, filePath string) (*Node, error)
	FindNodes(filters any, pageSize, pageToken int) (*NodeFilterResult, error)
	EvaluateNode(uuid string) ([]Node, error)
	DownloadNode(uuid, downloadPath string) error
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
