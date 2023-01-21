package server

import (
	"time"

	"github.com/spf13/cobra"

	"url-shortner/config"
	database "url-shortner/db"
	"url-shortner/handler"
)

func Register(root *cobra.Command, cfg config.Config) {
	root.AddCommand(
		&cobra.Command{
			Use:   "server",
			Short: "Run server",
			Run: func(cmd *cobra.Command, args []string) {
				db := database.DataBase{}
				db.NewConnection(cfg.Database.Host,
					cfg.Database.Retry,
					10*time.Second,
					cfg.Database.User,
					cfg.Database.Password,
					cfg.Database.DB,
					cfg.Database.Port)
				defer db.CloseDataBase()
				api := handler.API{
					DB: db,
				}
				api.Run()
			},
		},
	)
}
