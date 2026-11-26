package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v76/github"
	"github.com/olekukonko/tablewriter"
	"github.com/self-actuated/actuated-cli/pkg"
	"github.com/spf13/cobra"
)

type actuatedEvent struct {
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

func makeEvents() *cobra.Command {
	cmd := &cobra.Command{Use: "events [OWNER]", Short: "List recent notifications", Long: "List recent notifications across your GitHub organizations. The controller only returns events from the past day. --since filters those events.", Args: cobra.MaximumNArgs(1), RunE: runEventsE}
	cmd.Flags().BoolP("json", "j", false, "Output JSON")
	cmd.Flags().Duration("since", 0, "Only show events from the past duration (for example 12h)")
	return cmd
}

func eventOwners(client *github.Client, cmd *cobra.Command) ([]string, error) {
	ctx := cmd.Context()
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get GitHub user: %w", err)
	}
	owners := []string{user.GetLogin()}
	if owners[0] == "" {
		return nil, fmt.Errorf("GitHub returned an empty user login")
	}
	seen := map[string]bool{strings.ToLower(owners[0]): true}
	opts := &github.ListOrgMembershipsOptions{State: "active", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		memberships, response, err := client.Organizations.ListOrgMemberships(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("get GitHub organization memberships: %w", err)
		}
		for _, membership := range memberships {
			name := membership.GetOrganization().GetLogin()
			if membership.GetState() == "active" && name != "" && !seen[strings.ToLower(name)] {
				owners = append(owners, name)
				seen[strings.ToLower(name)] = true
			}
		}
		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}
	return owners, nil
}

func runEventsE(cmd *cobra.Command, args []string) error {
	since, err := cmd.Flags().GetDuration("since")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("since") && since <= 0 {
		return fmt.Errorf("--since must be a positive duration")
	}
	pat, err := getPat(cmd)
	if err != nil {
		return err
	}
	if pat == "" {
		return fmt.Errorf("pat is required")
	}
	var owners []string
	if len(args) == 1 {
		owner := strings.TrimSpace(args[0])
		if owner == "" {
			return fmt.Errorf("owner cannot be empty")
		}
		owners = []string{owner}
	} else {
		owners, err = eventOwners(github.NewClient(http.DefaultClient).WithAuthToken(pat), cmd)
		if err != nil {
			return err
		}
	}
	staff, err := cmd.Flags().GetBool("staff")
	if err != nil {
		return err
	}
	body, status, err := pkg.NewClient(http.DefaultClient, os.Getenv("ACTUATED_URL")).ListEvents(pat, strings.Join(owners, ","), staff)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", status, body)
	}
	var events []actuatedEvent
	if err := json.Unmarshal([]byte(body), &events); err != nil {
		return fmt.Errorf("decode events: %w", err)
	}
	events = filterEvents(events, since, time.Now())
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(events)
	}
	if len(events) == 0 {
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "No recent events.")
		return err
	}
	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetHeader([]string{"OWNER", "TIME", "MESSAGE"})
	table.SetAutoWrapText(false)
	for _, event := range events {
		table.Append([]string{event.Owner, event.CreatedAt.Format(time.RFC3339), event.Message})
	}
	table.Render()
	return nil
}

func filterEvents(events []actuatedEvent, since time.Duration, now time.Time) []actuatedEvent {
	result := make([]actuatedEvent, 0, len(events))
	for _, event := range events {
		if since == 0 || !event.CreatedAt.Before(now.Add(-since)) {
			result = append(result, event)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result
}
