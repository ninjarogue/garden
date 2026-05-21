package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/app"
	"github.com/aric/garden/internal/output"
	"github.com/aric/garden/internal/retrieval"
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
		Short:         "Local context memory router for coding agents",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)
	root.AddCommand(
		newInitCommand(opts.App),
		newRememberCommand(opts.App),
		newPackCommand(opts.App),
		newAgentsCommand(opts.App),
		newListCommand(opts.App),
		newEditCommand(opts.App),
		newRemoveCommand(opts.App),
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
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Initialized .garden/memories.json")
			return err
		},
	}
}

func newRememberCommand(garden *app.App) *cobra.Command {
	var scopes []string
	var tags []string
	var always bool
	var priority string

	cmd := &cobra.Command{
		Use:   "remember <memory>",
		Short: "Store a scoped memory",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("memory cannot be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mem, err := garden.Remember(app.RememberInput{
				Memory:   args[0],
				Scope:    scopes,
				Always:   always,
				Tags:     tags,
				Priority: priority,
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Remembered %s\n", mem.ID)
			return err
		},
	}
	cmd.Flags().StringArrayVar(&scopes, "scope", nil, "repo-relative glob scope (repeatable)")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "memory tag (repeatable)")
	cmd.Flags().BoolVar(&always, "always", false, "consider this memory for any path")
	cmd.Flags().StringVar(&priority, "priority", "normal", "priority: low, normal, or high")
	return cmd
}

func newPackCommand(garden *app.App) *cobra.Command {
	var path string
	var task string
	var max int
	var maxChars int
	var explain bool

	cmd := &cobra.Command{
		Use:   "pack --path <path> --task <task>",
		Short: "Build a ranked context pack",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			results, err := garden.Pack(app.PackInput{Path: path, Task: task, Max: max, MaxChars: maxChars})
			if err != nil {
				return err
			}
			return output.WritePack(cmd.OutOrStdout(), path, task, results, explain)
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "repo-relative path for the task")
	cmd.Flags().StringVar(&task, "task", "", "task description")
	cmd.Flags().IntVar(&max, "max", retrieval.DefaultMax, "maximum memories to include")
	cmd.Flags().IntVar(&maxChars, "max-chars", retrieval.DefaultMaxChars, "maximum memory characters to include")
	cmd.Flags().BoolVar(&explain, "explain", false, "include memory selection reasons")
	return cmd
}

func newAgentsCommand(garden *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage Garden-owned AGENTS.md sections",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newAgentsUpdateCommand(garden),
		newAgentsSyncCommand(garden),
		newAgentsLintCommand(garden),
	)
	return cmd
}

func newAgentsUpdateCommand(garden *app.App) *cobra.Command {
	var input app.AgentsUpdateInput
	var maps []string
	var docs []string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Preview or apply Garden-managed AGENTS.md guidance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			entries, err := parseAgentEntries("--map", maps)
			if err != nil {
				return err
			}
			input.Map = entries
			entries, err = parseAgentEntries("--doc", docs)
			if err != nil {
				return err
			}
			input.Doc = entries
			change, err := garden.AgentsUpdate(input)
			if err != nil {
				return err
			}
			return output.WriteAgentsChange(cmd.OutOrStdout(), output.AgentsChange{
				Path:     change.Path,
				Before:   change.Before,
				After:    change.After,
				Applied:  change.Applied,
				Findings: change.Findings,
			}, "update")
		},
	}
	cmd.Flags().StringVar(&input.Purpose, "purpose", "", "project purpose")
	cmd.Flags().StringArrayVar(&input.Setup, "setup", nil, "setup command (repeatable)")
	cmd.Flags().StringArrayVar(&input.Build, "build", nil, "build command (repeatable)")
	cmd.Flags().StringArrayVar(&input.Lint, "lint", nil, "lint command (repeatable)")
	cmd.Flags().StringArrayVar(&input.Typecheck, "typecheck", nil, "typecheck command (repeatable)")
	cmd.Flags().StringArrayVar(&input.Test, "test", nil, "test command (repeatable)")
	cmd.Flags().StringArrayVar(&maps, "map", nil, "project structure entry as <path>: <purpose> (repeatable)")
	cmd.Flags().StringArrayVar(&input.Convention, "convention", nil, "repo convention (repeatable)")
	cmd.Flags().StringArrayVar(&docs, "doc", nil, "docs pointer as <path>: <reason> (repeatable)")
	cmd.Flags().StringArrayVar(&input.Note, "note", nil, "workflow note (repeatable)")
	cmd.Flags().BoolVar(&input.Apply, "apply", false, "write AGENTS.md instead of previewing")
	return cmd
}

func parseAgentEntries(flag string, values []string) ([]agents.Entry, error) {
	entries := make([]agents.Entry, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		entry, err := agents.ParseEntry(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", flag, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func newAgentsSyncCommand(garden *app.App) *cobra.Command {
	var input app.AgentsSyncInput
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Refresh generated Garden AGENTS.md index data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			change, err := garden.AgentsSync(input)
			if err != nil {
				return err
			}
			return output.WriteAgentsChange(cmd.OutOrStdout(), output.AgentsChange{
				Path:     change.Path,
				Before:   change.Before,
				After:    change.After,
				Applied:  change.Applied,
				Findings: change.Findings,
			}, "sync")
		},
	}
	cmd.Flags().BoolVar(&input.Apply, "apply", false, "write AGENTS.md instead of previewing")
	return cmd
}

func newAgentsLintCommand(garden *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Lint AGENTS.md shape, safety, and maintainability",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			findings, err := garden.AgentsLint()
			if err != nil {
				return err
			}
			failed, err := output.WriteAgentsLint(cmd.OutOrStdout(), findings)
			if err != nil {
				return err
			}
			if failed {
				return fmt.Errorf("AGENTS.md lint failed")
			}
			return nil
		},
	}
}

func newListCommand(garden *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored memories",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			memories, err := garden.List()
			if err != nil {
				return err
			}
			return output.WriteList(cmd.OutOrStdout(), memories)
		},
	}
}

func newEditCommand(garden *app.App) *cobra.Command {
	var memoryText string
	var scopes []string
	var tags []string
	var always bool
	var priority string

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a stored memory",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("memory id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mem, err := garden.Edit(app.EditInput{
				ID:          args[0],
				Memory:      memoryText,
				MemorySet:   cmd.Flags().Changed("memory"),
				Scope:       scopes,
				ScopeSet:    cmd.Flags().Changed("scope"),
				Always:      always,
				AlwaysSet:   cmd.Flags().Changed("always"),
				Tags:        tags,
				TagsSet:     cmd.Flags().Changed("tag"),
				Priority:    priority,
				PrioritySet: cmd.Flags().Changed("priority"),
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Edited %s\n", mem.ID)
			return err
		},
	}
	cmd.Flags().StringVar(&memoryText, "memory", "", "replacement memory text")
	cmd.Flags().StringArrayVar(&scopes, "scope", nil, "replacement repo-relative glob scope (repeatable)")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "replacement memory tag (repeatable)")
	cmd.Flags().BoolVar(&always, "always", false, "replace scope with always candidate")
	cmd.Flags().StringVar(&priority, "priority", "", "replacement priority: low, normal, or high")
	return cmd
}

func newRemoveCommand(garden *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a stored memory",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("memory id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := garden.Remove(args[0]); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", args[0])
			return err
		},
	}
}
