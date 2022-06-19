package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func completedCmd(initTaskRepo func() (taskRepository, error)) *cobra.Command {
	const sinceLayout = "2006-01-02 15:04"
	var since time.Time

	cmd := &cobra.Command{
		Use:   "completed [flags]",
		Short: "List all of your completed tasks since a given date and time.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRepo, err := initTaskRepo()
			if err != nil {
				return err
			}

			tasks, err := taskRepo.ListCompleted(since)
			if err != nil {
				return fmt.Errorf("Unable to list completed tasks: %s", err)
			}

			formattedSince := since.Format(sinceLayout)
			if len(tasks) == 0 {
				fmt.Printf("You have no completed tasks since %s.\n", formattedSince)
				return nil
			}

			fmt.Printf("You have completed the following tasks since %s:\n", formattedSince)
			for _, task := range tasks {
				fmt.Printf("- %s\n", task)
			}

			return nil
		},
	}

	sinceUsage := fmt.Sprintf("List tasks completed since this date and time [format: %s]", sinceLayout)
	cmd.Flags().VarP(newSinceValue(&since, sinceLayout, midnight()), "since", "s", sinceUsage)

	cmd.Flags().SortFlags = false

	return cmd
}

func midnight() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

type sinceValue struct {
	p      *time.Time
	layout string
}

func newSinceValue(p *time.Time, layout string, value time.Time) sinceValue {
	*p = value
	return sinceValue{
		p:      p,
		layout: layout,
	}
}

func (v sinceValue) Set(s string) error {
	t, err := time.ParseInLocation(v.layout, s, time.Local)
	if err != nil {
		return err
	}
	*v.p = t
	return nil
}

func (v sinceValue) String() string {
	return strconv.Quote(v.p.Format(v.layout))
}

func (v sinceValue) Type() string {
	return "time.Time"
}
