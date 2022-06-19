package cmd

import (
	"errors"
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const configDir = ".task"
const defaultDBFile = "tasks.db"

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	homeDir := getHomeDir()
	viper.AddConfigPath(path.Join(homeDir, configDir))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.BindPFlag("tasks_file", rootCmd.PersistentFlags().Lookup("file")); err != nil {
		cobra.CheckErr(fmt.Errorf("Unable to bind tasks_file config to --file flag: %s", err))
	}

	if err := viper.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			cobra.CheckErr(fmt.Errorf("Unable to read config: %s", err))
		}
	}

	viper.AutomaticEnv()
}
