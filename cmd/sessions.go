package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	listSessions bool
	deleteName   string
	deleteAll    bool
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage cde tmux sessions",
	Args:  cobra.NoArgs,
	RunE:  runSession,
}

var sessionDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete cde tmux sessions",
	Args:  cobra.NoArgs,
	RunE:  runSessionDelete,
}

func init() {
	sessionCmd.Flags().BoolVarP(&listSessions, "list", "l", false, "list all cde sessions")
	sessionDeleteCmd.Flags().StringVarP(&deleteName, "name", "n", "", "name of session to delete")
	sessionDeleteCmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "delete all cde sessions")
	sessionCmd.AddCommand(sessionDeleteCmd)
	rootCmd.AddCommand(sessionCmd)
}

func isCdeSession(name string) bool {
	return strings.HasSuffix(name, "_ide") ||
		strings.HasSuffix(name, "_wtree") ||
		strings.HasPrefix(name, "mrepo_")
}

var listCdeSessions = func() ([]string, error) {
	out, err := runner.Run("list-sessions", "-F", "#{session_name}")
	if err != nil {
		return nil, nil // no tmux server or no sessions
	}
	var sessions []string
	for _, name := range strings.Split(out, "\n") {
		name = strings.TrimSpace(name)
		if name != "" && isCdeSession(name) {
			sessions = append(sessions, name)
		}
	}
	return sessions, nil
}

func runSession(cmd *cobra.Command, args []string) error {
	if !listSessions {
		return cmd.Help()
	}

	sessions, err := listCdeSessions()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Println("No cde sessions found.")
		return nil
	}

	for _, s := range sessions {
		fmt.Println(s)
	}
	return nil
}

func runSessionDelete(cmd *cobra.Command, args []string) error {
	if !deleteAll && deleteName == "" {
		return fmt.Errorf("specify --name or --all")
	}

	if deleteAll {
		sessions, err := listCdeSessions()
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			fmt.Println("No cde sessions found.")
			return nil
		}
		for _, s := range sessions {
			if _, err := runner.Run("kill-session", "-t", "="+s); err != nil {
				return fmt.Errorf("failed to delete session %q: %w", s, err)
			}
			fmt.Printf("Deleted %s\n", s)
		}
		return nil
	}

	if !runner.HasSession(deleteName) {
		return fmt.Errorf("session %q not found", deleteName)
	}
	if _, err := runner.Run("kill-session", "-t", "="+deleteName); err != nil {
		return fmt.Errorf("failed to delete session %q: %w", deleteName, err)
	}
	fmt.Printf("Deleted %s\n", deleteName)
	return nil
}
