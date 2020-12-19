package db

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"gitlab.com/pangestu18/janji-online/chat/config"
)

var db *gorm.DB

func Connect(connection string, configs map[string]interface{}) *gorm.DB {
	dbconf := config.GetDBConfig(connection, configs)
	dbConnPool, err := gorm.Open(dbconf.DBType, dbconf.DSN)

	if err != nil {
		log.Fatal(err)
	}

	if err := dbConnPool.DB().Ping(); err != nil {
		log.Fatal(err)
	}

	if config.Get("DB_IS_DEBUG").Bool() {
		dbConnPool = dbConnPool.Debug()
	}

	maxOpenConns := config.Get("DB_MAX_OPEN_CONNS").Int()
	maxIdleConns := config.Get("DB_MAX_IDLE_CONNS").Int()
	connMaxLifetime := config.Get("DB_CONN_MAX_LIFETIME").Duration()

	dbConnPool.DB().SetMaxIdleConns(maxIdleConns)
	dbConnPool.DB().SetMaxOpenConns(maxOpenConns)
	dbConnPool.DB().SetConnMaxLifetime(connMaxLifetime)

	db = dbConnPool
	return db
}
