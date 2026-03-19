package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

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

var transactionalCmd = &cobra.Command{
	Use:   "transactional",
	Short: "Manage transactional emails",
}

var transactionalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List published transactional emails",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		client := api.NewClient(cfg.EndpointURL, cfg.APIKey)

		var emails []api.TransactionalEmail
		if params.Cursor != "" {
			emails, _, err = client.ListTransactional(params)
		} else {
			emails, err = api.Paginate(func(cursor string) ([]api.TransactionalEmail, *api.Pagination, error) {
				return client.ListTransactional(api.PaginationParams{
					PerPage: params.PerPage,
					Cursor:  cursor,
				})
			})
		}
		if err != nil {
			return err
		}

		if isJSONOutput() {
			if emails == nil {
				emails = []api.TransactionalEmail{}
			}
			return printJSON(emails)
		}

		if len(emails) == 0 {
			fmt.Println("No transactional emails found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tLAST UPDATED\tVARIABLES")
		for _, e := range emails {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.ID, e.Name, e.LastUpdated, strings.Join(e.DataVariables, ", "))
		}
		w.Flush()

		return nil
	},
}

var transactionalSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a transactional email",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("email")
		id, _ := cmd.Flags().GetString("id")
		dataRaw, _ := cmd.Flags().GetString("data")

		req := api.SendTransactionalRequest{
			Email:           email,
			TransactionalID: id,
		}

		if cmd.Flags().Changed("add-to-audience") {
			v, _ := cmd.Flags().GetBool("add-to-audience")
			req.AddToAudience = &v
		}

		if dataRaw != "" {
			if err := json.Unmarshal([]byte(dataRaw), &req.DataVariables); err != nil {
				return fmt.Errorf("--data must be a valid JSON object: %w", err)
			}
		}

		paths, _ := cmd.Flags().GetStringArray("attachment")
		for _, path := range paths {
			a, err := attachmentFromPath(path)
			if err != nil {
				return err
			}
			req.Attachments = append(req.Attachments, a)
		}

		client := api.NewClient(cfg.EndpointURL, cfg.APIKey)
		if err := client.SendTransactional(req); err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(Result{Success: true})
		}
		fmt.Println("Sent.")
		return nil
	},
}

func init() {
	addPaginationFlags(transactionalListCmd)
	transactionalCmd.AddCommand(transactionalListCmd)

	transactionalSendCmd.Flags().String("email", "", "Recipient email address")
	transactionalSendCmd.Flags().String("id", "", "Transactional email ID")
	transactionalSendCmd.Flags().BoolP("add-to-audience", "a", false, "Create a contact if one doesn't exist")
	transactionalSendCmd.Flags().String("data", "", "Data variables as a JSON object")
	transactionalSendCmd.Flags().StringArrayP("attachment", "A", nil, "Path to a file to attach (repeatable)")
	transactionalSendCmd.MarkFlagRequired("email")
	transactionalSendCmd.MarkFlagRequired("id")
	transactionalCmd.AddCommand(transactionalSendCmd)

	rootCmd.AddCommand(transactionalCmd)
}
