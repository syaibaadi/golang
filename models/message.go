package models

import (
	"fmt"
	"math/rand"
	"strings"

	swearfilter "github.com/JoshuaDoes/gofuckyourself"
	"gitlab.com/pangestu18/janji-online/chat/config"
	"gitlab.com/pangestu18/janji-online/chat/helpers"
	"gitlab.com/pangestu18/janji-online/chat/helpers/phpserialize"
)

type Message struct {
	ID         string
	FromUser   int64  `gorm:"type:bigint"`
	ToUser     int64  `gorm:"type:bigint"`
	Message    string `gorm:"type:text"`
	MsgType    string
	Status     int8 `gorm:"type:tinyint"`
	ExpireDate string
	CreatedAt  string
	UpdatedAt  string
}

type BlockedWord struct {
	ID          int64
	Word        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

func DeleteOldMessages(ctx helpers.Context) {
	helpers.GetDB(ctx, "").
		Where("created_at <= ?", helpers.SubDateTime("day", "now", config.Get("MESSAGE_EXPIRED").Int(), "Y-m-d H:i:s")).
		Delete(&Message{})
}

func CreateMessage(ctx helpers.Context, from, to int64, message string) (string, error) {
	if message != "" {
		id := helpers.NewUUID()
		now := helpers.GetCurrentTime("Y-m-d H:i:s")
		exp := helpers.AddDateTime("minute", now, 2, "Y-m-d H:i:s")
		fmt.Println(exp)
		message = FilterBadWords(ctx, message)
		msg := Message{ID: id, FromUser: from, ToUser: to, Message: message, Status: 1, CreatedAt: now, UpdatedAt: now, ExpireDate: exp, MsgType: "chat"}
		result := helpers.GetDB(ctx, "").Create(&msg)
		return message, result.Error
	} else {
		return message, nil
	}
}

func FilterBadWords(ctx helpers.Context, message string) string {
	badwords := helpers.GetCacheValue(ctx, "badwords")
	var swears []string = make([]string, 0)
	if len(badwords.Map()) < 0 {
		for _, b := range badwords.Map() {
			swears = append(swears, b.(string))
		}
	} else {
		censor := []BlockedWord{}
		helpers.GetDB(ctx, "").Find(&censor)
		var ret phpserialize.PhpSlice
		for _, msg := range censor {
			ret = append(ret, msg.Word)
			swears = append(swears, msg.Word)
		}
		helpers.SetCacheValue(ctx, "badwords", ret, 0)
	}
	filter := swearfilter.New(false, false, false, false, false, swears...)
	swearFound, swearsFound, err := filter.Check(message)
	if err == nil && swearFound {
		for _, x := range swearsFound {
			if strings.Trim(x, " ") != "" {
				fmt.Println("Replace ", x)
				message = strings.ReplaceAll(message, x, RandomCensor(x, '*', len(x)))
			}
		}
	}
	return message
}

func RandomCensor(word string, ch rune, n int) string {
	if len(word) > 0 {
		l := n / len(word)
		for i := 0; i < l+2; i++ {
			r := rand.Intn(len(word) - 1)
			word = helpers.ReplaceAtIndex(word, ch, r)
		}
	}
	return word
}
