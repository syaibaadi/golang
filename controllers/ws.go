package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/pangestu18/janji-online/chat/helpers"
	"gitlab.com/pangestu18/janji-online/chat/models"

	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	conn helpers.Connection
}

type subscription struct {
	target helpers.Subscription
}

type Message struct {
	Action      string
	TextMessage string `json:"msg"`
	Token       string
	MessageTo   string `json:"to"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (s subscription) readPump(ctx echo.Context) {
	c := s.target.Conn
	defer func() {
		helpers.ActiveHub.Unregister <- s.target
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(helpers.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(helpers.PongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(helpers.PongWait)); return nil })
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					fmt.Printf("error: %v", err)
				}
			}
			break
		}
		m := Message{}
		json.Unmarshal(msg, &m)
		user, valid := models.ValidateUserByToken(ctx, m.Token, s.target.Room)
		if !valid {
			msg = []byte(`{"error":"Invalid authentication token"}`)
			b := helpers.Message{msg, s.target.Room}
			helpers.ActiveHub.Broadcast <- b
		} else {
			if m.Action != "statuschat" {
				userDetail := models.GetUserDetail(ctx, user.ID)
				toUser := models.GetUserBySignature(ctx, m.MessageTo)
				fmt.Println("Chat From ", user, " To ", toUser)
				hasDiamond, hasFreeChat, availableChat, availableFreeChat := models.CanSendMessage(ctx, userDetail, toUser)
				fmt.Println("AvailableChat ", hasDiamond, hasFreeChat, availableChat)
				if hasDiamond || hasFreeChat {
					if toUser.ID != 0 {
						m.TextMessage, err = models.CreateMessage(ctx, user.ID, toUser.ID, m.TextMessage)
						if err == nil {
							sentmsg := []byte(`{"msg":"` + m.TextMessage + `","from":"` + s.target.Room + `","action":"receivechat"}`)
							b1 := helpers.Message{sentmsg, m.MessageTo}

							if hasFreeChat {
								models.InsertFreeChatHistoryIfNotExist(ctx, user.ID, toUser.ID)
								availableFreeChat--
							} else if hasDiamond {
								models.UpdateDiamond(ctx, userDetail)
								availableChat--
							}
							statusmsg := []byte(`{"msg":"` + m.TextMessage + `","to":"` + m.MessageTo + `","action":"statuschat","available_chat":"` + helpers.Convert(availableChat).String() + `","free_chat":"` + helpers.Convert(availableFreeChat).String() + `"}`)
							b2 := helpers.Message{statusmsg, s.target.Room}

							helpers.ActiveHub.Broadcast <- b1
							helpers.ActiveHub.Broadcast <- b2
						}
					}
				} else {
					msg = []byte(`{"error":"Please buy heart to send chat!"}`)
					b := helpers.Message{msg, s.target.Room}
					helpers.ActiveHub.Broadcast <- b
				}
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (s *subscription) writePump(ctx echo.Context) {
	c := s.target.Conn
	ticker := time.NewTicker(helpers.PingPeriod)
	memberTkr := time.NewTicker(helpers.UpdateOnlineMemberPeriod())
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				s.target.Conn.Write(websocket.CloseMessage, []byte{})
				return
			}
			if err := s.target.Conn.Write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := s.target.Conn.Write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-memberTkr.C:
			models.DeleteOldMessages(ctx)
			user := models.GetUserBySignature(ctx, s.target.Room)
			online := models.GetOnlineSidebar(ctx, user)
			fmt.Println(online)
			b, err := json.Marshal(online)
			if err == nil {
				if err = s.target.Conn.Write(websocket.TextMessage, b); err != nil {
					return
				}
			} else {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(c echo.Context) error {
	upgrader.CheckOrigin = helpers.CheckOrigin
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("Error", err)
		log.Println(err)
		return err
	} else {
		fmt.Println("Will connect to ", c.Param("room"))
	}

	ws := &connection{conn: helpers.Connection{Send: make(chan []byte, 256), Conn: conn}}
	client := subscription{target: helpers.Subscription{&ws.conn, c.Param("room")}}

	helpers.ActiveHub.Register <- client.target

	// Allow collecti,,,,,f memory referenced by the caller by doing all work in
	// new goroutines.
	go client.readPump(c)
	go client.writePump(c)

	return nil
}
