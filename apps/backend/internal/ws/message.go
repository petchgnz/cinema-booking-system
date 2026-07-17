package ws

type SeatEvent struct {
	Type        string `json:"type"`
	ShowtimeID  string `json:"showtime_id"`
	SeatNumber  string `json:"seat_number"`
	Status      string `json:"status"`
}

// RoomMessage คือ internal message ที่ Hub ใช้ส่งไปหา clients ใน room
type RoomMessage struct {
	RoomID  string
	Payload []byte
}