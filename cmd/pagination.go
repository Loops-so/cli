package cmd

import (
	"github.com/loops-so/cli/internal/api"
	"github.com/spf13/cobra"
)

func addPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().String("per-page", "", "Results per page (10-50, default 20)")
	cmd.Flags().String("cursor", "", "Pagination cursor for a specific page")
}

func paginationParams(cmd *cobra.Command) api.PaginationParams {
	perPage, _ := cmd.Flags().GetString("per-page")
	cursor, _ := cmd.Flags().GetString("cursor")
	return api.PaginationParams{
		PerPage: perPage,
		Cursor:  cursor,
	}
}
