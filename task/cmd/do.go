package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/marcuscaisey/gophercises/task/repo"
)

func doCmd(initTaskRepo func() (taskRepository, error)) *cobra.Command {
	return &cobra.Command{
		Use:     "do [flags] task-number",
		Short:   "Mark a task on your TODO list as complete",
		Example: `task do 2`,
		Args:    cobra.MatchAll(cobra.ExactArgs(1), taskNumIsPositiveInt),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRepo, err := initTaskRepo()
			if err != nil {
				return err
			}

			taskNum, _ := strconv.Atoi(args[0])

			task, err := taskRepo.MarkAsComplete(taskNum)
			if err != nil {
				if errors.Is(err, repo.ErrTaskMissing) {
					return fmt.Errorf("task %d does not exist.", taskNum)
				}
				return fmt.Errorf("Unable to mark task %d as completed: %s", taskNum, err)
			}

			fmt.Printf("You have completed the %q task.\n", task)
			return nil
		},
	}
}

func taskNumIsPositiveInt(cmd *cobra.Command, args []string) error {
	taskNum, err := strconv.Atoi(args[0])
	if err != nil || taskNum < 1 {
		return errors.New("task-number must be a positive integer")
	}
	return nil
}
