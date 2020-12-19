package models

import (
	"gitlab.com/pangestu18/janji-online/chat/helpers"
)

type User struct {
	ID             int64
	Name           string
	StatusOnline   int
	StatusChatting int
	Signature      string
	StatusIM       bool `gorm:"column:status_im_ready;type:tinyint"`
}

type UserDetail struct {
	ID         int64 `gorm:"column:id_det_user"`
	IdUser     int64 `gorm:"column:iduser"`
	Name       string
	NickName   string `gorm:"column:nickname"`
	Gender     int8   `gorm:"column:gender"`
	SeekGender int8   `gorm:"column:s_gender"`
	Category   int8   `gorm:"column:kat_user"`
	Diamond    int64  `gorm:"column:total_diamond"`
	CreatedAt  string
	UpdatedAt  string
}

type Token struct {
	ID        string
	UserId    int64
	ExpiresAt string
	Revoke    bool `gorm:"type:tinyint"`
}

func ValidateUserByToken(ctx helpers.Context, token string, sign string) (User, bool) {
	var result = User{}
	var isValid bool = false

	helpers.GetDB(ctx, "").
		Table("users as u").
		Joins("join oauth_access_tokens as t on t.user_id = u.id").
		Where("t.id = ? ", token).
		Where("u.signature = ? ", sign).
		Where("revoked = 0").
		Select("u.*").
		Scan(&result)

	if result.ID != 0 {
		isValid = true
	}
	return result, isValid
}

func GetUserBySignature(ctx helpers.Context, sign string) User {
	user := User{}
	helpers.GetDB(ctx, "").Where("signature = ?", sign).First(&user)
	return user
}

func UpdateDiamond(ctx helpers.Context, user UserDetail) {
	db := helpers.GetDB(ctx, "")
	user.Diamond -= 15
	user.UpdatedAt = helpers.GetCurrentTime("Y-m-d")
	helpers.UpdateFromStruct(db, "tdetailuser", user, helpers.SetWrapping("", db.Dialect().GetName(), "id_det_user")+" = '"+helpers.Convert(user.ID).String()+"'")
}

func GetUserDetail(ctx helpers.Context, uid int64) UserDetail {
	var result = UserDetail{}

	helpers.GetDB(ctx, "").
		Table("tdetailuser as d").
		Where("d.iduser = ? ", uid).
		Select("d.*").
		Scan(&result)

	return result
}
