import type { Showtime } from '../types'
import api from './axios'

export const getShowtimes = () => api.get<Showtime[]>('/showtimes')

export const getShowtimeById = (id: string) => api.get<Showtime>(`/showtimes/${id}`)