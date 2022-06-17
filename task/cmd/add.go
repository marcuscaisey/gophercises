package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func addCmd(initTaskRepo func() (taskRepository, error)) *cobra.Command {
	return &cobra.Command{
		Use:     "add [flags] task",
		Short:   "Add a new task to your TODO list",
		Example: `task add "talk proposal"`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRepo, err := initTaskRepo()
			if err != nil {
				return err
			}

			task := strings.Join(args, " ")

			if err := taskRepo.Add(task); err != nil {
				return fmt.Errorf("Unable to add task %q: %s", task, err)
			}

			fmt.Printf("Added %q to your task list.\n", task)
			return nil
		},
	}
}
