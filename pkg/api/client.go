package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kacperkwapisz/sortpath/internal/config"
)

type LLMResponse struct {
	Path   string
	Reason string
}

func QueryLLM(conf *config.Config, prompt string) (*LLMResponse, error) {
	reqBody := map[string]interface{}{
		"model": conf.Model,
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", conf.APIBase+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+conf.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(b))
	}
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if len(apiResp.Choices) == 0 {
		return nil, errors.New("no response from model")
	}
	// Parse XML output (simple, not robust)
	content := apiResp.Choices[0].Message.Content
	path, reason := parseXML(content)
	return &LLMResponse{Path: path, Reason: reason}, nil
}

func parseXML(s string) (string, string) {
	// Very basic XML extraction for <path> and <reason>
	get := func(tag string) string {
		start := fmt.Sprintf("<%s>", tag)
		end := fmt.Sprintf("</%s>", tag)
		i := len(start) + findIndex(s, start)
		j := findIndex(s, end)
		if i < len(start) || j < 0 {
			return ""
		}
		return s[i:j]
	}
	return get("path"), get("reason")
}

func findIndex(s, sub string) int {
	idx := -1
	if i := bytes.Index([]byte(s), []byte(sub)); i >= 0 {
		idx = i
	}
	return idx
}
