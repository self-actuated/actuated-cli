package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"

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
  actuated-cli ssh connect --host HOST
`,
	}

	cmd.RunE = runSshConnectE

	cmd.Flags().String("host", "", "Host to connect to or leave blank to connect to the first")
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

	host, err := cmd.Flags().GetString("host")
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

	sessions := []sshSession{}
	if err = json.NewDecoder(res.Body).Decode(&sessions); err != nil {
		return err
	}

	if len(sessions) == 0 {
		return fmt.Errorf("no sessions found")
	}

	var found *sshSession
	for _, session := range sessions {
		if session.Actor != login {
			continue
		}

		if host == "" {
			found = &session
			break
		} else if session.Hostname == host {
			found = &session
			break
		}
	}

	if found == nil {
		return fmt.Errorf("no session found for hostname %s", host)
	}

	us, _ := url.Parse(SshGw)

	if printOnly {
		fmt.Printf("ssh -p %d runner@%s\n", found.Port, us.Host)
		return nil
	}

	sshCmd := exec.CommandContext(ctx, "ssh", "-p", strconv.Itoa(found.Port), "runner@"+us.Host)

	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout

	if err = sshCmd.Run(); err != nil {
		return err
	}

	return nil
}
