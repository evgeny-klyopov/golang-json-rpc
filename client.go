package jsonRpc

import (
	"bytes"
	"fmt"
	"github.com/gorilla/rpc/v2/json2"
	"net/http"
	"time"
)

type ClientCredential struct {
	Protocol string
	User     string
	Password string
	Host     string
	Port     string
}

type client struct {
	credential ClientCredential
	url        string
	timeout    time.Duration
}

type JsonRpcConnector interface {
	SetTimeout(timeout time.Duration) JsonRpcConnector
	Request(method string, params interface{}, data interface{}) (*int, error)
}

func NewClient(credential ClientCredential, timeout time.Duration, prefixUrl string) JsonRpcConnector {
	url := fmt.Sprintf("%s://", credential.Protocol)

	if credential.User != "" && credential.Password != "" {
		url += fmt.Sprintf(
			"%s:%s@",
			credential.User,
			credential.Password,
		)
	}
	url += fmt.Sprintf(
		"%s:%s%s",
		credential.Host,
		credential.Port,
		prefixUrl,
	)

	return &client{
		credential: credential,
		url:        url,
		timeout:    timeout,
	}
}

func (c *client) SetTimeout(timeout time.Duration) JsonRpcConnector {
	c.timeout = timeout

	return c
}
func (c *client) Request(method string, params interface{}, data interface{}) (*int, error) {
	message, err := json2.EncodeClientRequest(method, params)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", c.url, bytes.NewBuffer(message))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{
		Timeout: c.timeout,
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	err = json2.DecodeClientResponse(response.Body, &data)
	if err != nil {
		return &response.StatusCode, err
	}

	return &response.StatusCode, nil
}
