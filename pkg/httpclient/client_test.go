package httpclient

import (
	"context"
	"testing"
	"time"
)

type headersResponse struct {
	Headers map[string][]string `json:"headers"`
}

func TestHTTPClient(t *testing.T) {
	client := NewHTTPClient(&Config{
		Timeout:  time.Second * 5,
		Response: &DirectResponseHandler{},
	})

	var resp headersResponse
	err := client.Get(context.Background(), "https://httpbin.dev/headers", nil, &resp)
	if err != nil {
		t.Fatal("Request to httpbin failed. ", err)
	}

	if resp.Headers["Host"][0] != "httpbin.dev" {
		t.Fatal("Host header not match. ", resp.Headers["Host"])
	}
}

func TestAuthWithBearerToken(t *testing.T) {
	token := "sk-aksdimu93i33323"
	client := NewHTTPClient(&Config{
		Timeout:  time.Second * 5,
		Response: &DirectResponseHandler{},
		Auth: &AuthBearerToken{
			Token: token,
		},
	})

	var resp headersResponse
	err := client.Get(context.Background(), "https://httpbin.dev/headers", nil, &resp)
	if err != nil {
		t.Fatal("Request to httpbin failed. ", err)
	}

	if resp.Headers["Authorization"][0] != "Bearer "+token {
		t.Fatal("Authorization header not match. ", resp.Headers["Authorization"])
	}
}

func TestGetWithQuery(t *testing.T) {
	client := NewHTTPClient(&Config{
		Timeout:  time.Second * 5,
		Response: &DirectResponseHandler{},
	})
	q := struct {
		Name string `url:"name"`
		Sex  string `url:"sex"`
	}{
		Name: "abcd",
		Sex:  "male",
	}

	var resp headersResponse
	err := client.Get(context.Background(), "https://httpbin.dev/response-headers", &q, &resp)
	if err != nil {
		t.Fatal("Request to httpbin failed. ", err)
	}
}

func TestAuthWithAPIKey(t *testing.T) {
	apikey := "sk-aksdimu93i33323"
	client := NewHTTPClient(&Config{
		Timeout:  time.Second * 5,
		Response: &DirectResponseHandler{},
		Auth: &AuthAPIKey{
			Key:  apikey,
			Name: "token",
			In:   "header",
		},
	})

	var resp headersResponse
	err := client.Get(context.Background(), "https://httpbin.dev/headers", nil, &resp)
	if err != nil {
		t.Fatal("Request to httpbin failed. ", err)
	}

	if resp.Headers["Token"][0] != apikey {
		t.Fatal("APIKey header not match. ", resp.Headers["Token"])
	}
}
