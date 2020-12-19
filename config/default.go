package config

import (
	"os"
)

var Default = map[string]Config{
	"APP_ENV":                  "production",
	"APP_PORT":                 "8899",
	"APP_URL":                  "https://go.zahironline.com",
	"APP_KEY":                  "wAGyTpFQX5uKV3JInABXXEdpgFkQLPTf",
	"IS_ON_PREMISE":            "true",
	"DB_MAX_OPEN_CONNS":        "100",
	"DB_MAX_IDLE_CONNS":        "2",
	"DB_CONN_MAX_LIFETIME":     "0ms",
	"DB_IS_DEBUG":              "false",
	"DB_SSL_MODE":              "disable",
	"MACHINE_TO_PORT":          "22",
	"REDIS_HOST":               "127.0.0.1",
	"REDIS_PORT":               "6379",
	"REDIS_PASSWORD":           "",
	"UPDATE_IM_ONLINE":         "30s",
	"REDIS_DB":                 "0",
	"REDIS_CACHE_DB":           "1",
	"LOGS_STORAGE_PATH":        "logs",
	"LOG_FILENAME":             "zo.log",
	"API_IS_DEBUG":             "false",
	"OAUTH2_ACCESS_EXPIRE_IN":  "1h",   // 1 hour
	"OAUTH2_REFRESH_EXPIRE_IN": "720h", // 30 days
	"GEOIP_URL":                "http://api.ipapi.com",
	"GEOIP_API_KEY":            "f653b82b6da754397846f4b2c453547f",
	"COURIER_URL":              "https://pro.rajaongkir.com",
	"COURIER_API_KEY":          "481c946ef3707d9e5776d7b5e60b747a",
	"MESSAGE_EXPIRED":          "30",
}

var TranslationMaps = map[string]string{
	"0": "trans.en.message",
	"1": "trans.en.validation",
	"2": "trans.id.validation",
	"3": "trans.id.message",
}

var DBDriverMaps = map[string]string{
	"firebird": "firebirdsql",
	"pgsql":    "postgres",
	"mysql":    "mysql",
	"mssql":    "mssql",
}

var SlackNotificationColors = map[string]string{
	"error":   "#A30200",
	"info":    "#2EB886",
	"warning": "#F8C64F",
	"debug":   "#1D9BD1",
	"default": "#DDDDDD",
}

var SlacksMap = map[string]interface{}{
	"websocket": map[string]interface{}{
		"development": Get("SLACK_WEBSOCKET_DEV_URL"),
		"production":  Get("SLACK_WEBSOCKET_URL"),
		"default":     Get("SLACK_ERROR_URL"),
	},
	"monitoring": map[string]interface{}{
		"development": Get("SLACK_MONITORING_DEV_URL"),
		"production":  Get("SLACK_MONITORING_URL"),
		"demo":        Get("SLACK_MONITORING_DEMO_URL"),
		"testing":     Get("SLACK_MONITORING_TESTING_URL"),
		"deadlock":    Get("SLACK_MONITORING_DEADLOCK_URL"),
		"onpremise":   Get("SLACK_MONITORING_ON_PREMISE_URL"),
		"default":     Get("SLACK_ERROR_URL"),
	},
	"other": map[string]interface{}{
		"error":   Get("SLACK_ERROR_URL"),
		"debug":   Get("SLACK_DEBUGING_URL"),
		"default": Get("SLACK_ERROR_URL"),
	},
}

var DBConnections = map[string]interface{}{
	"central": map[string]func(configs map[string]interface{}) Config{
		"driver": func(configs map[string]interface{}) Config {
			ret := GetFromMap(configs, "driver", "DB_DRIVER")
			if ret == "" {
				ret := GetFromMap(configs, "driver", "DB_CONNECTION")
				if ret == "" {
					ret = Config("mysql")
				}
			}
			return ret
		},
		"host": func(configs map[string]interface{}) Config {
			ret := GetFromMap(configs, "host", "DB_HOST")
			if ret == "" {
				ret = Config("localhost")
			}
			return ret
		},
		"port": func(configs map[string]interface{}) Config {
			ret := GetFromMap(configs, "port", "DB_PORT")
			if ret == "" {
				ret = Config("3306")
			}
			return ret
		},
		"database": func(configs map[string]interface{}) Config {
			ret := GetFromMap(configs, "database", "DB_DATABASE")
			if ret == "" {
				ret = Config("forge")
			}
			return ret
		},
		"username": func(configs map[string]interface{}) Config {
			ret := GetFromMap(configs, "username", "DB_USERNAME")
			if ret == "" {
				ret = Config("root")
			}
			return ret
		},
		"password": func(configs map[string]interface{}) Config {
			ret := GetFromMap(configs, "password", "DB_PASSWORD")
			if ret == "" {
				ret = Config("")
			}
			return ret
		},
		"ssl": func(configs map[string]interface{}) Config {
			ret := GetFromMap(configs, "ssl", "DB_SSL_MODE")
			return ret
		},
	},
}

func GetCurrentDir() string {
	path, _ := os.Getwd()
	return path
}
