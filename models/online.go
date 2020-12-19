package models

import (
	"math"

	"gitlab.com/pangestu18/janji-online/chat/helpers"
)

type Member struct {
	ID            int64  `gorm:"column:id_det_user" json:"id"`
	IdUser        int64  `gorm:"column:iduser" json:"id_user"`
	Name          string `json:"name"`
	NickName      string `gorm:"column:nickname" json:"nick_name"`
	Photo         string `gorm:"column:foto" json:"photo"`
	UnreadMessage int64  `gorm:"column:total_msg" json:"unread_message"`
	Signature     string `json:"signature"`
}

type Privilege struct {
	ID       int64  `gorm:"column:id_prev" json:"id"`
	Name     string `gorm:"column:nama" json:"name"`
	Category int8   `gorm:"column:kategori" json:"category"`
	FreeChat int64  `gorm:"column:free_chat" json:"free_chat"`
}

type FreeChat struct {
	ID         string `gorm:"column:id" json:"id"`
	IdUser     int64  `gorm:"column:id_user" json:"id_user"`
	IdMember   int64  `gorm:"column:id_member" json:"id_member"`
	AccessDate string `gorm:"column:access_date" json:"access_date"`
}

func GetOnlineMembers(ctx helpers.Context, user UserDetail, limit int64) []Member {
	var OL []Member = []Member{}
	helpers.GetDB(ctx, "").
		Raw("Select d.iduser,d.id_det_user,d.nickname,count(m.id) total_msg, ft.foto,u.signature "+
			"from tdetailuser as d join users as u on u.id = d.iduser "+
			"left join messages as m on m.from_user = u.id and m.msg_type = 'chat' and m.status = 1 "+
			"and m.expire_date >= convert_tz(now(),@@session.time_zone,'+07:00') "+
			"join tfoto as ft on ft.id_detail = d.id_det_user and ft.image_as = 1 "+
			"where d.kat_user <= ? AND d.stats_approved = 1  AND u.status_chatting = 1 AND u.status_online = 1 "+
			"AND d.id_det_user != ? AND d.gender = ? "+
			"group by d.iduser,d.id_det_user,d.nickname, ft.foto,u.signature order by d.nickname", user.Category, user.ID, user.SeekGender).
		Scan(&OL)
	return OL
}

func GetOnlineFriends(ctx helpers.Context, user UserDetail, limit int64) []Member {
	var OL []Member = []Member{}
	helpers.GetDB(ctx, "").
		Raw("Select d.iduser,d.id_det_user,d.nickname,count(m.id) total_msg, ft.foto,u.signature "+
			"from tdetailuser as d join users as u on u.id = d.iduser "+
			"join tfriends as f on u.id = f.id_friend and f.status = 2 and f.id_user = ? "+
			"left join messages as m on m.from_user = u.id and m.msg_type = 'chat' and m.status = 1 "+
			"and m.expire_date >= convert_tz(now(),@@session.time_zone,'+07:00') "+
			"join tfoto as ft on ft.id_detail = d.id_det_user and ft.image_as = 1 "+
			"where d.kat_user <= ? AND d.stats_approved = 1  AND u.status_chatting = 1 AND u.status_online = 1 "+
			"AND d.id_det_user != ? AND d.gender = ? "+
			"group by d.iduser,d.id_det_user,d.nickname, ft.foto,u.signature order by d.nickname", user.IdUser, user.Category, user.ID, user.SeekGender).
		Scan(&OL)
	return OL
}

func GetTotalOnlineMembers(ctx helpers.Context, user UserDetail) int64 {
	var total int64 = 0
	helpers.GetDB(ctx, "").
		Raw("Select count(d.id_det_user) total from tdetailuser as d join users as u on u.id = d.iduser "+
			"join tfoto as ft on ft.id_detail = d.id_det_user and ft.image_as = 1 "+
			"where d.kat_user <= ? AND d.stats_approved = 1  AND u.status_chatting = 1 AND u.status_online = 1 "+
			"AND d.id_det_user != ? AND d.gender = ? ", user.Category, user.ID, user.SeekGender).
		Count(&total)
	return total
}

func GetTotalOnlineFriends(ctx helpers.Context, user UserDetail) int64 {
	var total int64 = 0
	helpers.GetDB(ctx, "").
		Raw("Select count(distinct d.id_det_user) total from tdetailuser as d join users as u on u.id = d.iduser "+
			"join tfoto as ft on ft.id_detail = d.id_det_user and ft.image_as = 1 "+
			"join tfriends as f on u.id = f.id_friend and f.id_user = ?  and f.status = 2 where d.stats_approved = 1 AND u.status_chatting = 1 "+
			"AND d.gender =  ? ", user.IdUser, user.SeekGender).
		Count(&total)
	return total
}

func GetFreeChatByCategory(ctx helpers.Context, category int8) int64 {
	var total int64 = 0
	helpers.GetDB(ctx, "").
		Raw("Select free_chat from tprevilege where kategori = ? ", category).
		Count(&total)
	return total
}

func GetFreeChatUsedToday(ctx helpers.Context, sender, receiver int64) int64 {
	var total int64 = 0
	helpers.GetDB(ctx, "").
		Raw("Select count(id) total from freechats where id_user = ? and id_member != ? and access_date = ? ", sender, receiver, helpers.GetCurrentTime("Y-m-d")).
		Count(&total)
	return total
}

func HasFreeChat(ctx helpers.Context, user UserDetail, receiver User) (bool, int64) {
	var has bool = false
	var free int64 = GetFreeChatByCategory(ctx, user.Category)
	if free == 0 {
		has = true
	} else {
		var used int64 = GetFreeChatUsedToday(ctx, user.IdUser, receiver.ID)
		if free > used {
			has = true
			free -= used
		}
	}
	return has, free
}

func InsertFreeChatHistoryIfNotExist(ctx helpers.Context, sender, receiver int64) {
	db := helpers.GetDB(ctx, "")
	fc := FreeChat{}
	db.Table("freechats").
		Where("id_user = ? and id_member = ? and access_date = ?", sender, receiver, helpers.GetCurrentTime("Y-m-d")).
		Find(&fc)
	if fc.ID == "" {
		fc.ID = helpers.NewUUID()
		fc.IdUser = sender
		fc.IdMember = receiver
		fc.AccessDate = helpers.GetCurrentTime("Y-m-d")
		db.Table("freechats").Create(&fc)
	}
}

func CanSendMessage(ctx helpers.Context, user UserDetail, receiver User) (bool, bool, int64, int64) {
	var can bool = false
	hasFreeChat, availableFreeChat := HasFreeChat(ctx, user, receiver)
	var availableChat int64 = int64(math.Floor(float64(user.Diamond) / 15))
	if hasFreeChat {
		can = true
	} else if user.Diamond >= 15 {
		can = true
	}
	return can, hasFreeChat, availableChat, availableFreeChat
}

func GetOnlineSidebar(ctx helpers.Context, user User) map[string]interface{} {
	res := map[string]interface{}{}
	var userDetail UserDetail = GetUserDetail(ctx, user.ID)
	res["total_member"] = GetTotalOnlineMembers(ctx, userDetail)
	res["total_friend"] = GetTotalOnlineFriends(ctx, userDetail)
	showFriend := helpers.GetMinimum(res["total_friend"].(int64), 10)
	showMember := helpers.GetMinimum(res["total_member"].(int64), 25-showFriend)
	res["friends"] = GetOnlineFriends(ctx, userDetail, showFriend)
	res["members"] = GetOnlineMembers(ctx, userDetail, showMember)
	result := map[string]interface{}{
		"action": "onlinemembers",
		"data":   res,
	}
	return result
}
