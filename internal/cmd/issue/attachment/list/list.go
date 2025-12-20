package list

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `List attachments of an issue.`
	examples = `$ jira issue attachment list ISSUE-1

# Output as JSON
$ jira issue attachment list ISSUE-1 --raw`
)

func NewCmdAttachmentList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List attachments of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"ls"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: listAttachments,
	}

	cmd.Flags().Bool("raw", false, "Print raw JSON output")

	return &cmd
}

func listAttachments(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())

	client := api.DefaultClient(params.debug)

	issue, err := func() (*jira.Issue, error) {
		s := cmdutil.Info("Fetching attachments...")
		defer s.Stop()

		return api.ProxyGetIssue(client, params.issueKey)
	}()
	cmdutil.ExitIfError(err)

	if len(issue.Fields.Attachments) == 0 {
		fmt.Println("No attachments found.")
		return
	}

	if params.raw {
		printRaw(issue.Fields.Attachments)
		return
	}

	printTable(issue.Fields.Attachments)
}

type listParams struct {
	issueKey string
	raw      bool
	debug    bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *listParams {
	if len(args) < 1 {
		cmdutil.Failed("Issue key is required")
	}

	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	raw, err := flags.GetBool("raw")
	cmdutil.ExitIfError(err)

	return &listParams{
		issueKey: issueKey,
		raw:      raw,
		debug:    debug,
	}
}

func printRaw(attachments []jira.Attachment) {
	data, err := json.MarshalIndent(attachments, "", "  ")
	cmdutil.ExitIfError(err)
	fmt.Println(string(data))
}

func printTable(attachments []jira.Attachment) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "ID\tFILENAME\tSIZE\tAUTHOR\tCREATED")

	for _, a := range attachments {
		created := cmdutil.FormatDateTimeHuman(a.Created, jira.RFC3339MilliLayout)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			a.ID,
			a.Filename,
			formatSize(a.Size),
			a.Author.DisplayName,
			created,
		)
	}
	w.Flush()
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
