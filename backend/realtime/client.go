package realtime

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
	subs map[string]bool
}

// 处理客户端消息
func (c *Client) readPump() {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		var req struct {
			Action  string   `json:"action"`
			Symbols []string `json:"symbols"`
		}
		if err := json.Unmarshal(msg, &req); err != nil {
			continue
		}
		if req.Action == "subscribe" {
			for _, s := range req.Symbols {
				c.subs[s] = true
			}
		} else if req.Action == "unsubscribe" {
			for _, s := range req.Symbols {
				delete(c.subs, s)
			}
		}
	}
}

// 向客户端发送消息
func (c *Client) writePump() {
	for {
		msg, ok := <-c.send
		if !ok {
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
