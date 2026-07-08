package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const configFileName = "config.json"

// CLIConfig is the top-level configuration for the actuated CLI.
// It stores per-controller settings keyed by the controller URL.
type CLIConfig struct {
	Controllers map[string]ControllerConfig `json:"controllers"`
}

// ControllerConfig holds the configuration for a single actuated controller.
type ControllerConfig struct {
	Platform string `json:"platform"`
	Token    string `json:"token,omitempty"`

	// OIDC fields: used to refresh short-lived id_tokens (JWTs).
	// Currently used for GitLab OIDC authentication.
	// The refresh_token is long-lived and may be rotated on each refresh.
	RefreshToken string `json:"refresh_token,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	URL          string `json:"url,omitempty"`

	// IDToken is a cached OIDC id_token (JWT). It has a short validity
	// (e.g. ~2 minutes for GitLab) but is reused when still valid to
	// avoid unnecessary refresh calls.
	IDToken string `json:"id_token,omitempty"`
}

// configFilePath returns the full path to the config file.
func configFilePath() string {
	return os.ExpandEnv(path.Join(basePath, configFileName))
}

// loadConfig reads the config file from disk.
// Returns an empty config (not an error) if the file does not exist.
func loadConfig() (*CLIConfig, error) {
	data, err := os.ReadFile(configFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{
				Controllers: make(map[string]ControllerConfig),
			}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg CLIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.Controllers == nil {
		cfg.Controllers = make(map[string]ControllerConfig)
	}

	return &cfg, nil
}

// saveConfig writes the config to disk, creating the directory if needed.
func saveConfig(cfg *CLIConfig) error {
	os.MkdirAll(os.ExpandEnv(basePath), 0755)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(configFilePath(), data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// normalizeURL ensures the URL has no trailing slash for consistent map keys.
func normalizeURL(u string) string {
	return strings.TrimRight(u, "/")
}

// getControllerURL returns the actuated controller URL.
// It reads from the ACTUATED_URL environment variable.
func getControllerURL() (string, error) {
	v, ok := os.LookupEnv("ACTUATED_URL")
	if !ok || v == "" {
		return "", fmt.Errorf("ACTUATED_URL environment variable is not set, see the CLI tab in the dashboard for instructions")
	}

	if strings.Contains(v, "o6s.io") {
		return "", fmt.Errorf("the ACTUATED_URL loaded from your shell is out of date, visit https://dashboard.actuated.com and click \"CLI\" for the latest URL and edit export ACTUATED_URL=... in your bash or zsh profile")
	}

	return normalizeURL(v), nil
}

// getControllerConfig returns the controller config for the current controller URL.
// If no config entry exists for the URL, it returns a zero-value ControllerConfig
// and found=false.
func getControllerConfig() (ControllerConfig, string, bool, error) {
	controllerURL, err := getControllerURL()
	if err != nil {
		return ControllerConfig{}, "", false, err
	}

	cfg, err := loadConfig()
	if err != nil {
		return ControllerConfig{}, controllerURL, false, err
	}

	cc, found := cfg.Controllers[controllerURL]
	return cc, controllerURL, found, nil
}

// jwtLeeway is subtracted from the token's expiry time to ensure the
// token is still usable by the upstream API when it arrives.
const jwtLeeway = 30 * time.Second

// isIDTokenValid checks whether a cached id_token (JWT) is still valid.
// It decodes the payload (without signature verification — the controller
// does that) and compares the "exp" claim against the current time minus
// a leeway. Returns true if the token can be reused.
func isIDTokenValid(idToken string) bool {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return false
	}

	// JWT payload is base64url-encoded without padding
	payload := parts[1]
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return false
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return false
	}

	if claims.Exp == 0 {
		return false
	}

	expiry := time.Unix(claims.Exp, 0)
	return time.Now().Add(jwtLeeway).Before(expiry)
}

// refreshOIDCToken uses the stored refresh_token to obtain a fresh id_token
// from the OIDC token endpoint. It also updates the config file with the
// rotated refresh_token. Currently used for GitLab OIDC authentication.
//
// Returns the new id_token (JWT) to use as a bearer token.
func refreshOIDCToken(controllerURL string, cc ControllerConfig) (string, error) {
	if cc.RefreshToken == "" {
		return "", fmt.Errorf("no refresh_token stored for %s, run \"actuated-cli auth --url %s\" to re-authenticate", controllerURL, controllerURL)
	}

	if cc.URL == "" {
		return "", fmt.Errorf("no url stored for %s, run \"actuated-cli auth --url %s\" to re-authenticate", controllerURL, controllerURL)
	}

	if cc.ClientID == "" {
		return "", fmt.Errorf("no client_id stored for %s, run \"actuated-cli auth --url %s\" to re-authenticate", controllerURL, controllerURL)
	}

	tokenURL := fmt.Sprintf("%s/oauth/token", cc.URL)

	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("refresh_token", cc.RefreshToken)
	params.Set("client_id", cc.ClientID)

	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewBuffer([]byte(params.Encode())))
	if err != nil {
		return "", fmt.Errorf("creating refresh request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("refreshing OIDC token: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading refresh response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OIDC token refresh failed (HTTP %d): %s\nRun \"actuated-cli auth --url %s\" to re-authenticate",
			res.StatusCode, string(body), controllerURL)
	}

	var tokenRes struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return "", fmt.Errorf("parsing refresh response: %w", err)
	}

	if tokenRes.Error != "" {
		return "", fmt.Errorf("OIDC token refresh error: %s - %s\nRun \"actuated-cli auth --url %s\" to re-authenticate",
			tokenRes.Error, tokenRes.ErrorDesc, controllerURL)
	}

	if tokenRes.IDToken == "" {
		return "", fmt.Errorf("refresh response did not include id_token, ensure the OAuth app has the openid scope")
	}

	// Update the config with the rotated refresh token and cached id_token
	cfg, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("loading config for refresh token update: %w", err)
	}

	updated := cfg.Controllers[controllerURL]
	updated.RefreshToken = tokenRes.RefreshToken
	updated.IDToken = tokenRes.IDToken
	cfg.Controllers[controllerURL] = updated

	if err := saveConfig(cfg); err != nil {
		return "", fmt.Errorf("saving rotated refresh token: %w", err)
	}

	return tokenRes.IDToken, nil
}
