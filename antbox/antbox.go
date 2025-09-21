package antbox

import "net/http"

type Antbox interface {
	Login() error
	GetNode(uuid string) (*Node, error)
	ListNodes(parent string) ([]Node, error)
	SetAuthHeader(req *http.Request)
	CreateFolder(parent, name string) (*Node, error)
}

func NewClient(serverURL, apiKey, root, jwt string) Antbox {
	return &client{
		ServerURL: serverURL,
		APIKey:    apiKey,
		Root:      root,
		JWT:       jwt,
		client:    &http.Client{},
	}
}
