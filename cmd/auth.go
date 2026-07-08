// MIT License
// Copyright (c) 2023 Alex Ellis, OpenFaaS Ltd

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
)

func makeAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate and save credentials to the config file",
		Example: `  # Authenticate with GitHub (default)
  actuated-cli auth

  # Authenticate with GitLab (gitlab.com)
  actuated-cli auth --platform gitlab

  # Authenticate with a self-managed GitLab instance
  actuated-cli auth --platform gitlab --gitlab-url https://gitlab.example.com
`,
	}

	cmd.RunE = runAuthE

	cmd.Flags().String("platform", "", "Platform to authenticate with (github or gitlab)")
	cmd.Flags().String("url", "", "URL of the actuated controller (can also be set via ACTUATED_URL env var)")
	cmd.Flags().String("gitlab-url", "https://gitlab.com", "GitLab instance URL for authentication")
	cmd.Flags().String("client-id", "", "OAuth application client ID for GitLab authentication")

	return cmd
}

func runAuthE(cmd *cobra.Command, args []string) error {

	platform, err := cmd.Flags().GetString("platform")
	if err != nil {
		return err
	}

	if len(platform) == 0 {
		platform = PlatformGitHub
	}

	if err := validatePlatform(platform); err != nil {
		return err
	}

	// Resolve the controller URL from --url flag or ACTUATED_URL env var
	controllerURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	if controllerURL == "" {
		controllerURL = os.Getenv("ACTUATED_URL")
	}

	if controllerURL == "" {
		return fmt.Errorf("controller URL is required, set --url flag or ACTUATED_URL environment variable")
	}

	controllerURL = normalizeURL(controllerURL)

	var token string

	switch platform {
	case PlatformGitHub:
		token, err = runGitHubAuth()
		if err != nil {
			return err
		}

		if len(token) == 0 {
			return fmt.Errorf("no token was obtained")
		}

		// Save to config file
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		cfg.Controllers[controllerURL] = ControllerConfig{
			Platform: platform,
			Token:    token,
		}

		if err := saveConfig(cfg); err != nil {
			return err
		}

		// Also write to legacy PAT file for backward compatibility
		os.MkdirAll(os.ExpandEnv(basePath), 0755)
		if err := os.WriteFile(os.ExpandEnv(path.Join(basePath, "PAT")), []byte(token), 0600); err != nil {
			return fmt.Errorf("writing legacy PAT file: %w", err)
		}

	case PlatformGitLab:
		gitlabURL, err := cmd.Flags().GetString("gitlab-url")
		if err != nil {
			return err
		}
		clientID, err := cmd.Flags().GetString("client-id")
		if err != nil {
			return err
		}

		tokenRes, err := runGitLabAuth(gitlabURL, clientID)
		if err != nil {
			return err
		}

		if tokenRes.RefreshToken == "" {
			return fmt.Errorf("no refresh_token was obtained from GitLab")
		}

		// Save refresh token, client_id, and gitlab_url to config.
		// The short-lived id_token will be obtained via refresh before each API call.
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		resolvedClientID := clientID
		if resolvedClientID == "" {
			resolvedClientID = gitlabDefaultClientID
		}

		cfg.Controllers[controllerURL] = ControllerConfig{
			Platform:     platform,
			RefreshToken: tokenRes.RefreshToken,
			IDToken:      tokenRes.IDToken,
			ClientID:     resolvedClientID,
			URL:          gitlabURL,
		}

		if err := saveConfig(cfg); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	fmt.Printf("Credentials saved to: %s\n", configFilePath())
	fmt.Printf("  Controller: %s\n", controllerURL)
	fmt.Printf("  Platform:   %s\n", platform)

	return nil
}

func runGitHubAuth() (string, error) {
	clientID := "8c5dc5d9750ff2a8396a"

	dcParams := url.Values{}
	dcParams.Set("client_id", clientID)
	dcParams.Set("redirect_uri", "http://127.0.0.1:31111/oauth/callback")
	dcParams.Set("scope", "read:user,read:org,user:email")

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/device/code", bytes.NewBuffer([]byte(dcParams.Encode())))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, string(body))
	}

	auth := DeviceAuth{}

	if err := json.Unmarshal(body, &auth); err != nil {
		return "", err
	}

	fmt.Printf("Please visit: %s\n", auth.VerificationURI)
	fmt.Printf("and enter the code: %s\n", auth.UserCode)

	for i := 0; i < 60; i++ {
		urlv := url.Values{}
		urlv.Set("client_id", clientID)
		urlv.Set("device_code", auth.DeviceCode)

		urlv.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

		req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewBuffer([]byte(urlv.Encode())))
		if err != nil {
			return "", err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}

		body, _ := io.ReadAll(res.Body)
		parts, err := url.ParseQuery(string(body))
		if err != nil {
			return "", err
		}
		if parts.Get("error") == "authorization_pending" {
			fmt.Println("Waiting for authorization...")
			time.Sleep(time.Second * 5)
			continue
		} else if parts.Get("access_token") != "" {
			return parts.Get("access_token"), nil
		} else {
			return "", fmt.Errorf("something went wrong")
		}
	}

	return "", fmt.Errorf("timed out waiting for authorization")
}

const gitlabDefaultClientID = "222c0ecd207277ddd78864e94f72709663babf81dfd70513cfb82334ba4a8a2a"

func runGitLabAuth(gitlabURL, clientID string) (*GitLabTokenResponse, error) {
	if len(clientID) == 0 {
		clientID = gitlabDefaultClientID
	}

	deviceURL := fmt.Sprintf("%s/oauth/authorize_device", gitlabURL)
	tokenURL := fmt.Sprintf("%s/oauth/token", gitlabURL)

	dcParams := url.Values{}
	dcParams.Set("client_id", clientID)
	dcParams.Set("scope", "openid")

	req, err := http.NewRequest(http.MethodPost, deviceURL, bytes.NewBuffer([]byte(dcParams.Encode())))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from GitLab device auth: %d, body: %s", res.StatusCode, string(body))
	}

	auth := GitLabDeviceAuth{}
	if err := json.Unmarshal(body, &auth); err != nil {
		return nil, err
	}

	fmt.Printf("Please visit: %s\n", auth.VerificationURI)
	fmt.Printf("and enter the code: %s\n", auth.UserCode)

	interval := auth.Interval
	if interval < 5 {
		interval = 5
	}

	maxAttempts := auth.ExpiresIn / interval
	if maxAttempts <= 0 {
		maxAttempts = 60
	}

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(time.Duration(interval) * time.Second)

		tokenParams := url.Values{}
		tokenParams.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		tokenParams.Set("device_code", auth.DeviceCode)
		tokenParams.Set("client_id", clientID)

		req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewBuffer([]byte(tokenParams.Encode())))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, _ := io.ReadAll(res.Body)

		var tokenRes GitLabTokenResponse
		if err := json.Unmarshal(body, &tokenRes); err != nil {
			return nil, err
		}

		if tokenRes.Error == "authorization_pending" {
			fmt.Println("Waiting for authorization...")
			continue
		} else if tokenRes.Error == "slow_down" {
			interval += 5
			fmt.Println("Waiting for authorization...")
			continue
		} else if tokenRes.Error == "expired_token" {
			return nil, fmt.Errorf("device code expired, please try again")
		} else if tokenRes.Error == "access_denied" {
			return nil, fmt.Errorf("authorization request was denied")
		} else if len(tokenRes.Error) > 0 {
			return nil, fmt.Errorf("error from GitLab: %s - %s", tokenRes.Error, tokenRes.ErrorDescription)
		}

		if len(tokenRes.IDToken) > 0 {
			return &tokenRes, nil
		}

		if len(tokenRes.AccessToken) > 0 {
			return nil, fmt.Errorf("received access_token but no id_token, ensure the OAuth app has the openid scope enabled")
		}

		return nil, fmt.Errorf("unexpected response from GitLab token endpoint")
	}

	return nil, fmt.Errorf("timed out waiting for authorization")
}

// DeviceAuth is the device auth response from GitHub and is
// used to exchange for a personal access token
type DeviceAuth struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// GitLabDeviceAuth is the device authorization response from GitLab.
type GitLabDeviceAuth struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// GitLabTokenResponse is the token response from GitLab's OAuth token endpoint.
// When the openid scope is requested, the response includes an id_token (JWT).
type GitLabTokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	Scope            string `json:"scope,omitempty"`
	CreatedAt        int64  `json:"created_at,omitempty"`
	IDToken          string `json:"id_token,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}
