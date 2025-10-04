package realtime

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var h = &Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte, 256),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func RunHub() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			var q Quote
			if err := json.Unmarshal(msg, &q); err != nil {
				// broadcast raw if cannot parse
				h.mu.RLock()
				for c := range h.clients {
					select {
					case c.send <- msg:
					default:
						close(c.send)
						delete(h.clients, c)
					}
				}
				h.mu.RUnlock()
				continue
			}
			h.mu.RLock()
			for c := range h.clients {
				if c.subs[q.Code] {
					select {
					case c.send <- msg:
					default:
						close(c.send)
						delete(h.clients, c)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func Broadcast(msg []byte) {
	h.broadcast <- msg
}
