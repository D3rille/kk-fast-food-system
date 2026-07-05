package ws

import (
	"encoding/json"
	"log/slog"

	"github.com/D3rille/kk-fast-food-system/internal/service"
)

// Hub maintains the set of active WebSocket clients and fans out order events.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan service.OrderEvent
	register   chan *Client
	unregister chan *Client
	log        *slog.Logger
}

// NewHub creates an idle Hub that must be started with Run.
func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan service.OrderEvent, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		log:        log,
	}
}

// Run processes register/unregister/broadcast events. Must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.log.Info("ws: client connected", "total_clients", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.log.Info("ws: client disconnected", "total_clients", len(h.clients))
			}

		case event := <-h.broadcast:
			msg, err := json.Marshal(event)
			if err != nil {
				h.log.Error("ws: failed to marshal event", "error", err)
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Slow client — drop and disconnect.
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Broadcast implements service.EventBroadcaster, sending the event to all connected kitchen clients.
func (h *Hub) Broadcast(event service.OrderEvent) {
	select {
	case h.broadcast <- event:
	default:
		h.log.Warn("ws: broadcast channel full, dropping event", "order_id", event.OrderID, "type", event.Type)
	}
}
