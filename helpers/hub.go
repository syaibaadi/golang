package helpers

import (
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/pangestu18/janji-online/chat/config"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// Maximum message size allowed from peer.
	MaxMessageSize = 5120
)

func UpdateOnlineMemberPeriod() time.Duration {
	return config.Get("UPDATE_IM_ONLINE").Duration()
}

// connection is an middleman between the websocket connection and the hub.
type Connection struct {
	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte
}

type Message struct {
	Data []byte
	Room string
}

type Subscription struct {
	Conn *Connection
	Room string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	Rooms map[string]map[*Connection]bool

	// Inbound messages from the connections.
	Broadcast chan Message

	// Register requests from the connections.
	Register chan Subscription

	// Unregister requests from connections.
	Unregister chan Subscription
}

var ActiveHub = Hub{
	Broadcast:  make(chan Message),
	Register:   make(chan Subscription),
	Unregister: make(chan Subscription),
	Rooms:      make(map[string]map[*Connection]bool),
}

// write writes a message with the given message type and payload.
func (c *Connection) Write(mt int, payload []byte) error {
	c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
	return c.Conn.WriteMessage(mt, payload)
}

func (h *Hub) Run() {
	for {
		select {
		case s := <-h.Register:
			connections := h.Rooms[s.Room]
			if connections == nil {
				connections = make(map[*Connection]bool)
				h.Rooms[s.Room] = connections
			}
			h.Rooms[s.Room][s.Conn] = true
		case s := <-h.Unregister:
			connections := h.Rooms[s.Room]
			if connections != nil {
				if _, ok := connections[s.Conn]; ok {
					delete(connections, s.Conn)
					close(s.Conn.Send)
					if len(connections) == 0 {
						delete(h.Rooms, s.Room)
					}
				}
			}
		case m := <-h.Broadcast:
			connections := h.Rooms[m.Room]
			for c := range connections {
				select {
				case c.Send <- m.Data:
				default:
					close(c.Send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.Rooms, m.Room)
					}
				}
			}
		}
	}
}
