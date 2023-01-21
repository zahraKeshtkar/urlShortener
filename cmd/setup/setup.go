package setupdb

import (
	"time"

	"github.com/spf13/cobra"

	"url-shortner/config"
	"url-shortner/db"
	"url-shortner/log"
)

func Register(root *cobra.Command, cfg config.Config) {
	log.InitLogger()
	root.AddCommand(&cobra.Command{
		Use:   "setupdb",
		Short: "Migration",
		Run: func(cmd *cobra.Command, args []string) {
			db := database.DataBase{}
			db.NewConnection(cfg.Database.Host,
				cfg.Database.Retry,
				10*time.Second,
				cfg.Database.User,
				cfg.Database.Password,
				cfg.Database.DB,
				cfg.Database.Port)
			db.CreateTable()
		},
	})
}
