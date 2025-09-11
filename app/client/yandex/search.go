package yandex

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const webSearchURL = "https://searchapi.api.cloud.yandex.net/v2/web/search"

type WebSearchRequest struct {
	Query          WebSearchQuery      `json:"query"`
	SortSpec       *WebSearchSortSpec  `json:"sortSpec,omitempty"`
	GroupSpec      *WebSearchGroupSpec `json:"groupSpec,omitempty"`
	MaxPassages    string              `json:"maxPassages,omitempty"`
	Region         string              `json:"region,omitempty"`
	L10n           string              `json:"l10n,omitempty"`
	FolderId       string              `json:"folderId,omitempty"`
	ResponseFormat string              `json:"responseFormat,omitempty"`
	UserAgent      string              `json:"userAgent,omitempty"`
}

type WebSearchQuery struct {
	SearchType  string `json:"searchType"`
	QueryText   string `json:"queryText"`
	FamilyMode  string `json:"familyMode,omitempty"`
	Page        string `json:"page,omitempty"`
	FixTypoMode string `json:"fixTypoMode,omitempty"`
}

type WebSearchSortSpec struct {
	SortMode  string `json:"sortMode"`
	SortOrder string `json:"sortOrder"`
}

type WebSearchGroupSpec struct {
	GroupMode    string `json:"groupMode"`
	GroupsOnPage string `json:"groupsOnPage"`
	DocsInGroup  string `json:"docsInGroup"`
}

type WebSearchResponse struct {
	RawData string `json:"rawData"`
}

func (c *Client) WebSearch(ctx context.Context, query string) (string, error) {
	req := &WebSearchRequest{
		Query: WebSearchQuery{
			SearchType: "SEARCH_TYPE_COM",
			QueryText:  query,
			FamilyMode: "FAMILY_MODE_NONE",
		},
		ResponseFormat: "FORMAT_XML",
		FolderId:       c.cfg.Yandex.FolderID,
	}

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	iamToken, err := c.getIAMToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get IAM token: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		webSearchURL,
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+iamToken)
	httpReq.Header.Set("X-Folder-ID", c.cfg.Yandex.FolderID)

	res, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	bytez, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got response with status %d: %s", res.StatusCode, string(bytez))
	}

	var webSearchRes WebSearchResponse
	if err = json.Unmarshal(bytez, &webSearchRes); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(webSearchRes.RawData)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return string(decodedBytes), nil
}
