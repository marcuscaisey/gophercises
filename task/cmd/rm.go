package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/marcuscaisey/gophercises/task/repo"
)

func rmCmd(initTaskRepo func() (taskRepository, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "rm [flags] task-number",
		Short: "Remove an incomplete task",
		Long: "Remove an incomplete task\n\n" +
			"Note:\n" +
			"\"rm\" does not have the same effect as \"do\". Where \"do\" marks a task as\n " +
			"completed so that it shows up in the output of \"completed\" and not \"list\", \n" +
			"\"rm\" removes a task so that it shows up in neither.",
		Example: `task rm 2`,
		Args:    cobra.MatchAll(cobra.ExactArgs(1), taskNumIsPositiveInt),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRepo, err := initTaskRepo()
			if err != nil {
				return err
			}

			taskNum, _ := strconv.Atoi(args[0])

			task, err := taskRepo.Remove(taskNum)
			if err != nil {
				if errors.Is(err, repo.ErrTaskMissing) {
					return fmt.Errorf("Task %d does not exist", taskNum)
				}
				return fmt.Errorf("Unable to remove task %d: %s", taskNum, err)
			}

			fmt.Printf("You have removed the %q task.\n", task)
			return nil
		},
	}
}
