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
		Short: "Authenticate to GitHub to obtain a token and save it to $HOME/.actuated/PAT",
	}

	cmd.RunE = runAuthE

	return cmd
}

func runAuthE(cmd *cobra.Command, args []string) error {

	token := ""

	clientID := "8c5dc5d9750ff2a8396a"

	dcParams := url.Values{}
	dcParams.Set("client_id", clientID)
	dcParams.Set("redirect_uri", "http://127.0.0.1:31111/oauth/callback")
	dcParams.Set("scope", "read:user,read:org,user:email")

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/device/code", bytes.NewBuffer([]byte(dcParams.Encode())))

	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, string(body))
	}

	auth := DeviceAuth{}

	if err := json.Unmarshal(body, &auth); err != nil {
		return err
	}

	fmt.Printf("Please visit: %s\n", auth.VerificationURI)
	fmt.Printf("and enter the code: %s\n", auth.UserCode)

	for i := 0; i < 60; i++ {
		urlv := url.Values{}
		urlv.Set("client_id", clientID)
		urlv.Set("device_code", auth.DeviceCode)

		urlv.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewBuffer([]byte(urlv.Encode())))
		if err != nil {
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		body, _ := io.ReadAll(res.Body)
		parts, err := url.ParseQuery(string(body))
		if err != nil {
			return err
		}
		if parts.Get("error") == "authorization_pending" {
			fmt.Println("Waiting for authorization...")
			time.Sleep(time.Second * 5)
			continue
		} else if parts.Get("access_token") != "" {
			// fmt.Println(parts)
			token = parts.Get("access_token")

			break
		} else {
			return fmt.Errorf("something went wrong")
		}
	}

	const basePath = "$HOME/.actuated"
	os.Mkdir(os.ExpandEnv(basePath), 0755)

	if err := os.WriteFile(os.ExpandEnv(path.Join(basePath, "PAT")), []byte(token), 0644); err != nil {
		return err
	}

	fmt.Printf("Access token written to: %s\n", os.ExpandEnv(path.Join(basePath, "PAT")))

	return nil
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
