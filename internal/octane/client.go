package octane

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Server      string
	SharedSpace string
	Workspace   string
	User        string
	Password    string
	AccessToken string
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type PushResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ResultStatus struct {
	Status       string
	Until        string
	ErrorDetails string
	RawLog       string
	TimedOut     bool
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) PushResult(payload string) (PushResponse, error) {
	url := fmt.Sprintf("%s/api/shared_spaces/%s/workspaces/%s/test-results",
		strings.TrimSuffix(c.cfg.Server, "/"), c.cfg.SharedSpace, c.cfg.Workspace)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return PushResponse{}, err
	}
	req.Header.Set("Content-Type", "application/xml")
	applyAuth(req, c.cfg)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PushResponse{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return PushResponse{}, fmt.Errorf("octane API error %d: %s", resp.StatusCode, string(body))
	}
	var out PushResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&out); err != nil {
		return PushResponse{}, nil
	}
	return out, nil
}

func ParseResultLog(logText string) ResultStatus {
	lines := strings.Split(logText, "\n")
	statusRE := regexp.MustCompile(`(?i)^status\s*:\s*(.+)$`)
	untilRE := regexp.MustCompile(`(?i)^until\s*:\s*(.+)$`)
	out := ResultStatus{RawLog: logText}
	var errors []string
	for _, raw := range lines {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			continue
		}
		if m := statusRE.FindStringSubmatch(line); len(m) > 1 {
			out.Status = strings.TrimSpace(m[1])
			continue
		}
		if m := untilRE.FindStringSubmatch(line); len(m) > 1 {
			out.Until = strings.TrimSpace(m[1])
			continue
		}
		errors = append(errors, line)
	}
	if len(errors) > 0 {
		out.ErrorDetails = strings.Join(errors, " | ")
	}
	return out
}

func (c *Client) GetResultStatus(resultID string) (ResultStatus, error) {
	url := fmt.Sprintf("%s/api/shared_spaces/%s/workspaces/%s/test-results/%s/log",
		strings.TrimSuffix(c.cfg.Server, "/"), c.cfg.SharedSpace, c.cfg.Workspace, resultID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ResultStatus{}, err
	}
	applyAuth(req, c.cfg)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ResultStatus{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ResultStatus{}, fmt.Errorf("octane API status check error %d: %s", resp.StatusCode, string(body))
	}
	return ParseResultLog(string(body)), nil
}

func (c *Client) WaitForResultCompletion(resultID string, timeoutSec int, pollInterval time.Duration) (ResultStatus, error) {
	attempts := (timeoutSec * int(time.Second)) / int(pollInterval)
	if attempts < 1 {
		attempts = 1
	}
	var last ResultStatus
	for i := 0; i < attempts; i++ {
		status, err := c.GetResultStatus(resultID)
		if err != nil {
			return ResultStatus{}, err
		}
		last = status
		if !isInProgress(status.Status) {
			last.TimedOut = false
			return last, nil
		}
		if i < attempts-1 {
			time.Sleep(pollInterval)
		}
	}
	last.TimedOut = true
	return last, nil
}

func isInProgress(status string) bool {
	return status == "queued" || status == "running"
}

func applyAuth(req *http.Request, cfg Config) {
	if cfg.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
		return
	}
	if cfg.User != "" && cfg.Password != "" {
		raw := base64.StdEncoding.EncodeToString([]byte(cfg.User + ":" + cfg.Password))
		req.Header.Set("Authorization", "Basic "+raw)
	}
}
