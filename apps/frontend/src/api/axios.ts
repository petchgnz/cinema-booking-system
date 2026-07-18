import axios, { type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '../stores/auth'
import { auth } from '../firebase';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL + '/api/v1'
})

// intercepter
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const authStore = useAuthStore();
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

api.interceptors.response.use((response) => response, async (error) => {
  const originalRequest = error.config

  if (error.response?.status === 401 && !originalRequest._retry) {
    const newToken = await auth.currentUser?.getIdToken(true)
    if (newToken) {
      const authStore = useAuthStore();
      authStore.token = newToken

      error.config.headers.Authorization = `Bearer ${newToken}`
      return api.request(error.config)
    }
  }
  return Promise.reject(error)
})

export default api