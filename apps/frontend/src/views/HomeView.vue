<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuth } from '../composables/useAuth'
import { useAuthStore } from '../stores/auth'
import { getMovies } from '../api/movies'
import type { Movie } from '../types'

const { logout } = useAuth()
const authStore = useAuthStore()

const movies = ref<Movie[]>([])
const loading = ref<boolean>(true)

onMounted(async () => {
  try {
    const res = await getMovies()
    movies.value = res.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="min-h-screen bg-gray-900 text-white">
    <nav class="bg-gray-800 px-6 py-4 flex justify-between items-center">
      <h1 class="text-xl font-bold">Cinema Booking</h1>
      <div class="flex items-center gap-4">
        <span class="text-gray-400 text-sm">{{ authStore.user?.displayName }}</span>
        <button @click="logout" class="text-sm bg-gray-700 px-4 py-2 rounded-lg hover:bg-gray-600 transition">
          Logout
        </button>
      </div>
    </nav>

    <main class="max-w-4xl mx-auto px-6 py-10">
      <h2 class="text-2xl font-semibold mb-6">Now Showing</h2>

      <div v-if="loading" class="text-gray-400">Loading...</div>

      <div v-else class="grid grid-cols-1 gap-4">
        <div v-for="movie in movies" :key="movie.id" class="bg-gray-800 rounded-xl p-6 flex justify-between items-center">
          <div>
            <h3 class="text-lg font-semibold">{{ movie.title }}</h3>
            <p class="text-gray-400 text-sm mt-1">{{ movie.duration }} min</p>
            <p class="text-gray-300 text-sm mt-2">{{ movie.description }}</p>
          </div>

          <RouterLink
            :to="`/movies/${movie.id}/showtimes`"
            class="ml-6 shrink-0 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold px-4 py-2 rounded-lg transition"
          >
            Book Now
          </RouterLink>
        </div>
      </div>
    </main>
  </div>
</template>