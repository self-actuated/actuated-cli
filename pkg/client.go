package pkg

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (c *Client) ListJobs(patStr string, owner string, staff bool, json bool) (string, int, error) {

	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/job-queue"
	q := u.Query()

	if staff {
		q.Set("staff", "1")
	}

	if len(owner) > 0 {
		q.Set("owners", owner)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	if json {
		req.Header.Set("Accept", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+patStr)

	if os.Getenv("DEBUG") == "1" {
		sanitised := http.Header{}
		for k, v := range req.Header {

			if k == "Authorization" {
				v = []string{"redacted"}
			}
			sanitised[k] = v
		}

		fmt.Printf("URL %s\nHeaders: %v\n", u.String(), sanitised)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	return string(body), res.StatusCode, nil
}

func (c *Client) ListRunners(patStr string, owner string, staff bool, json bool) (string, int, error) {

	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/runners"
	q := u.Query()

	if staff {
		q.Set("staff", "1")
	}

	if len(owner) > 0 {
		q.Set("owner", owner)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	if json {
		req.Header.Set("Accept", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+patStr)

	if os.Getenv("DEBUG") == "1" {
		sanitised := http.Header{}
		for k, v := range req.Header {

			if k == "Authorization" {
				v = []string{"redacted"}
			}
			sanitised[k] = v
		}

		fmt.Printf("URL %s\nHeaders: %v\n", u.String(), sanitised)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	return string(body), res.StatusCode, nil
}

func (c *Client) Repair(patStr string, owner string, staff bool) (string, int, error) {

	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/repair"
	q := u.Query()

	if staff {
		q.Set("staff", "1")
	}

	q.Set("owner", owner)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	req.Header.Set("Authorization", "Bearer "+patStr)

	if os.Getenv("DEBUG") == "1" {
		sanitised := http.Header{}
		for k, v := range req.Header {

			if k == "Authorization" {
				v = []string{"redacted"}
			}
			sanitised[k] = v
		}

		fmt.Printf("URL %s\nHeaders: %v\n", u.String(), sanitised)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	return string(body), res.StatusCode, nil
}

func (c *Client) GetLogs(patStr, owner, host, id string, age time.Duration, staff bool) (string, int, error) {

	mins := int(age.Minutes())

	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/logs"

	q := u.Query()
	q.Set("owner", owner)
	q.Set("host", host)
	q.Set("age", fmt.Sprintf("%dm", mins))
	if len(id) > 0 {
		q.Set("id", id)
	}

	if staff {
		q.Set("staff", "1")
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	req.Header.Set("Authorization", "Bearer "+patStr)

	if os.Getenv("DEBUG") == "1" {
		sanitised := http.Header{}
		for k, v := range req.Header {

			if k == "Authorization" {
				v = []string{"redacted"}
			}
			sanitised[k] = v
		}

		fmt.Printf("URL %s\nHeaders: %v\n", u.String(), sanitised)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	return string(body), res.StatusCode, nil
}

func (c *Client) GetAgentLogs(patStr, owner, host string, age time.Duration, staff bool) (string, int, error) {

	mins := int(age.Minutes())

	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/service"

	q := u.Query()
	q.Set("owner", owner)
	q.Set("host", host)
	q.Set("age", fmt.Sprintf("%dm", mins))

	if staff {
		q.Set("staff", "1")
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	req.Header.Set("Authorization", "Bearer "+patStr)

	if os.Getenv("DEBUG") == "1" {
		sanitised := http.Header{}
		for k, v := range req.Header {

			if k == "Authorization" {
				v = []string{"redacted"}
			}
			sanitised[k] = v
		}

		fmt.Printf("URL %s\nHeaders: %v\n", u.String(), sanitised)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	return string(body), res.StatusCode, nil
}

func (c *Client) UpgradeAgent(patStr, owner, host string, force bool, staff bool) (string, int, error) {

	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/upgrade"

	q := u.Query()
	q.Set("owner", owner)
	q.Set("host", host)

	if staff {
		q.Set("staff", "1")
	}
	if force {
		q.Set("force", "1")

	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	req.Header.Set("Authorization", "Bearer "+patStr)

	if os.Getenv("DEBUG") == "1" {
		sanitised := http.Header{}
		for k, v := range req.Header {

			if k == "Authorization" {
				v = []string{"redacted"}
			}
			sanitised[k] = v
		}

		fmt.Printf("URL %s\nHeaders: %v\n", u.String(), sanitised)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	return string(body), res.StatusCode, nil
}

func (c *Client) RestartAgent(patStr, owner, host string, reboot bool, staff bool) (string, int, error) {

	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/restart"

	q := u.Query()
	q.Set("owner", owner)
	q.Set("host", host)

	if staff {
		q.Set("staff", "1")
	}
	if reboot {
		q.Set("reboot", "1")

	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	req.Header.Set("Authorization", "Bearer "+patStr)

	if os.Getenv("DEBUG") == "1" {
		sanitised := http.Header{}
		for k, v := range req.Header {

			if k == "Authorization" {
				v = []string{"redacted"}
			}
			sanitised[k] = v
		}

		fmt.Printf("URL %s\nHeaders: %v\n", u.String(), sanitised)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	return string(body), res.StatusCode, nil
}
