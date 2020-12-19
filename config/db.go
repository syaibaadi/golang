package config

import "fmt"

type DBConfig struct {
	Connection string
	DSN        string
	DBType     string
}

func CreateDSN(configs map[string]interface{}, conn string) string {
	host := GetDBConnection(configs, conn, "host").String()
	port := GetDBConnection(configs, conn, "port").Int()
	name := GetDBConnection(configs, conn, "database").String()
	user := GetDBConnection(configs, conn, "username").String()
	password := GetDBConnection(configs, conn, "password").String()
	driver := GetDBConnection(configs, conn, "driver").String()
	var dsn string
	driver = DBDriverMaps[driver]
	if driver == "mysql" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, name)
	} else if driver == "firebirdsql" {
		dsn = fmt.Sprintf("%s:%s@%s:%d/%s", user, password, host, port, name)
	} else if driver == "postgres" {
		ssl := GetDBConnection(configs, conn, "ssl").String()
		if ssl == "" && Get("APP_ENV") == "local" {
			ssl = "disable"
		}
		if ssl != "" {
			dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
				host, port, user, password, name, ssl)
		} else {
			dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
				host, port, user, password, name)
		}
	}
	return dsn
}

func GetDBConfig(conn string, configs map[string]interface{}) DBConfig {
	var dbconf DBConfig
	driver := GetDBConnection(configs, conn, "driver").String()
	dbconf.DBType = DBDriverMaps[driver]
	dbconf.DSN = CreateDSN(configs, conn)
	return dbconf
}
