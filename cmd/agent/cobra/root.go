package cmd

import (
	"os"

	"github.com/galogen13/yandex-go-metrics/internal/agent"
	"github.com/spf13/cobra"
)

var (
	hostAddress                  string
	pollInterval, reportInterval int
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "Сервис сбора метрик и алертинга: агент",

	Run: func(cmd *cobra.Command, args []string) {
		agent.Start(hostAddress, reportInterval, pollInterval)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&hostAddress, "a", "localhost:8080", "Host address")
	rootCmd.Flags().IntVar(&reportInterval, "r", 10, "Report interval, seconds")
	rootCmd.Flags().IntVar(&pollInterval, "p", 2, "Poll interval, seconds")
}
