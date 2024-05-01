package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v52/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func makeSshConnect() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connected to an SSH session",
		Example: `  # Connect to the first session
  actuated-cli ssh connect

  # Connect to a specific session
  actuated-cli ssh connect HOST
`,
		Aliases: []string{"c"},
	}

	cmd.RunE = runSshConnectE

	cmd.Flags().Bool("print", false, "Print the SSH command instead of running it")

	return cmd
}

func runSshConnectE(cmd *cobra.Command, args []string) error {
	pat, err := getPat(cmd)
	if err != nil {
		return err
	}

	printOnly, err := cmd.Flags().GetBool("print")
	if err != nil {
		return err
	}

	host := ""
	hostIndex := 0
	useIndex := false

	if len(args) > 0 {
		host = args[0]
		if i, err := strconv.Atoi(host); err == nil {
			useIndex = true
			hostIndex = i
		}
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

	sessions := []sshSession{}
	if err = json.NewDecoder(res.Body).Decode(&sessions); err != nil {
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

	if len(onlyActor) == 0 {
		return fmt.Errorf("no sessions found")
	}

	var found *sshSession
	for i, session := range onlyActor {
		if host == "" {
			found = &session
			break
		} else if session.Hostname == host {
			found = &session
			break
		} else if useIndex && hostIndex == i+1 {
			found = &session
			break
		}
	}

	// Try a fuzzy match
	if found == nil {
		for _, session := range onlyActor {
			if strings.HasPrefix(session.Hostname, host) {
				found = &session
				break
			}
		}
	}

	if found == nil {
		return fmt.Errorf("no session found for hostname %s", host)
	}

	us, _ := url.Parse(SshGw)
	sshArgs := []string{"-p", strconv.Itoa(found.Port), "runner@" + us.Host, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}

	if printOnly {
		fmt.Printf("ssh %s\n", strings.Join(sshArgs, " "))
		return nil
	}

	sshCmd := exec.CommandContext(ctx, "ssh", sshArgs...)

	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout

	if err = sshCmd.Run(); err != nil {
		return err
	}

	return nil
}
