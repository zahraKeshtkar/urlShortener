package cmd

import (
	"github.com/spf13/cobra"

	"url-shortner/config"
	"url-shortner/log"
)

var cfg config.Config

func Execute() error {
	conf, err := config.Init()
	cfg = conf
	if err != nil {
		log.Errorf("can not run the command err is  %s", err)

		return err
	}

	var rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {},
	}
	RegisterServer(rootCmd, cfg)
	RegisterDatabase(rootCmd, cfg)
	if err = rootCmd.Execute(); err != nil {
		log.Errorf("can not run the command err is  %s", err)

		return err
	}

	return nil
}
