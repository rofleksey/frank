package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type HTTPRequestCommand struct {
}

func NewHTTPRequestCommand() *HTTPRequestCommand {
	return &HTTPRequestCommand{}
}

type HTTPRequestCommandData struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    *string           `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"` // in seconds
}

type HTTPRequestResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func (c *HTTPRequestCommand) Execute(ctx context.Context, prompt dto.Prompt) (string, error) {
	logger := slog.With(
		slog.String("command", c.Name()),
		slog.String("prompt_id", prompt.ID.String()),
	)

	logger.InfoContext(ctx, "Executing http_request command",
		slog.Any("prompt", prompt),
	)

	var requestData HTTPRequestCommandData
	if err := json.Unmarshal([]byte(prompt.Text), &requestData); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	if requestData.URL == "" {
		return "", fmt.Errorf("empty URL")
	}

	if requestData.Method == "" {
		requestData.Method = http.MethodGet
	}

	logger.DebugContext(ctx, "Creating HTTP client with timeout")

	client := &http.Client{
		Timeout: time.Duration(requestData.Timeout) * time.Second,
	}
	defer client.CloseIdleConnections()

	var bodyReader io.Reader
	if requestData.Body != nil {
		bodyReader = bytes.NewReader([]byte(*requestData.Body))
		logger.DebugContext(ctx, "Request body included",
			slog.Int("body_length", len(*requestData.Body)),
		)
	} else {
		logger.DebugContext(ctx, "No request body")
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(requestData.Method), requestData.URL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	for key, value := range requestData.Headers {
		req.Header.Set(key, value)
	}

	if len(requestData.Headers) > 0 {
		logger.DebugContext(ctx, "Request headers set",
			slog.Int("header_count", len(requestData.Headers)),
		)
	}

	if requestData.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "text/plain")
		logger.DebugContext(ctx, "Set default Content-Type header")
	}

	startTime := time.Now()
	logger.InfoContext(ctx, "Sending HTTP request")

	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	logger.InfoContext(ctx, "HTTP request completed",
		slog.Int("status_code", resp.StatusCode),
		slog.Duration("duration", duration),
	)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	result := HTTPRequestResult{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(bodyBytes),
	}

	logger.DebugContext(ctx, "Response details",
		slog.Int("body_length", len(result.Body)),
		slog.Int("header_count", len(headers)),
	)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}

	logger.InfoContext(ctx, "HTTP request command completed successfully",
		slog.Int("result_length", len(resultJSON)),
	)

	return string(resultJSON), nil
}

func (c *HTTPRequestCommand) Name() string {
	return "http_request"
}

func (c *HTTPRequestCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
      - url
    properties:
      command:
        type: string
        enum: 
          - http_request
      url:
        type: string
        description: The URL to send the request to
      method:
        type: string
        enum: [GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS]
        default: GET
        description: HTTP method to use
      headers:
        type: object
        additionalProperties:
          type: string
        description: HTTP headers to include in the request
      body:
        type: string
        nullable: true
        description: Request body content. Set to null to omit body.
      timeout:
        type: integer
        minimum: 1
        description: Request timeout in seconds
    description: Executes an HTTP request and returns the response with status code, headers, and body

    RESULT SPEC:

    type: object
    properties:
      status_code:
        type: integer
        description: HTTP status code
        example: 200
      headers:
        type: object
        description: HTTP response headers
        additionalProperties:
          type: string
        example:
          Content-Type: application/json
          Cache-Control: no-cache
      body:
        type: string
        description: HTTP response body
        example: '{"message": "Success"}'
    required:
      - status_code
      - headers
      - body
  `)
}
