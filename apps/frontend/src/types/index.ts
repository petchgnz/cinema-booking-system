export interface Movie {
  id: string
  title: string
  description: string
  duration: number
  poster_url: string
}

export interface Seat {
  seat_number: string
  status: 'available' | 'locked' | 'booked'
}

export interface Showtime {
  id: string
  movie_id: string
  start_time: string
  end_time: string
  hall: string
  seats: Seat[]
}

export interface Booking {
  id: string
  showtime_id: string
  seat_nubmers: string[]
  status: 'pending' | 'confirmed' | 'cancelled'
  created_at: string
}

export interface SeatEvent {
  type: 'seat_locked' | 'seat_booked'
  showtime_id: string
  seat_number: string
  status: 'locked' | 'booked'
}