package truenas

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type NFSShare struct {
	ID           int      `json:"id,omitempty"`
	Path         string   `json:"path"`
	Comment      string   `json:"comment"`
	Networks     []string `json:"networks,omitempty"`
	Hosts        []string `json:"hosts,omitempty"`
	MapRootUser  string   `json:"maproot_user,omitempty"`
	MapRootGroup string   `json:"maproot_group,omitempty"`
	ReadOnly     bool     `json:"ro"`
	Enabled      bool     `json:"enabled"`
}

type Dataset struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Pool          string                 `json:"pool"`
	Type          string                 `json:"type"`
	Used          map[string]interface{} `json:"used,omitempty"`
	Available     map[string]interface{} `json:"available,omitempty"`
	Mountpoint    string                 `json:"mountpoint,omitempty"`
	Compression   map[string]interface{} `json:"compression,omitempty"`
	Deduplication map[string]interface{} `json:"deduplication,omitempty"`
}

func NewClient() (*Client, error) {
	baseURL := os.Getenv("TRUENAS_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("TRUENAS_URL environment variable not set")
	}

	apiKey := os.Getenv("TRUENAS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TRUENAS_API_KEY environment variable not set")
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}, nil
}

func NewClientWithParams(server, token string) (*Client, error) {
	if server == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	if token == "" {
		return nil, fmt.Errorf("API token is required")
	}

	return &Client{
		baseURL: server,
		apiKey:  token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}, nil
}

func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s/api/v2.0%s", c.baseURL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) CreateNFSShare(share *NFSShare) (int, error) {
	share.Enabled = true

	respBody, err := c.doRequest("POST", "/sharing/nfs", share)
	if err != nil {
		return 0, err
	}

	var result NFSShare
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.ID, nil
}

func (c *Client) ListNFSShares() ([]NFSShare, error) {
	respBody, err := c.doRequest("GET", "/sharing/nfs", nil)
	if err != nil {
		return nil, err
	}

	var shares []NFSShare
	if err := json.Unmarshal(respBody, &shares); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return shares, nil
}

func (c *Client) DeleteNFSShare(id int) error {
	endpoint := fmt.Sprintf("/sharing/nfs/id/%d", id)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

func (c *Client) ListDatasets(poolName string) ([]Dataset, error) {
	respBody, err := c.doRequest("GET", "/pool/dataset", nil)
	if err != nil {
		return nil, err
	}

	var allDatasets []Dataset
	if err := json.Unmarshal(respBody, &allDatasets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Filter datasets by pool name
	var datasets []Dataset
	for _, ds := range allDatasets {
		if ds.Pool == poolName || ds.ID == poolName {
			datasets = append(datasets, ds)
		} else if poolName != "" && len(ds.ID) > len(poolName) && ds.ID[:len(poolName)+1] == poolName+"/" {
			// Include child datasets
			datasets = append(datasets, ds)
		}
	}

	return datasets, nil
}

func (c *Client) CreateDataset(poolName, datasetName string, quotaGB int) (string, error) {
	fullName := fmt.Sprintf("%s/%s", poolName, datasetName)

	payload := map[string]interface{}{
		"name": fullName,
		"type": "FILESYSTEM",
	}

	// Add quota if specified (convert GB to bytes)
	if quotaGB > 0 {
		quotaBytes := int64(quotaGB) * 1024 * 1024 * 1024
		payload["refquota"] = quotaBytes
	}

	respBody, err := c.doRequest("POST", "/pool/dataset", payload)
	if err != nil {
		return "", err
	}

	var result Dataset
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.ID, nil
}

func (c *Client) DeleteDataset(datasetID string) error {
	// URL encode the dataset ID to handle slashes
	encodedID := url.PathEscape(datasetID)
	endpoint := fmt.Sprintf("/pool/dataset/id/%s", encodedID)

	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

func (c *Client) ClearDataset(datasetID string) error {
	// Delete the dataset and recreate it to wipe all data
	// This is done by deleting with recursive=true and then recreating
	encodedID := url.PathEscape(datasetID)
	endpoint := fmt.Sprintf("/pool/dataset/id/%s", encodedID)

	// First, get the current dataset info to preserve settings
	getEndpoint := fmt.Sprintf("/pool/dataset/id/%s", encodedID)
	respBody, err := c.doRequest("GET", getEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to get dataset info: %w", err)
	}

	var dataset Dataset
	if err := json.Unmarshal(respBody, &dataset); err != nil {
		return fmt.Errorf("failed to unmarshal dataset: %w", err)
	}

	// Delete the dataset with all its contents
	deleteBody := map[string]interface{}{
		"recursive": true,
		"force":     true,
	}
	_, err = c.doRequest("DELETE", endpoint, deleteBody)
	if err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}

	// Recreate the dataset
	createPayload := map[string]interface{}{
		"name": dataset.ID,
		"type": dataset.Type,
	}

	_, err = c.doRequest("POST", "/pool/dataset", createPayload)
	if err != nil {
		return fmt.Errorf("failed to recreate dataset: %w", err)
	}

	return nil
}
