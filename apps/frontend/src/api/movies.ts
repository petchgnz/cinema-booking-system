import type { Movie } from '../types'
import api from './axios'

export const getMovies = () => api.get<Movie[]>('/movies')

export const getMovieById = (id: string) => api.get<Movie>(`/movies/${id}`)
