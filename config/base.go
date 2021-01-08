package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config string

func Init(path string) {
	log.Println("Loading chat config for " + path)
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Loading .env file from /var/www/html/" + path + "/.env")
		err = godotenv.Load("/var/www/html/" + path + "/.env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func Get(key string) Config {
	value := Config(os.Getenv(key))
	if value == "" || value == "null" {
		value = Default[key]
	}
	return value
}

func GetDBConnection(configs map[string]interface{}, conn string, key string) Config {
	for i, m := range DBConnections {
		if i == conn {
			if mapf, ok := m.(map[string]func(map[string]interface{}) Config); ok {
				if f, ok := mapf[key]; ok {
					return f(configs)
				}
			}
		}
	}
	return Get(key)
}

func GetFromMap(configs map[string]interface{}, key string, altkey string) Config {
	if d, ok := configs[key]; ok {
		value := Config(d.(string))
		return value
	} else {
		return Get(altkey)
	}
}

func (c Config) String() string {
	return string(c)
}

func (c Config) Int() int {
	v, err := strconv.Atoi(c.String())
	if err != nil {
		return 0
	}
	return v
}

func (c Config) Bool() bool {
	if strings.ToLower(c.String()) == "true" {
		return true
	}
	return false
}

func (c Config) Duration() time.Duration {
	v, err := time.ParseDuration(c.String())
	if err != nil {
		return 0
	}
	return v
}
