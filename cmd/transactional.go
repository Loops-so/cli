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
	for _, pair := range vars {
		idx := strings.IndexByte(pair, '=')
		if idx < 0 {
			return nil, fmt.Errorf("--var %q: expected KEY=value", pair)
		}
		if m == nil {
			m = make(map[string]any)
		}
		m[pair[:idx]] = pair[idx+1:]
	}
	return m, nil
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
	client := api.NewClient(cfg.EndpointURL, cfg.APIKey)
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
	return api.NewClient(cfg.EndpointURL, cfg.APIKey).SendTransactional(req)
}

var transactionalCmd = &cobra.Command{
	Use:   "transactional",
	Short: "Manage transactional emails",
}

var transactionalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List published transactional emails",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		w := newTableWriter(cmd.OutOrStdout())
		fmt.Fprintln(w, "ID\tNAME\tLAST UPDATED\tVARIABLES")
		for _, e := range emails {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.ID, e.Name, e.LastUpdated, strings.Join(e.DataVariables, ", "))
		}
		w.Flush()

		return nil
	},
}

func transactionalSendRunE(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	email, _ := cmd.Flags().GetString("email")
	id, _ := cmd.Flags().GetString("id")
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
	cmd.Flags().String("id", "", "Transactional email ID")
	cmd.Flags().BoolP("add-to-audience", "a", false, "Create a contact if one doesn't exist")
	cmd.Flags().StringArrayP("var", "v", nil, "Data variable as KEY=value (repeatable)")
	cmd.Flags().StringP("json-vars", "j", "", "Path to a JSON file of data variables")
	cmd.Flags().StringArrayP("attachment", "A", nil, "Path to a file to attach (repeatable)")
	cmd.Flags().String("idempotency-key", "", "Idempotency key to prevent duplicate sends")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("id")
}

var transactionalSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a transactional email",
	RunE:  transactionalSendRunE,
}

func init() {
	addPaginationFlags(transactionalListCmd)
	transactionalCmd.AddCommand(transactionalListCmd)

	addTransactionalSendFlags(transactionalSendCmd)
	transactionalCmd.AddCommand(transactionalSendCmd)

	rootCmd.AddCommand(transactionalCmd)
}
