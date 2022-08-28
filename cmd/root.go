package cmd

import (
	"github.com/chain710/manga/internal/flagvalues"
	"github.com/chain710/manga/internal/log"
	"go.uber.org/zap/zapcore"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	logLevel    = flagvalues.LogLevel{Level: zapcore.ErrorLevel}
	logEncoding = "json"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "manga",
	Short: "a manga web service",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viperBindPFlag(cmd.Flags())
		viperBindPFlag(cmd.PersistentFlags())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.manga.yaml)")
	rootCmd.PersistentFlags().VarP(&logLevel, "log-level", "L", "log level: error|warn|info|debug")
	rootCmd.PersistentFlags().StringVar(&logEncoding, "log-enc", "json", "log encoding: console|json")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	log.Init(log.WithLogLevel(logLevel.Level), log.WithLogEncoding(logEncoding))
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".manga" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".manga")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}
