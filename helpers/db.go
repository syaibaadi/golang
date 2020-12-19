package helpers

import (
	"database/sql"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func GetDB(ctx Context, connection string) *gorm.DB {
	if connection == "" {
		connection = "central"
	}
	return ctx.Get("db_" + connection).(*gorm.DB)
}

func GetTxDB(ctx Context, connection string) *gorm.DB {
	if connection == "" {
		connection = "central"
	}
	return ctx.Get("tx_" + connection).(*gorm.DB)
}

func SetDBConfig(c echo.Context, m map[interface{}]interface{}) map[string]interface{} {
	var cfg map[string]interface{} = make(map[string]interface{})
	for i, v := range m {
		switch i.(string) {
		case "id":
			c.Set("company_id", v)
		case "status":
			c.Set("subscription_status", v)
		case "expired_date":
			c.Set("subscription_expired_date", v)
		default:
			if strings.Contains(i.(string), "db_") {
				ik := strings.Split(i.(string), "db_")
				switch v.(type) {
				case string:
					kc := strings.Trim(ik[1], " ")
					if kc == "username" || kc == "password" {
						dc, err := EncryptDecrypt(strings.Trim(v.(string), " "), false)
						if err == nil {
							cfg[kc] = dc
						} else {
							cfg[kc] = strings.Trim(v.(string), " ")
						}
					} else {
						cfg[kc] = strings.Trim(v.(string), " ")
					}
				default:
					cfg[strings.Trim(ik[1], " ")] = Convert(v).String()
				}
			}
		}
	}
	return cfg
}

func GetResults(rows *sql.Rows) []map[string]interface{} {
	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	length := len(columns)
	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		current := makeResultReceiver(length)
		if err := rows.Scan(current...); err != nil {
			panic(err)
		}
		value := make(map[string]interface{})
		for i := 0; i < length; i++ {
			value[columns[i]] = *(current[i]).(*interface{})
		}
		result = append(result, value)
	}
	return result
}

func GetResult(rows *sql.Rows) map[string]interface{} {
	result := make(map[string]interface{})
	res := GetResults(rows)
	if len(res) > 0 {
		result = res[0]
	}
	return result
}

func makeResultReceiver(length int) []interface{} {
	result := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		var current interface{}
		current = struct{}{}
		result = append(result, &current)
	}
	return result
}
