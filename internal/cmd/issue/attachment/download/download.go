package download

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Download attachments from an issue.`
	examples = `# Download by attachment ID
$ jira issue attachment download ISSUE-1 --id 12345

# Download by filename
$ jira issue attachment download ISSUE-1 --name "screenshot.png"

# Download all attachments
$ jira issue attachment download ISSUE-1 --all

# Download to a specific directory
$ jira issue attachment download ISSUE-1 --all --output-dir ./downloads

# Overwrite existing files
$ jira issue attachment download ISSUE-1 --all --overwrite`
)

func NewCmdAttachmentDownload() *cobra.Command {
	cmd := cobra.Command{
		Use:     "download ISSUE-KEY",
		Short:   "Download attachments from an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"dl"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: downloadAttachments,
	}

	cmd.Flags().String("id", "", "Attachment ID to download")
	cmd.Flags().String("name", "", "Attachment filename to download")
	cmd.Flags().StringP("output-dir", "o", ".", "Destination directory")
	cmd.Flags().Bool("all", false, "Download all attachments")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing files")

	return &cmd
}

func downloadAttachments(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())

	if params.id == "" && params.name == "" && !params.all {
		cmdutil.Failed("One of --id, --name, or --all is required")
	}

	client := api.DefaultClient(params.debug)

	issue, err := func() (*jira.Issue, error) {
		s := cmdutil.Info("Fetching issue attachments...")
		defer s.Stop()

		return api.ProxyGetIssue(client, params.issueKey)
	}()
	cmdutil.ExitIfError(err)

	attachments := issue.Fields.Attachments
	if len(attachments) == 0 {
		cmdutil.Failed("No attachments found in issue %s", params.issueKey)
	}

	var toDownload []jira.Attachment

	switch {
	case params.id != "":
		for _, a := range attachments {
			if a.ID == params.id {
				toDownload = append(toDownload, a)
				break
			}
		}
		if len(toDownload) == 0 {
			cmdutil.Failed("Attachment with ID %q not found", params.id)
		}

	case params.name != "":
		for _, a := range attachments {
			if a.Filename == params.name {
				toDownload = append(toDownload, a)
			}
		}
		if len(toDownload) == 0 {
			cmdutil.Failed("Attachment with name %q not found", params.name)
		}
		if len(toDownload) > 1 {
			cmdutil.Failed("Multiple attachments found with name %q, use --id instead", params.name)
		}

	case params.all:
		toDownload = attachments
	}

	if err := os.MkdirAll(params.outputDir, 0o755); err != nil {
		cmdutil.ExitIfError(fmt.Errorf("failed to create output directory: %w", err))
	}

	for _, a := range toDownload {
		err := downloadOne(client, a, params.outputDir, params.overwrite)
		if err != nil {
			cmdutil.Fail("Failed to download %s: %v", a.Filename, err)
		} else {
			cmdutil.Success("Downloaded %s", a.Filename)
		}
	}
}

type downloadParams struct {
	issueKey  string
	id        string
	name      string
	outputDir string
	all       bool
	overwrite bool
	debug     bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *downloadParams {
	if len(args) < 1 {
		cmdutil.Failed("Issue key is required")
	}

	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	id, err := flags.GetString("id")
	cmdutil.ExitIfError(err)

	name, err := flags.GetString("name")
	cmdutil.ExitIfError(err)

	outputDir, err := flags.GetString("output-dir")
	cmdutil.ExitIfError(err)

	all, err := flags.GetBool("all")
	cmdutil.ExitIfError(err)

	overwrite, err := flags.GetBool("overwrite")
	cmdutil.ExitIfError(err)

	return &downloadParams{
		issueKey:  issueKey,
		id:        id,
		name:      name,
		outputDir: outputDir,
		all:       all,
		overwrite: overwrite,
		debug:     debug,
	}
}

func downloadOne(client *jira.Client, a jira.Attachment, dir string, overwrite bool) error {
	s := cmdutil.Info(fmt.Sprintf("Downloading %s...", a.Filename))
	defer s.Stop()

	reader, err := api.ProxyGetAttachmentContent(client, a.ID)
	if err != nil {
		return err
	}

	return saveAttachment(reader, dir, a.Filename, overwrite)
}

func saveAttachment(r io.ReadCloser, dir, filename string, overwrite bool) error {
	defer func() { _ = r.Close() }()

	filename = sanitizeFilename(filename)
	path := filepath.Join(dir, filename)

	flags := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("file %q already exists, use --overwrite to replace", filename)
		}
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, r)
	return err
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	// filepath.Base on Unix doesn't treat backslash as separator
	name = strings.ReplaceAll(name, "\\", "_")
	if name == "" || name == "." || name == ".." {
		name = "attachment"
	}
	return name
}
