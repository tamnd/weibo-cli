package cli

import (
	"github.com/spf13/cobra"
)

// hotCmd returns the hot command.
func (a *App) hotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hot",
		Short: "List Weibo hot search topics (微博热搜榜)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(30)
			a.progressf("fetching %d hot search topics...", n)
			items, err := a.client.HotSearch(cmd.Context(), n)
			if err != nil {
				return codeError(exitError, err)
			}
			return a.renderOrEmpty(items, len(items))
		},
	}
}
