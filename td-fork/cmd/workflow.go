package cmd

import (
	"fmt"
	"strings"

	"github.com/marcus/td/internal/models"
	"github.com/marcus/td/internal/workflow"
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Show issue status workflow",
	Long: `Displays the issue status workflow state machine.

Shows all valid status transitions and any guards applied.`,
	GroupID: "system",
	RunE: func(cmd *cobra.Command, args []string) error {
		showMermaid, _ := cmd.Flags().GetBool("mermaid")
		showDot, _ := cmd.Flags().GetBool("dot")

		if showMermaid {
			return printMermaidDiagram()
		}
		if showDot {
			return printDotDiagram()
		}
		return printWorkflow()
	},
}

func printWorkflow() error {
	sm := workflow.DefaultMachine()

	fmt.Println("ISSUE STATUS WORKFLOW")
	fmt.Println("=====================")
	fmt.Println()

	// Show statuses
	fmt.Println("STATUSES:")
	for _, s := range workflow.AllStatuses() {
		fmt.Printf("  • %s\n", s)
	}
	fmt.Println()

	// Show transitions by source
	fmt.Println("TRANSITIONS:")
	for _, from := range workflow.AllStatuses() {
		allowed := sm.GetAllowedTransitions(from)
		if len(allowed) > 0 {
			fmt.Printf("  %s →\n", from)
			for _, to := range allowed {
				name := workflow.TransitionName(from, to)
				t := sm.GetTransition(from, to)
				guardStr := ""
				if t != nil && len(t.Guards) > 0 {
					var guardNames []string
					for _, g := range t.Guards {
						guardNames = append(guardNames, g.Name())
					}
					guardStr = fmt.Sprintf(" [%s]", strings.Join(guardNames, ", "))
				}
				fmt.Printf("    %s (%s)%s\n", to, name, guardStr)
			}
		}
	}
	fmt.Println()

	// Show workflow modes
	fmt.Println("MODES:")
	fmt.Println("  • Liberal  - Guards disabled (default)")
	fmt.Println("  • Advisory - Guards warn but allow")
	fmt.Println("  • Strict   - Guards block transitions")
	fmt.Println()

	// Show guards
	fmt.Println("GUARDS (applied in Advisory/Strict modes):")
	fmt.Println("  • BlockedGuard          - Requires --force to start blocked issues")
	fmt.Println("  • DifferentReviewerGuard - Prevents self-approval (except minor tasks)")
	fmt.Println()

	return nil
}

func printMermaidDiagram() error {
	sm := workflow.DefaultMachine()

	fmt.Println("```mermaid")
	fmt.Println("stateDiagram-v2")

	// Show transitions
	for _, from := range workflow.AllStatuses() {
		for _, to := range sm.GetAllowedTransitions(from) {
			name := workflow.TransitionName(from, to)
			fmt.Printf("    %s --> %s: %s\n", from, to, name)
		}
	}

	fmt.Println("```")
	return nil
}

func printDotDiagram() error {
	sm := workflow.DefaultMachine()

	fmt.Println("digraph workflow {")
	fmt.Println("    rankdir=LR;")
	fmt.Println("    node [shape=box];")
	fmt.Println()

	// Node styling
	fmt.Printf("    %s [style=filled,fillcolor=lightblue];\n", models.StatusOpen)
	fmt.Printf("    %s [style=filled,fillcolor=lightyellow];\n", models.StatusInProgress)
	fmt.Printf("    %s [style=filled,fillcolor=lightpink];\n", models.StatusBlocked)
	fmt.Printf("    %s [style=filled,fillcolor=lightorange];\n", models.StatusInReview)
	fmt.Printf("    %s [style=filled,fillcolor=lightgreen];\n", models.StatusClosed)
	fmt.Println()

	// Transitions
	for _, from := range workflow.AllStatuses() {
		for _, to := range sm.GetAllowedTransitions(from) {
			name := workflow.TransitionName(from, to)
			fmt.Printf("    %s -> %s [label=\"%s\"];\n", from, to, name)
		}
	}

	fmt.Println("}")
	return nil
}

func init() {
	rootCmd.AddCommand(workflowCmd)

	workflowCmd.Flags().Bool("mermaid", false, "Output Mermaid diagram")
	workflowCmd.Flags().Bool("dot", false, "Output GraphViz DOT diagram")
}
