package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

type RequestParams struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Ctx     context.Context
}

func NewHttpRequest() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   6 * time.Second,
			KeepAlive: 16 * time.Second,
		}).DialContext,

		MaxIdleConns:        256,
		MaxIdleConnsPerHost: 256,
		MaxConnsPerHost:     256,
		IdleConnTimeout:     30 * time.Second,

		TLSHandshakeTimeout:   2 * time.Second,
		ExpectContinueTimeout: 500 * time.Millisecond,
		ForceAttemptHTTP2:     false,
		DisableCompression:    true,
		DisableKeepAlives:     false,
	}

	return &http.Client{
		Timeout:   6 * time.Second,
		Transport: transport,
	}
}

func Do[T any](client *http.Client, params RequestParams, out *T) (*http.Response, error) {
	if params.Ctx == nil {
		params.Ctx = context.Background()
	}

	var bodyReader io.Reader
	if len(params.Body) > 0 {
		bodyReader = bytes.NewBuffer(params.Body)
	}

	req, err := http.NewRequestWithContext(params.Ctx, params.Method, params.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range params.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return resp, errors.New("erro HTTP: " + resp.Status + " - " + string(body))
	}

	if out == nil {
		return resp, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	if len(body) == 0 {
		return resp, nil
	}

	if err := json.Unmarshal(body, out); err != nil {
		return resp, err
	}

	return resp, nil
}
