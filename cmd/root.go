package cmd

import (
	"github.com/spf13/cobra"

	"url-shortner/cmd/server"
	setupdb "url-shortner/cmd/setup"
	"url-shortner/config"
	"url-shortner/log"
)

func Execute() {
	cfg := config.Init()
	var rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {},
	}
	server.Register(rootCmd, cfg)
	setupdb.Register(rootCmd, cfg)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(" %s", err)
	}

}
