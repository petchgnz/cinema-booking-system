package handler

import (
	"log"
	"net/http"

	"cinema-booking/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// allow every origins when development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsHandler struct {
	hub *ws.Hub
}

func NewWsHandler(hub *ws.Hub) *WsHandler {
	return &WsHandler{
		hub: hub,
	}
}

// GET /ws/showtimes/:id
func (h *WsHandler) ServeWs(c *gin.Context) {
	showtimeID := c.Param("id")
	if showtimeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "showtime_id is required"})
		return
	}

	// upgrade http --> websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Failed to upgrade connection: %v", err)
		return
	}

	client := ws.NewClient(h.hub, conn, showtimeID)
	h.hub.Register(client)

	log.Printf("[WS] New client connected to showtime: %s", showtimeID)

	go client.WritePump()
	go client.ReadPump()
}
