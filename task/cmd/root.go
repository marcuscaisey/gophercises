package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/marcuscaisey/gophercises/task/repo"
)

type taskRepository interface {
	Add(task string) error
	ListIncomplete() ([]string, error)
	ListCompleted(since time.Time) ([]string, error)
	MarkAsComplete(taskNum int) (string, error)
	Remove(taskNum int) (string, error)
}

var rootCmd = &cobra.Command{
	Use:   "task [command]",
	Short: "task is a CLI for managing your TODOs.",
	Long: "task is a CLI for managing your TODOs.\n\n" +
		"The --file flag can be passed to customise where tasks are stored or you can set\n" +
		"the \"tasks_file\" key in the config file located at ~/.task/config.yaml.",
}

// Execute runs the task CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.EnableCommandSorting = false

	rootCmd.PersistentFlags().StringP("file", "f", "", fmt.Sprintf(`file used to store tasks (default "~/%s/%s")`, configDir, defaultDBFile))

	rootCmd.AddCommand(addCmd(initTaskRepo))
	rootCmd.AddCommand(listCmd(initTaskRepo))
	rootCmd.AddCommand(completedCmd(initTaskRepo))
	rootCmd.AddCommand(doCmd(initTaskRepo))
	rootCmd.AddCommand(rmCmd(initTaskRepo))
}

func initTaskRepo() (taskRepository, error) {
	dbFile := viper.GetString("tasks_file")
	if dbFile == "" {
		homeDir := getHomeDir()
		dbFile = path.Join(homeDir, configDir, defaultDBFile)
	}

	dbDir := path.Dir(dbFile)
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return nil, fmt.Errorf("Unable to create tasks DB directory: %s", err)
	}

	taskRepo, err := repo.NewBoltRepo(dbFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to initialise tasks DB: %s", err)
	}

	return taskRepo, nil
}

func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("Unable to determine home directory: %s", err))
	}
	return homeDir
}
