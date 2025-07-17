package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseHandler interface {
	Handle(*http.Response, interface{}) error
}

type DirectResponse struct{}

type CodeWrapperResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// DirectResponseHandler handle response directly
type DirectResponseHandler struct{}

func (d *DirectResponseHandler) Handle(resp *http.Response, result interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// CodeWrapperResponseHandler handle response with code wrapper
type CodeWrapperResponseHandler struct{}

func (c *CodeWrapperResponseHandler) Handle(resp *http.Response, result interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var codeWrapper CodeWrapperResponse
	codeWrapper.Data = result
	if err := json.NewDecoder(resp.Body).Decode(&codeWrapper); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if codeWrapper.Code != 0 {
		return fmt.Errorf("unexpected code: %d, msg: %s", codeWrapper.Code, codeWrapper.Msg)
	}

	return nil
}
