package httpclient

import "net/http"

// AuthProvider 定义认证接口
type AuthProvider interface {
	Apply(req *http.Request)
}

type AuthNone struct{}

func (a *AuthNone) Apply(req *http.Request) {}

// AuthBearerToken OAuth2 Bearer Token
type AuthBearerToken struct {
	Token string
}

func (a *AuthBearerToken) Apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.Token)
}

// AuthAPIKey API Key 认证
type AuthAPIKey struct {
	Key  string
	In   string // "header" or "query"
	Name string // 键名，如 "X-API-Key" 或 "api_key"
}

func (a *AuthAPIKey) Apply(req *http.Request) {
	switch a.In {
	case "header":
		req.Header.Set(a.Name, a.Key)
	case "query":
		q := req.URL.Query()
		q.Add(a.Name, a.Key)
		req.URL.RawQuery = q.Encode()
	}
}
