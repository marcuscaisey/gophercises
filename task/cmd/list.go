package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func listCmd(initTaskRepo func() (taskRepository, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list [flags]",
		Short: "List all of your incomplete tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRepo, err := initTaskRepo()
			if err != nil {
				return err
			}

			tasks, err := taskRepo.ListIncomplete()
			if err != nil {
				return fmt.Errorf("Unable to list incomplete tasks: %s", err)
			}

			if len(tasks) == 0 {
				fmt.Println("You have no incomplete tasks.")
				return nil
			}

			fmt.Println("You have the following tasks:")
			for i, task := range tasks {
				fmt.Printf("%d. %s\n", i+1, task)
			}

			return nil
		},
	}
}
