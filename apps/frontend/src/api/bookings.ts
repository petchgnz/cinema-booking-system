import type { Booking } from '../types';
import api from './axios';

export const lockSeats = (showtimeId: string, seatNumbers: string[]) =>
  api.post('/bookings/lock', {
    showtime_id: showtimeId,
    seat_numbers: seatNumbers,
  });

export const createBooking = (showtimeId: string, seatNumbers: string[]) =>
  api.post<Booking>('/bookings', {
    showtime_id: showtimeId,
    seat_numbers: seatNumbers,
  });
