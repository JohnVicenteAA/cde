package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var attachName string

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to an existing cde tmux session",
	Args:  cobra.NoArgs,
	RunE:  runAttach,
}

func init() {
	attachCmd.Flags().StringVarP(&attachName, "name", "n", "", "name of session to attach")
	attachCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(attachCmd)
}

func runAttach(cmd *cobra.Command, args []string) error {
	if !runner.HasSession(attachName) {
		return fmt.Errorf("session %q not found", attachName)
	}
	return runner.Attach(attachName)
}
