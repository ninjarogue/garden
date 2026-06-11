package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/aric/garden/internal/app"
	"github.com/aric/garden/internal/output"
)

type Options struct {
	App    *app.App
	Stdout io.Writer
	Stderr io.Writer
}

func NewRoot(opts Options) *cobra.Command {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	root := &cobra.Command{
		Use:           "garden",
		Short:         "Maintain AGENTS.md and Markdown context cards for coding agents",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)
	root.AddCommand(
		newInitCommand(opts.App),
		newNewCommand(opts.App),
		newRemoveCommand(opts.App),
		newAgentsCommand(opts.App),
		newLintCommand(opts.App),
		newCheckCommand(opts.App),
	)
	return root
}

func newInitCommand(garden *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "init",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := garden.Init(); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Initialized .garden/context")
			return err
		},
	}
}

func newNewCommand(garden *app.App) *cobra.Command {
	var scopes []string
	var tags []string

	cmd := &cobra.Command{
		Use:   "new <slug>",
		Short: "Create a Markdown context card",
		Args:  exactlyOneCardSlug,
		RunE: func(cmd *cobra.Command, args []string) error {
			card, err := garden.NewCard(app.CreateCardInput{
				Slug:  args[0],
				Scope: scopes,
				Tags:  tags,
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", card.Path)
			return err
		},
	}
	cmd.Flags().StringArrayVar(&scopes, "scope", nil, "repo-relative glob scope (repeatable)")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "context tag (repeatable)")
	return cmd
}

func newRemoveCommand(garden *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <slug>",
		Short: "Remove a Markdown context card",
		Args:  exactlyOneCardSlug,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := garden.RemoveCard(args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", path)
			return err
		},
	}
}

func exactlyOneCardSlug(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("accepts exactly one context card slug")
	}
	return nil
}

func newAgentsCommand(garden *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage Garden-owned AGENTS.md sections",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			}
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAgentsSyncCommand(garden))
	return cmd
}

func newAgentsSyncCommand(garden *app.App) *cobra.Command {
	var input app.AgentsSyncInput
	var verbose bool
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Refresh generated Garden AGENTS.md context index",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			change, err := garden.AgentsSync(input)
			if err != nil {
				return err
			}
			return output.WriteAgentsChange(cmd.OutOrStdout(), change, "sync", verbose)
		},
	}
	cmd.Flags().BoolVar(&input.Apply, "apply", false, "write AGENTS.md instead of previewing")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show the generated diff when applying")
	return cmd
}

func newLintCommand(garden *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Lint Markdown context cards and AGENTS.md sync",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			findings, err := garden.Lint()
			if err != nil {
				return err
			}
			failed, err := output.WriteLint(cmd.OutOrStdout(), findings)
			if err != nil {
				return err
			}
			if failed {
				return fmt.Errorf("garden lint failed")
			}
			return nil
		},
	}
}

func newCheckCommand(garden *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "check <changed-path>...",
		Short: "Report review context for changed files",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one changed path is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input := app.CheckInput{ChangedPaths: args}
			report, err := garden.Check(input)
			if err != nil {
				return err
			}
			return output.WriteCheckReport(cmd.OutOrStdout(), report)
		},
	}
}
