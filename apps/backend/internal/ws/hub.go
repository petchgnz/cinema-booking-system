package ws
import (
	"encoding/json"
	"log"
)

type Hub struct {
	rooms map[string]map[*Client]bool

	broadcast  chan *RoomMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *RoomMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.rooms[client.roomID]; !ok {
				h.rooms[client.roomID] = make(map[*Client]bool)
			}
			h.rooms[client.roomID][client] = true
			log.Printf("[Hub] Client joined room: %s (total: %d)", client.roomID, len(h.rooms[client.roomID]))

		case client := <-h.unregister:
			if room, ok := h.rooms[client.roomID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.send)

					if len(room) == 0 {
						delete(h.rooms, client.roomID)
					}
				}
			}
			log.Printf("[Hub] Client left room: %s", client.roomID)

		case msg := <-h.broadcast:
			log.Printf("[Hub] Received broadcast for room: %s", msg.RoomID)
			if room, ok := h.rooms[msg.RoomID]; ok {
				for client := range room {
					select {
					case client.send <- msg.Payload:
					default:
						close(client.send)
						delete(room, client)
					}
				}
			} else {
				log.Printf("[Hub] No clients in room: %s", msg.RoomID)
			}
		}
	}
}

// send Seat event to every client in room
func (h *Hub) BroadcastSeatUpdate(showtimeID, eventType, seatNumber, status string) {
	event := SeatEvent{
		Type:       eventType,
		ShowtimeID: showtimeID,
		SeatNumber: seatNumber,
		Status:     status,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[Hub] Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- &RoomMessage{
		RoomID:  showtimeID,
		Payload: payload,
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}
