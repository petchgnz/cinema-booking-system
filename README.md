# Cinema Ticket Booking System

A full-stack cinema ticket booking system built with Go (Gin), Vue 3, MongoDB, Redis, RabbitMQ, and WebSocket — demonstrating distributed locking, real-time seat updates, and async event processing.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.24 + Gin |
| Frontend | Vue 3 + TypeScript + Tailwind CSS v4 |
| Database | MongoDB 7 |
| Distributed Lock | Redis 7 (SetNX) |
| Message Queue | RabbitMQ 3.13 |
| Real-time | WebSocket (gorilla/websocket) |
| Authentication | Firebase Google OAuth |
| Deployment | Docker Compose |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                    Vue 3 Frontend                   │
│         (Pinia + Vue Router + Axios + WS)           │
└────────────────────┬────────────────────────────────┘
                     │ HTTP / WebSocket
┌────────────────────▼────────────────────────────────┐
│                  Go (Gin) Backend                   │
│  ┌──────────┐  ┌──────────┐  ┌────────────────┐    │
│  │ Handler  │→ │ Service  │→ │  Repository    │    │
│  └──────────┘  └────┬─────┘  └───────┬────────┘    │
│                     │                │              │
│            ┌────────┴────────┐       │              │
│            │                 │       ▼              │
│       Redis Lock        RabbitMQ   MongoDB          │
│       (SetNX TTL)       Publisher                   │
│                              │                      │
│                    ┌─────────▼──────────┐           │
│                    │  RabbitMQ Consumer │           │
│                    │  (goroutine)       │           │
│                    └────────────────────┘           │
│                                                     │
│  ┌─────────────────────────────────────────────┐    │
│  │          WebSocket Hub (goroutine)          │    │
│  │     rooms: map[showtimeID][]*Client         │    │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

### Clean Architecture Layers

```
cmd/
  main.go              ← Wires all dependencies (composition root)
  seed/main.go         ← Database seeder

internal/
  config/              ← Env config + Firebase init
  model/               ← Domain structs (Movie, Showtime, Booking, Seat)
  dto/                 ← Request/Response shapes
  repository/          ← MongoDB data access layer
  service/             ← Business logic (booking, locking)
  handler/             ← HTTP + WebSocket handlers (Gin)
  middleware/          ← Firebase JWT auth
  messaging/           ← RabbitMQ publisher + consumer
  ws/                  ← WebSocket hub + client
```

---

## Key Design Decisions

### 1. Distributed Lock with Redis (SetNX)

**Problem:** Two users can select the same seat simultaneously — without a lock, both bookings would succeed.

**Solution:** When a user selects a seat, the backend calls `Redis.SetNX(key, userID, 5min)`. SetNX is atomic — only one caller wins.

```
lock:seat:{showtimeID}:{seatNumber} = {userID}  TTL: 5 minutes
```

- If SetNX returns `true` → lock acquired → seat is yours for 5 minutes
- If SetNX returns `false` → another user holds the lock → return 409 Conflict
- On confirmation → seat is marked `booked` in MongoDB → lock released
- On TTL expiry → lock auto-released → seat becomes available again

### 2. Message Queue with RabbitMQ

**Problem:** After confirming a booking, the system needs to update booking status to `confirmed`. Doing this synchronously in the HTTP handler would make the API slow and tightly coupled.

**Solution:** After creating a booking, the service publishes a `booking.created` event to RabbitMQ. A background consumer goroutine picks it up and updates the booking status to `confirmed`.

```
Publisher (HTTP request) ──→ booking.exchange ──→ booking.confirmation queue
                                                         │
                                                Consumer goroutine (async)
                                                         │
                                                 Update status = confirmed
```

- Exchange: `booking.exchange` (direct)
- Routing key: `booking.created`
- Ack strategy: manual — Ack on success, Nack+requeue on DB errors, Nack+discard on parse errors

### 3. Real-time Seat Updates with WebSocket

**Problem:** When User A locks a seat, User B (viewing the same showtime page) should see the seat turn yellow immediately — without polling.

**Solution:** A WebSocket Hub manages rooms keyed by `showtimeID`. When any seat changes state, `BroadcastSeatUpdate()` sends the event to all clients in that room.

```
User locks seat → BookingService → Hub.BroadcastSeatUpdate()
                                       │
                               broadcast to all clients
                               in room[showtimeID]
                                       │
                              Vue frontend updates
                              seat color in real-time
```

**Thread safety:** The Hub uses a single goroutine + channels instead of a mutex. All map operations (`register`, `unregister`, `broadcast`) are funneled through `hub.Run()` — no concurrent map writes possible.

### 4. SeatBroadcaster Interface (Circular Import Prevention)

`booking_service.go` needs to call the WebSocket hub, but importing the `ws` package from the `service` package would create a circular dependency (`service → ws → service`).

**Solution:** Define a `SeatBroadcaster` interface inside the `service` package. The `ws.Hub` implements it without knowing about the service layer.

```go
// service/booking_service.go
type SeatBroadcaster interface {
    BroadcastSeatUpdate(showtimeID, eventType, seatNumber, status string)
}
```

### 5. Firebase Credentials in Docker

Instead of copying the credentials file into the container (a security risk), the backend reads `FIREBASE_CREDENTIALS_JSON` from environment variables. Local development falls back to a file.

---

## Running with Docker

### Prerequisites

- Docker Desktop
- A Firebase project with Google Sign-In enabled
- A Firebase service account credentials JSON

### 1. Clone the repository

```bash
git clone <repo-url>
cd cinema-booking
```

### 2. Set up environment variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

Edit `.env`:

```env
# MongoDB
MONGO_USER=admin
MONGO_PASSWORD=your_mongo_password

# Redis
REDIS_PASSWORD=your_redis_password

# RabbitMQ
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=your_rabbitmq_password

# Firebase (paste your service account JSON as a single line)
FIREBASE_CREDENTIALS_JSON={"type":"service_account","project_id":"..."}

# Firebase (Frontend)
VITE_FIREBASE_API_KEY=your_api_key
VITE_FIREBASE_AUTH_DOMAIN=your_project.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=your_project_id
```

> **Getting `FIREBASE_CREDENTIALS_JSON`:** Go to Firebase Console → Project Settings → Service Accounts → Generate new private key. Open the downloaded JSON, copy the entire content, and paste it as a single line (remove all newlines).

### 3. Start all services

```bash
docker compose up --build
```

This starts: MongoDB, Redis, RabbitMQ, Backend (Go), Frontend (Vue/nginx)

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| RabbitMQ Management | http://localhost:15672 |

### 4. Seed demo data (first time only)

```bash
docker compose run --rm backend ./seed
```

This creates 2 movies (The Avengers, Inception) with 2 showtimes each and 40 seats per showtime. Running the seed again is safe — it checks for existing data and skips if already seeded.

### 5. Start without rebuilding (subsequent runs)

```bash
docker compose up -d
```

---

## API Endpoints

Base URL: `http://localhost:8080`

### Health Check

```http
GET /health
```

### Movies

```http
GET /api/v1/movies
```

```http
GET /api/v1/movies/:id
```

### Showtimes

```http
GET /api/v1/showtimes
GET /api/v1/showtimes/:id
```

### Bookings (requires Firebase JWT in `Authorization: Bearer <token>`)

#### Lock seats (temporary hold — 5 minutes)

```http
POST /api/v1/bookings/lock
Authorization: Bearer <firebase_id_token>
Content-Type: application/json

{
  "showtime_id": "685a1f2e3c4b5d6e7f8a9b0c",
  "seat_numbers": ["A1", "A2"]
}
```

Response:
```json
{
  "message": "seats locked successfully",
  "expires_in": "5 minutes",
  "seat_numbers": ["A1", "A2"]
}
```

#### Confirm booking

```http
POST /api/v1/bookings
Authorization: Bearer <firebase_id_token>
Content-Type: application/json

{
  "showtime_id": "685a1f2e3c4b5d6e7f8a9b0c",
  "seat_numbers": ["A1", "A2"]
}
```

Response:
```json
{
  "_id": "685b2c3d4e5f6a7b8c9d0e1f",
  "user_id": "firebase_uid_here",
  "showtime_id": "685a1f2e3c4b5d6e7f8a9b0c",
  "seat_numbers": ["A1", "A2"],
  "status": "pending",
  "created_at": "2025-07-19T10:30:00Z"
}
```

> Note: `status` starts as `pending`. The RabbitMQ consumer updates it to `confirmed` asynchronously within seconds.

### WebSocket

```
ws://localhost:8080/ws/showtimes/:showtimeId
```

**Incoming events (server → client):**

```json
{ "type": "seat_locked", "showtime_id": "...", "seat_number": "A1", "status": "locked" }
{ "type": "seat_booked", "showtime_id": "...", "seat_number": "A1", "status": "booked" }
```

---

## Booking Flow

```
User selects seats
      │
      ▼
POST /bookings/lock
  → Redis SetNX per seat (atomic, 5min TTL)
  → WebSocket broadcast: seat_locked → all users see seat turn yellow
      │
      ▼
User clicks "Confirm"
      │
      ▼
POST /bookings
  → Verify Redis lock still held by this user
  → Create Booking in MongoDB (status: pending)
  → Update seat status to "booked" in Showtime document
  → Release Redis lock
  → WebSocket broadcast: seat_booked → all users see seat turn red
  → Publish booking.created event to RabbitMQ
      │
      ▼ (async)
RabbitMQ Consumer
  → Update Booking status: pending → confirmed
```

---

## Project Structure

```
cinema-booking/
├── docker-compose.yml
├── .env.example
├── apps/
│   ├── backend/
│   │   ├── Dockerfile
│   │   ├── cmd/
│   │   │   ├── main.go          ← App entrypoint + dependency wiring
│   │   │   └── seed/main.go     ← Database seeder
│   │   └── internal/
│   │       ├── config/          ← Env + Firebase config
│   │       ├── model/           ← Domain models
│   │       ├── dto/             ← Request/Response DTOs
│   │       ├── repository/      ← MongoDB repositories
│   │       ├── service/         ← Business logic
│   │       ├── handler/         ← Gin HTTP + WS handlers
│   │       ├── middleware/       ← Firebase auth middleware
│   │       ├── messaging/       ← RabbitMQ publisher + consumer
│   │       └── ws/              ← WebSocket hub + client
│   └── frontend/
│       ├── Dockerfile
│       ├── nginx.conf
│       └── src/
│           ├── api/             ← Axios API calls
│           ├── composables/     ← useWebSocket, useAuth
│           ├── stores/          ← Pinia auth store
│           ├── router/          ← Vue Router + auth guards
│           ├── views/           ← Page components
│           └── types/           ← TypeScript types
```

---

## Concurrency Handling

| Concern | Approach |
|---|---|
| Double-booking same seat | Redis SetNX (atomic) — only one user acquires the lock |
| WebSocket map race condition | Single goroutine Hub pattern — no mutex needed |
| RabbitMQ publish thread safety | New AMQP channel per publish call |
| Lock expiry | 5-minute TTL auto-releases locks if user abandons checkout |
| RabbitMQ startup race | Docker healthcheck (`check_port_connectivity`) + `condition: service_healthy` |
