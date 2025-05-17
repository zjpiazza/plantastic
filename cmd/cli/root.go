package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	projectBase string
	userLicense string
	rootCmd     = &cobra.Command{
		Use:   "plantastic",
		Short: "Plantastic CLI for managing your garden",
		Long: `Plantastic is a CLI tool that helps you manage your garden,
including beds, plants, and tasks.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&projectBase, "api-url", "u", "", "Plantastic API URL")
	viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.SetDefault("api-url", "http://localhost:8080")

	apiUrl := viper.GetString("api-url")
	// Add subcommands
	rootCmd.AddCommand(bedsCmd(apiUrl))
	rootCmd.AddCommand(gardensCmd(apiUrl))
	rootCmd.AddCommand(tasksCmd(apiUrl))
}
