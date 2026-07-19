# Cinema Ticket Booking System

A full-stack cinema ticket booking system built with Go (Gin), Vue 3, MongoDB, Redis, RabbitMQ, and WebSocket — demonstrating distributed locking, real-time seat updates, and async event processing.

---

## 1. System Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│                    Vue 3 Frontend                   │
│         (Pinia + Vue Router + Axios + WS)           │
└────────────────────┬────────────────────────────────┘
                     │ HTTP / WebSocket
┌────────────────────▼────────────────────────────────┐
│                  Go (Gin) Backend                   │
│  ┌──────────┐  ┌──────────┐  ┌────────────────┐     │
│  │ Handler  │→ │ Service  │→ │  Repository    │     │
│  └──────────┘  └────┬─────┘  └───────┬────────┘     │
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
  model/               ← Domain structs (Movie, Showtime, Booking, Seat, AuditLog)
  dto/                 ← Request/Response shapes
  repository/          ← MongoDB data access layer
  service/             ← Business logic (booking, locking, audit)
  handler/             ← HTTP + WebSocket handlers (Gin)
  middleware/          ← Firebase JWT auth + Admin role check
  messaging/           ← RabbitMQ publisher + consumer
  notification/        ← Notifier interface + MockNotifier
  ws/                  ← WebSocket hub + client
```

---

## 2. Tech Stack Overview

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

## 3. Booking Flow

```
1. ผู้ใช้เลือกที่นั่ง (frontend)
         │
         ▼
2. POST /api/v1/bookings/lock
   → Redis SetNX per seat (atomic, 5 min TTL)
   → ถ้า SetNX = false → seat ถูก lock โดยคนอื่น → 409 Conflict
   → ถ้า SetNX = true  → WebSocket broadcast: seat_locked
                        → ทุก user เห็นที่นั่งเปลี่ยนเป็นสีเหลืองทันที
         │
         ▼
3. ผู้ใช้กด Confirm ภายใน 5 นาที
         │
         ▼
4. POST /api/v1/bookings
   → ตรวจสอบว่า Redis lock ยังเป็นของ user นี้อยู่
   → สร้าง Booking ใน MongoDB (status: pending)
   → อัปเดต seat status → booked
   → Release Redis lock
   → WebSocket broadcast: seat_booked → ที่นั่งเปลี่ยนเป็นสีแดง
   → Publish booking.created event → RabbitMQ
         │
         ▼ (async)
5. RabbitMQ Consumer
   → Update Booking status: pending → confirmed
   → Trigger MockNotifier (log การแจ้งเตือน)

กรณีไม่ชำระภายใน 5 นาที:
   → Redis TTL หมด → lock ถูกปล่อยอัตโนมัติ
   → seat กลับเป็น available
```

---

## 4. Redis Lock Strategy

**ปัญหา:** ผู้ใช้ 2 คนสามารถเลือกที่นั่งเดียวกันพร้อมกันได้ ถ้าไม่มี lock ทั้งคู่จะ booking สำเร็จทั้งคู่

**วิธีแก้:** ใช้ `Redis.SetNX` ซึ่งเป็น atomic operation — มีแค่ผู้ชนะ 1 คนเท่านั้น

```
lock:seat:{showtimeID}:{seatNumber} = {userID}   TTL: 5 minutes
```

**ทำไมถึงเลือก SetNX:**
- **Atomic** — ไม่มี race condition ระหว่าง check และ set
- **TTL (Time To Live) อัตโนมัติ** — ถ้าผู้ใช้ทิ้ง tab หรือ timeout ที่นั่งจะถูกปล่อยโดยอัตโนมัติโดยไม่ต้องมี cleanup job
- **userID เป็น value** — ป้องกัน user อื่น release lock ของคนอื่น (check value ก่อน delete เสมอ)

**Trade-offs:**
- Lock อยู่แค่ใน Redis ไม่ได้ sync กับ MongoDB seat status → ต้อง verify lock ก่อน confirm เสมอ
- ถ้า Redis ล่ม lock หายทั้งหมด → seat อาจถูก double-book ได้ชั่วคราว

---

## 5. Message Queue (RabbitMQ)

**ปัญหา:** การอัปเดต booking status เป็น `confirmed` ไม่ควรทำใน HTTP request โดยตรง เพราะทำให้ API ช้า

**วิธีแก้:** หลัง booking สร้างสำเร็จ service จะ publish event ไป RabbitMQ แล้ว background consumer จัดการต่อ

```
Publisher (HTTP request)
    │
    └──→ booking.exchange (direct exchange)
              │
              └──→ booking.confirmation queue
                          │
                   Consumer goroutine (async)
                          │
                   Update status = confirmed
                          │
                   Trigger MockNotifier
```

**Ack Strategy:**
- **Ack** — อัปเดต DB สำเร็จ
- **Nack + requeue=true** — DB error → retry ได้
- **Nack + requeue=false** — parse error → message เสีย ไม่ควร retry

**ทำไมไม่ทำใน HTTP handler:**
- HTTP handler ควรตอบกลับเร็ว → ผู้ใช้ไม่รอ (respones ส่วนสำคัญไปก่อน แล้ว tasks ที่เหลือไปทำต่อใน consumer)
- ถ้า DB ช้าหรือล่ม RabbitMQ จะ retry ให้อัตโนมัติ
- Decoupling — consumer สามารถ scale แยกจาก API ได้

---

## 6. วิธีรันระบบ

### Prerequisites

- Docker Desktop
- Firebase project ที่เปิด Google Sign-In
- Firebase service account credentials JSON

### ขั้นตอน

```bash
# 1. Clone
git clone <repo-url>
cd cinema-booking

# 2. ตั้งค่า environment
cp .env.example .env
# แก้ไข .env ให้ครบ (ดู .env.example สำหรับรายละเอียด)

# 3. Build และรัน
docker compose up --build -d

# 4. Seed ข้อมูลตัวอย่าง (ครั้งแรกเท่านั้น)
docker compose run --rm backend ./seed

# 5. รันครั้งต่อไป (ไม่ต้อง build ใหม่)
docker compose up -d
```

### Services

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| RabbitMQ Management | http://localhost:15672 |

### Environment Variables (.env)

```env
MONGO_USER=admin
MONGO_PASSWORD=your_password
REDIS_PASSWORD=your_password
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=your_password
FIREBASE_CREDENTIALS_JSON={"type":"service_account",...}  # single line
VITE_FIREBASE_API_KEY=...
VITE_FIREBASE_AUTH_DOMAIN=...
VITE_FIREBASE_PROJECT_ID=...
ADMIN_EMAIL=your-email@gmail.com  # user นี้จะได้ role=admin
```

> **FIREBASE_CREDENTIALS_JSON:** Firebase Console → Project Settings → Service Accounts → Generate new private key → copy ทั้งหมดเป็น single line

### Admin Role

user ที่ email ตรงกับ `ADMIN_EMAIL` จะได้ role `admin` โดยอัตโนมัติตอน login ครั้งแรก

Admin เท่านั้นที่สร้าง movie และ showtime ได้ (`POST /api/v1/movies`, `POST /api/v1/showtimes`)

---

## 7. API Endpoints

Base URL: `http://localhost:8080`

### Public

```http
GET /health
GET /api/v1/movies
GET /api/v1/movies/:id
GET /api/v1/showtimes
GET /api/v1/showtimes/:id
```

### User (requires `Authorization: Bearer <firebase_id_token>`)

#### Lock seats

```http
POST /api/v1/bookings/lock

{
  "showtime_id": "685a1f2e3c4b5d6e7f8a9b0c",
  "seat_numbers": ["A1", "A2"]
}
```

#### Confirm booking

```http
POST /api/v1/bookings

{
  "showtime_id": "685a1f2e3c4b5d6e7f8a9b0c",
  "seat_numbers": ["A1", "A2"]
}
```

### Admin only (requires admin role)

```http
POST /api/v1/movies
POST /api/v1/showtimes
```

### WebSocket

```
ws://localhost:8080/ws/showtimes/:showtimeId
```

Events (server → client):

```json
{ "type": "seat_locked", "showtime_id": "...", "seat_number": "A1", "status": "locked" }
{ "type": "seat_booked", "showtime_id": "...", "seat_number": "A1", "status": "booked" }
```

---

## 8. Assumptions & Trade-offs

**Assumptions:**
- ผู้ใช้ต้อง login ด้วย Google ก่อน lock หรือ book ที่นั่งได้
- 1 lock request = lock ทุกที่นั่งที่เลือกพร้อมกัน ถ้า seat ใด seat หนึ่งล้มเหลวจะ rollback ทั้งหมด

**Trade-offs:**

| Decision | Why | Alternative |
|---|---|---|
| Redis SetNX TTL 5 min | พอดีกับ checkout flow ไม่นานเกินไป | Shorter TTL = UX แย่ลง, Longer = seat ถูก hold นานเกิน |
| RabbitMQ direct exchange | Simple, predictable routing | Topic exchange ถ้าต้องการ event types หลายแบบ |
| WebSocket Hub single goroutine | Thread safe โดยไม่ต้องใช้ mutex | Mutex ง่ายกว่าแต่เสี่ยง deadlock มากกว่า |
| MockNotifier | ลด dependency ภายนอก ใช้ interface ขยายได้ | Real email/Line พร้อม production |
| Firebase Auth | ไม่ต้องสร้าง auth ระบบเอง | JWT custom อ่อนแอกว่าถ้า implement ไม่ดี |
