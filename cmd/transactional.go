package cmd

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/cmdutil"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func parseDataVars(vars []string, jsonFile string) (map[string]any, error) {
	var m map[string]any
	if jsonFile != "" {
		var err error
		m, err = cmdutil.ParseJSONFile("json-vars", jsonFile)
		if err != nil {
			return nil, err
		}
	}
	return cmdutil.ParseKeyValuePairs("var", vars, m)
}

func attachmentFromPath(path string) (api.Attachment, error) {
	info, err := os.Stat(path)
	if err != nil {
		return api.Attachment{}, fmt.Errorf("attachment %q: %w", path, err)
	}
	if info.IsDir() {
		return api.Attachment{}, fmt.Errorf("attachment %q: is a directory", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return api.Attachment{}, fmt.Errorf("attachment %q: %w", path, err)
	}

	sniff := data
	if len(sniff) > 512 {
		sniff = sniff[:512]
	}
	contentType := http.DetectContentType(sniff)

	return api.Attachment{
		Filename:    filepath.Base(path),
		ContentType: contentType,
		Data:        base64.StdEncoding.EncodeToString(data),
	}, nil
}

func runTransactionalList(cfg *config.Config, params api.PaginationParams) ([]api.TransactionalEmail, error) {
	client := newAPIClient(cfg)
	if params.Cursor != "" {
		emails, _, err := client.ListTransactional(params)
		return emails, err
	}
	return api.Paginate(func(cursor string) ([]api.TransactionalEmail, *api.Pagination, error) {
		return client.ListTransactional(api.PaginationParams{
			PerPage: params.PerPage,
			Cursor:  cursor,
		})
	})
}

func runTransactionalSend(cfg *config.Config, req api.SendTransactionalRequest) error {
	return newAPIClient(cfg).SendTransactional(req)
}

var transactionalCmd = &cobra.Command{
	Use:   "transactional",
	Short: "Manage transactional emails",
}

var transactionalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List published transactional emails",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validatePickFlags(cmd); err != nil {
			return err
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		emails, err := runTransactionalList(cfg, params)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			if emails == nil {
				emails = []api.TransactionalEmail{}
			}
			return printJSON(cmd.OutOrStdout(), emails)
		}

		if len(emails) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No transactional emails found.")
			return nil
		}

		headers := []string{"ID", "NAME", "LAST UPDATED", "VARIABLES"}
		rows := make([][]string, 0, len(emails))
		for _, e := range emails {
			rows = append(rows, []string{e.ID, e.Name, e.LastUpdated, strings.Join(e.DataVariables, ", ")})
		}

		if isPicking(cmd) {
			return runPicker(headers, rows, copyColumnAction(rows, 0, "transactional ID", cmd.OutOrStdout()))
		}

		t := newStyledTable(cmd.OutOrStdout(), headers...)
		for _, r := range rows {
			t.Row(r...)
		}
		return t.Render()
	},
}

func transactionalSendRunE(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	id := args[0]
	email, _ := cmd.Flags().GetString("email")
	idempotencyKey, _ := cmd.Flags().GetString("idempotency-key")

	req := api.SendTransactionalRequest{
		Email:           email,
		TransactionalID: id,
		IdempotencyKey:  idempotencyKey,
	}

	if cmd.Flags().Changed("add-to-audience") {
		v, _ := cmd.Flags().GetBool("add-to-audience")
		req.AddToAudience = &v
	}

	varPairs, _ := cmd.Flags().GetStringArray("var")
	jsonFile, _ := cmd.Flags().GetString("json-vars")
	dataVars, err := parseDataVars(varPairs, jsonFile)
	if err != nil {
		return err
	}
	if len(dataVars) > 0 {
		req.DataVariables = dataVars
	}

	paths, _ := cmd.Flags().GetStringArray("attachment")
	for _, path := range paths {
		a, err := attachmentFromPath(path)
		if err != nil {
			return err
		}
		req.Attachments = append(req.Attachments, a)
	}

	if err := runTransactionalSend(cfg, req); err != nil {
		return err
	}

	if isJSONOutput() {
		return printJSON(cmd.OutOrStdout(), Result{Success: true})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Sent.")
	return nil
}

func addTransactionalSendFlags(cmd *cobra.Command) {
	cmd.Flags().String("email", "", "Recipient email address")
	cmd.Flags().BoolP("add-to-audience", "a", false, "Create a contact if one doesn't exist")
	cmd.Flags().StringArrayP("var", "v", nil, "Data variable as KEY=value (repeatable)")
	cmd.Flags().StringP("json-vars", "j", "", "Path to a JSON file of data variables")
	cmd.Flags().StringArrayP("attachment", "A", nil, "Path to a file to attach (repeatable)")
	cmd.Flags().String("idempotency-key", "", "Idempotency key to prevent duplicate sends")
	cmd.MarkFlagRequired("email")
}

var transactionalSendCmd = &cobra.Command{
	Use:   "send <id>",
	Short: "Send a transactional email",
	Args:  cobra.ExactArgs(1),
	RunE:  transactionalSendRunE,
}

func init() {
	addPaginationFlags(transactionalListCmd)
	addPickFlag(transactionalListCmd)
	transactionalCmd.AddCommand(transactionalListCmd)

	addTransactionalSendFlags(transactionalSendCmd)
	transactionalCmd.AddCommand(transactionalSendCmd)

	rootCmd.AddCommand(transactionalCmd)
}
