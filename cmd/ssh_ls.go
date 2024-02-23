package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/google/go-github/v52/github"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const SshGw = "https://sshgw.actuated.dev"

func makeSshList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List SSH sessions",
	}

	cmd.Flags().BoolP("json", "j", false, "Request output in JSON format")

	cmd.RunE = runSshListE

	return cmd
}

func runSshListE(cmd *cobra.Command, args []string) error {

	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return err
	}

	login := user.GetLogin()

	u, _ := url.Parse(SshGw)
	u.Path = "/list"
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	jsonFormat, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	table := tablewriter.NewWriter(buf)

	table.SetHeader([]string{"No.", "Actor", "Hostname", "RX", "TX", "Connected"})

	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoWrapText(false)

	var sessions []sshSession
	if err := json.NewDecoder(res.Body).Decode(&sessions); err != nil {
		return err
	}

	onlyActor := []sshSession{}
	for _, session := range sessions {
		if session.Actor == login {
			onlyActor = append(onlyActor, session)
		}
	}

	sort.Slice(onlyActor, func(i, j int) bool {
		return onlyActor[i].ConnectedAt > onlyActor[j].ConnectedAt
	})

	for i, session := range onlyActor {
		connectedAt, _ := time.Parse(time.RFC3339, session.ConnectedAt)
		since := time.Since(connectedAt).Round(time.Second)
		table.Append([]string{
			strconv.Itoa(i + 1),
			session.Actor,
			session.Hostname,
			strconv.Itoa(session.Rx),
			strconv.Itoa(session.Tx),
			since.String(),
		})
	}

	if jsonFormat {
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		if err := e.Encode(onlyActor); err != nil {
			return err
		}
		return nil
	}

	table.Render()

	cmd.Print(buf.String())

	return nil

}

type sshSession struct {
	ConnectedAt string `json:"ConnectedAt"`
	Command     string `json:"Command"`
	Hostname    string `json:"Hostname"`
	Port        int    `json:"Port"`
	Actor       string `json:"Actor"`
	Rx          int    `json:"Rx"`
	Tx          int    `json:"Tx"`
	Connections int    `json:"Connections"`
}
