<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getShowtimeById  } from '../api/showtimes'
import { lockSeats, createBooking } from '../api/bookings'
import { useWebSocket } from '../composables/useWebSocket'
import type { Showtime, Seat, SeatEvent } from '../types'

const route = useRoute()
const showtimeId = route.params.id as string

const showtime = ref<Showtime | null>(null)
const loading = ref<boolean>(true)
const selectedSeats = ref<string[]>([])
const isLocked = ref<boolean>(false)
const message = ref<string>('')

// WebSocket composable
const { isConnected, connect } = useWebSocket(showtimeId)

onMounted(async () => {
  try {
    const res = await getShowtimeById(showtimeId)
    showtime.value = res.data
  } finally {
    loading.value = false
  }

  connect((event: SeatEvent) => {
    if (!showtime.value) return

    const seat = showtime.value.seats.find(s => s.seat_number === event.seat_number)
    if (seat) {
      seat.status = event.status
    }
  })
})

// จัดกลุ่ม seats ตาม row (A, B, C, ...)
const seatsByRow = computed(() => {
  if (!showtime.value) return {}

  const rows: Record<string, Seat[]> = {}
  for (const seat of showtime.value.seats) {
    const row = seat.seat_number[0]
    if (!rows[row]) rows[row] = []
    rows[row].push(seat)
  }
  return rows
})

function getSeatClass(seat: Seat): string {
  if (selectedSeats.value.includes(seat.seat_number)) {
    return 'bg-indigo-500 cursor-pointer'
  }
  switch (seat.status) {
    case 'available': return 'bg-gray-600 hover:bg-gray-500 cursor-pointer'
    case 'locked':    return 'bg-yellow-500 cursor-not-allowed opacity-70'
    case 'booked':    return 'bg-red-600 cursor-not-allowed opacity-70'
    default:          return 'bg-gray-600'
  }
}

function toggleSeat(seat: Seat): void {
  if (seat.status !== 'available') return
  if (isLocked.value) return

  const idx = selectedSeats.value.indexOf(seat.seat_number)
  if (idx === -1) {
    selectedSeats.value.push(seat.seat_number)
  } else {
    selectedSeats.value.splice(idx, 1)
  }
}

async function handleLock(): Promise<void> {
  if (selectedSeats.value.length === 0) return
  try {
    await lockSeats(showtimeId, selectedSeats.value)
    isLocked.value = true
    message.value = `Seats locked! Confirm within 5 minutes.`
  } catch (err: any) {
    message.value = err.response?.data?.error ?? 'Failed to lock seats'
  }
}

async function handleConfirm(): Promise<void> {
  try {
    await createBooking(showtimeId, selectedSeats.value)
    message.value = 'Booking confirmed!'
    isLocked.value = false
    selectedSeats.value = []
  } catch (err: any) {
    message.value = err.response?.data?.error ?? 'Failed to confirm booking'
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-900 text-white">
    <nav class="bg-gray-800 px-6 py-4 flex items-center gap-4">
      <RouterLink to="/" class="text-gray-400 hover:text-white text-sm transition">← Back</RouterLink>
      <h1 class="text-xl font-bold">Select Seats</h1>
      <span class="ml-auto text-sm" :class="isConnected ? 'text-green-400' : 'text-red-400'">
        {{ isConnected ? '● Live' : '○ Connecting...' }}
      </span>
    </nav>

    <main class="max-w-3xl mx-auto px-6 py-10">
      <div v-if="loading" class="text-gray-400">Loading...</div>

      <div v-else-if="showtime">
        <!-- Screen indicator -->
        <div class="bg-gray-700 text-center text-gray-400 text-sm py-2 rounded-lg mb-8">
          SCREEN
        </div>

        <!-- Seat Map -->
        <div class="space-y-3 mb-10">
          <div
            v-for="(seats, row) in seatsByRow"
            :key="row"
            class="flex items-center gap-2"
          >
            <span class="text-gray-500 text-sm w-4">{{ row }}</span>
            <div class="flex gap-2 flex-wrap">
              <button
                v-for="seat in seats"
                :key="seat.seat_number"
                class="w-9 h-9 rounded text-xs font-semibold transition"
                :class="getSeatClass(seat)"
                @click="toggleSeat(seat)"
              >
                {{ seat.seat_number.slice(1) }}
              </button>
            </div>
          </div>
        </div>

        <!-- Legend -->
        <div class="flex gap-6 text-sm text-gray-400 mb-8">
          <span class="flex items-center gap-2"><span class="w-4 h-4 rounded bg-gray-600 inline-block"></span> Available</span>
          <span class="flex items-center gap-2"><span class="w-4 h-4 rounded bg-indigo-500 inline-block"></span> Selected</span>
          <span class="flex items-center gap-2"><span class="w-4 h-4 rounded bg-yellow-500 inline-block"></span> Locked</span>
          <span class="flex items-center gap-2"><span class="w-4 h-4 rounded bg-red-600 inline-block"></span> Booked</span>
        </div>

        <!-- Actions -->
        <div class="bg-gray-800 rounded-xl p-6 space-y-4">
          <p class="text-gray-300 text-sm">
            Selected: <span class="font-semibold text-white">{{ selectedSeats.join(', ') || '-' }}</span>
          </p>

          <p v-if="message" class="text-yellow-400 text-sm">{{ message }}</p>

          <div class="flex gap-3">
            <button
              v-if="!isLocked"
              @click="handleLock"
              :disabled="selectedSeats.length === 0"
              class="bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed px-6 py-2 rounded-lg font-semibold text-sm transition"
            >
              Lock Seats
            </button>

            <button
              v-if="isLocked"
              @click="handleConfirm"
              class="bg-green-600 hover:bg-green-500 px-6 py-2 rounded-lg font-semibold text-sm transition"
            >
              Confirm Booking
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>