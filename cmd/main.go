package cmd

import (
	"fmt"

	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agglomerator",
	Short: "Blockchain Agglomerator CLI",
	Long: `A blockchain agglomerator that enables cross-chain operations 
using vector spaces for optimal routing and state management.`,
}

func init() {
	// Add commands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(chainCmd)
	rootCmd.AddCommand(txCmd)

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "config.yaml", "config file path")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")
}

func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
