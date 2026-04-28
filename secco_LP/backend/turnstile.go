package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type turnstileResponse struct {
	Success bool `json:"success"`
}

func verifyTurnstile(secret, token, remoteIP string) (bool, error) {
	if secret == "" {
		return true, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.PostForm(turnstileVerifyURL, url.Values{
		"secret":   {secret},
		"response": {token},
		"remoteip": {remoteIP},
	})
	if err != nil {
		return false, fmt.Errorf("turnstile request: %w", err)
	}
	defer resp.Body.Close()

	var result turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("turnstile decode: %w", err)
	}

	return result.Success, nil
}
