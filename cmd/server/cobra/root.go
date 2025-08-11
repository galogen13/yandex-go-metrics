package cmd

import (
	"os"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/router"
	"github.com/galogen13/yandex-go-metrics/internal/storage"
	"github.com/spf13/cobra"
)

var hostAddress string = ""

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Сервис сбора метрик и алертинга: сервер",

	Run: func(cmd *cobra.Command, args []string) {
		var storage models.Storage = storage.NewMemStorage()

		if err := router.Start(hostAddress, storage); err != nil {
			panic(err)
		}
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
}
