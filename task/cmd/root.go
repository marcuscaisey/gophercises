package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/marcuscaisey/gophercises/task/repo"
)

const configDir = ".task"
const defaultDBFile = "tasks.db"

type taskRepository interface {
	Add(task string) error
	MarkAsComplete(index int) (string, error)
	ListIncomplete() ([]string, error)
}

var rootCmd = &cobra.Command{
	Use:   "task [command]",
	Short: "task is a CLI for managing your TODOs.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.EnableCommandSorting = false
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("file", "f", "", fmt.Sprintf(`file used to store tasks (default "~/%s/%s")`, configDir, defaultDBFile))

	rootCmd.AddCommand(addCmd(initTaskRepo))
	rootCmd.AddCommand(listCmd(initTaskRepo))
	rootCmd.AddCommand(doCmd(initTaskRepo))
}

func initConfig() {
	homeDir := getHomeDir()
	viper.AddConfigPath(path.Join(homeDir, configDir))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.BindPFlag("tasks_file", rootCmd.PersistentFlags().Lookup("file"))

	if err := viper.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			cobra.CheckErr(fmt.Errorf("Unable to read config: %s", err))
		}
	}

	viper.AutomaticEnv()
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
