package data

import (
	"realworld/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewMysqlDb, NewProfileRepo, NewUserRepo, NewArticleRepo, NewCommentRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(db *gorm.DB, c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{db: db}, cleanup, nil
}

func NewMysqlDb(c *conf.Data, logger log.Logger) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.Database.Dns), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	InitDb(db)
	return db
}

func InitDb(db *gorm.DB) {
	if err := db.AutoMigrate(
		&User{},
		&Follow{},
		&Article{},
		&Comment{},
		&ArticleFavorite{},
	); err != nil {
		panic(err)
	}
}
