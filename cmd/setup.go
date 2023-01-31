package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"url-shortner/config"
	"url-shortner/db"
	"url-shortner/log"
	"url-shortner/repository"
)

func RegisterDatabase(root *cobra.Command, cfg config.Config) {
	log.InitLogger()
	root.AddCommand(&cobra.Command{
		Use:   "setupdb",
		Short: "Migration",
		RunE:  migrate,
	})
}

func migrate(cmd *cobra.Command, args []string) error {
	db, err := database.NewConnection(cfg.Database.Host,
		cfg.Database.Retry,
		10*time.Second,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port)
	if err != nil {
		return err
	}

	linkStore := repository.Link{
		DB: db,
	}
	err = linkStore.CreateTable()

	return err
}
