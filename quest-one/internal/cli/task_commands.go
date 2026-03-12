package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/domain"
)

func nextCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Show the highest-priority task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			task, err := app.NextTask(cmd.Context())
			if err != nil {
				return err
			}
			if task == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "No active tasks.")
				return nil
			}
			printTask(cmd, *task)
			return nil
		},
	}
}

func addCmd(app *application.App) *cobra.Command {
	var (
		desc    string
		tags    []string
		dueDate string
	)

	c := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := application.AddTaskInput{
				Title:       strings.Join(args, " "),
				Description: desc,
				Tags:        tags,
			}
			if dueDate != "" {
				t, err := time.Parse("2006-01-02", dueDate)
				if err != nil {
					return fmt.Errorf("invalid due date (use YYYY-MM-DD): %w", err)
				}
				in.DueDate = &t
			}
			task, err := app.AddTask(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added: %s (%s)\n", task.Title, task.ID)
			return nil
		},
	}
	c.Flags().StringVarP(&desc, "description", "d", "", "Task description")
	c.Flags().StringSliceVarP(&tags, "tags", "t", nil, "Comma-separated tags")
	c.Flags().StringVar(&dueDate, "due", "", "Due date (YYYY-MM-DD)")
	return c
}

func listCmd(app *application.App) *cobra.Command {
	var limit int

	c := &cobra.Command{
		Use:   "list",
		Short: "List active tasks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out, err := app.ListTasks(cmd.Context(), application.ListTasksInput{
				Statuses: []domain.TaskStatus{domain.TaskStatusTodo, domain.TaskStatusInProgress, domain.TaskStatusWaiting},
				Limit:    limit,
			})
			if err != nil {
				return err
			}
			if len(out.Tasks) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active tasks.")
				return nil
			}
			printTaskTable(cmd, out.Tasks)
			return nil
		},
	}
	c.Flags().IntVarP(&limit, "limit", "n", 20, "Maximum number of tasks to show")
	return c
}

func completeCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark a task as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := app.CompleteTask(cmd.Context(), domain.TaskID(args[0]))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Completed: %s\n", task.Title)
			return nil
		},
	}
}

func cancelCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "Cancel a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := app.CancelTask(cmd.Context(), domain.TaskID(args[0]))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Cancelled: %s\n", task.Title)
			return nil
		},
	}
}

func promoteCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "promote <id>",
		Short: "Increase a task's urgency by one level",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := app.PromoteTask(cmd.Context(), domain.TaskID(args[0]))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Promoted: %s (urgency=%d)\n", task.Title, task.Priority.ManualUrgency)
			return nil
		},
	}
}

func memoCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "memo <id> <text>",
		Short: "Set the memo for a task",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := strings.Join(args[1:], " ")
			_, err := app.AddMemo(cmd.Context(), domain.TaskID(args[0]), text)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Memo saved.")
			return nil
		},
	}
}

func candidatesCmd(app *application.App) *cobra.Command {
	var n int
	c := &cobra.Command{
		Use:   "candidates",
		Short: "Show the top N priority tasks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			tasks, err := app.Candidates(cmd.Context(), n)
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active tasks.")
				return nil
			}
			printTaskTable(cmd, tasks)
			return nil
		},
	}
	c.Flags().IntVarP(&n, "num", "n", 5, "Number of candidates")
	return c
}

func searchCmd(app *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search over tasks",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := strings.Join(args, " ")
			results, err := app.SearchTasks(cmd.Context(), q, 20)
			if err != nil {
				return err
			}
			tasks := make([]domain.Task, len(results))
			for i, r := range results {
				tasks[i] = r.Task
			}
			printTaskTable(cmd, tasks)
			return nil
		},
	}
}

// ---- print helpers ----

func printTask(cmd *cobra.Command, t domain.Task) {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "ID:     %s\n", t.ID)
	fmt.Fprintf(w, "Title:  %s\n", t.Title)
	fmt.Fprintf(w, "Status: %s\n", t.Status)
	if t.Description != "" {
		fmt.Fprintf(w, "Desc:   %s\n", t.Description)
	}
	if t.DueDate != nil {
		fmt.Fprintf(w, "Due:    %s\n", t.DueDate.Format("2006-01-02"))
	}
	if t.Memo != "" {
		fmt.Fprintf(w, "Memo:   %s\n", t.Memo)
	}
}

func printTaskTable(cmd *cobra.Command, tasks []domain.Task) {
	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATUS\tURGENCY\tTITLE")
	for _, t := range tasks {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			t.ID, t.Status, t.Priority.ManualUrgency, t.Title)
	}
	_ = tw.Flush()
	fmt.Fprintf(cmd.OutOrStdout(), "\n%d task(s)\n", len(tasks))
}

func logsCmd(_ *application.App) *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Show recent application log entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// TODO: tail the log file path from settings
			fmt.Fprintln(cmd.OutOrStdout(), "Log tailing not yet implemented. Check your data directory.")
			return nil
		},
	}
}
