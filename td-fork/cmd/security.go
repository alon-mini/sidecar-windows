package cmd

import (
	"fmt"
	"time"

	"github.com/marcus/td/internal/db"
	"github.com/marcus/td/internal/output"
	"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
	Use:     "security",
	Short:   "View security exception log (self-close exceptions)",
	Long:    `Shows audit log of issues closed using --self-close-exception.`, 
	GroupID: "system",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := getBaseDir()

		clearFlag, _ := cmd.Flags().GetBool("clear")
		if clearFlag {
			if err := db.ClearSecurityEvents(baseDir); err != nil {
				output.Error("failed to clear security events: %v", err)
				return err
			}
			fmt.Println("Cleared security exception log")
			return nil
		}

		events, err := db.ReadSecurityEvents(baseDir)
		if err != nil {
			output.Error("failed to read security events: %v", err)
			return err
		}

		if len(events) == 0 {
			fmt.Println("No security exceptions logged")
			return nil
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			for _, e := range events {
				fmt.Printf(`{"ts":"%s","issue_id":"%s","session_id":"%s","agent_type":"%s","reason":"%s"}`+"\n",
					e.Timestamp.Format(time.RFC3339),
					e.IssueID,
					e.SessionID,
					e.AgentType,
					escapeJSON(e.Reason))
			}
			return nil
		}

		// Human-readable output
		fmt.Printf("Security Exceptions (%d):\n\n", len(events))
		for _, e := range events {
			ts := e.Timestamp.Local().Format("2006-01-02 15:04:05")
			agent := e.AgentType
			if agent == "" {
				agent = "unknown"
			}

			fmt.Printf("%s  %s (Agent: %s)\n", ts, e.IssueID, agent)
			fmt.Printf("  Reason: %s\n", e.Reason)
			if e.SessionID != "" {
				fmt.Printf("  Session: %s\n", e.SessionID)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(securityCmd)

	securityCmd.Flags().Bool("clear", false, "Clear the security log")
	securityCmd.Flags().Bool("json", false, "Output as JSONL")
}
