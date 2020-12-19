package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"gitlab.com/pangestu18/janji-online/chat/config"
	"gitlab.com/pangestu18/janji-online/chat/db"
	"gitlab.com/pangestu18/janji-online/chat/helpers"
	"gitlab.com/pangestu18/janji-online/chat/helpers/phpserialize"
	"gitlab.com/pangestu18/janji-online/chat/models"
	"gitlab.com/pangestu18/janji-online/chat/routes"
)

var (
	syncOnce sync.Once
	dbc      *gorm.DB
)

func main() {
	arg := "janji-online"
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	config.Init(arg)

	syncOnce.Do(DBConnection)
	defer dbc.Close()

	cache := helpers.CacheConnection()
	DoCacheTranslation(dbc, cache)
	r := routes.Init(dbc, cache)
	err := r.Start(":" + config.Get("APP_PORT").String())
	if err != nil {
		r.Logger.Fatal(err)
	}
}

func DBConnection() {
	go helpers.ActiveHub.Run()
	dbc = db.Connect("central", map[string]interface{}{})
}

func DoCacheTranslation(db *gorm.DB, c *redis.Client) {
	censor := []models.BlockedWord{}
	db.Find(&censor)
	var ret phpserialize.PhpSlice
	for _, msg := range censor {
		ret = append(ret, msg.Word)
	}
	key := "badwords"
	cacheKey := config.Get("GO_CACHE_PREFIX").String() + ":" + key
	v, err := helpers.PhpSerialize(ret)
	if err == nil {
		c.Set(cacheKey, v, 0)
	} else {
		fmt.Println(err)
	}
}
