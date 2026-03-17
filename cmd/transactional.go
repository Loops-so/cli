package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

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

		client := api.NewClient(cfg.EndpointURL, cfg.APIKey)
		if err := client.SendTransactional(req); err != nil {
			return err
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
	transactionalSendCmd.MarkFlagRequired("email")
	transactionalSendCmd.MarkFlagRequired("id")
	transactionalCmd.AddCommand(transactionalSendCmd)

	rootCmd.AddCommand(transactionalCmd)
}
