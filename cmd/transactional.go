package cmd

import (
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

func init() {
	addPaginationFlags(transactionalListCmd)
	transactionalCmd.AddCommand(transactionalListCmd)
	rootCmd.AddCommand(transactionalCmd)
}
