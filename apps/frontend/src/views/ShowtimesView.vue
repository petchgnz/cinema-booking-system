<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getMovieById } from '../api/movies'
import { getShowtimes } from '../api/showtimes'
import type { Movie, Showtime } from '../types'

const route = useRoute()
const movieId = route.params.movieId as string

const movie = ref<Movie | null>(null)
const allShowtimes = ref<Showtime[]>([])
const loading = ref<boolean>(true)

onMounted(async () => {
  try {
    const [movieRes, showtimesRes] = await Promise.all([
      getMovieById(movieId),
      getShowtimes(),
    ])
    movie.value = movieRes.data
    allShowtimes.value = showtimesRes.data
  } finally {
    loading.value = false
  }
})

const showtimes = computed(() =>
  allShowtimes.value.filter(s => s.movie_id === movieId)
)

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('th-TH', {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}
</script>

<template>
  <div class="min-h-screen bg-gray-900 text-white">
    <nav class="bg-gray-800 px-6 py-4 flex items-center gap-4">
      <RouterLink to="/" class="text-gray-400 hover:text-white text-sm transition">← Back</RouterLink>
      <h1 class="text-xl font-bold">{{ movie?.title ?? 'Showtimes' }}</h1>
    </nav>

    <main class="max-w-4xl mx-auto px-6 py-10">
      <div v-if="loading" class="text-gray-400">Loading...</div>

      <div v-else-if="showtimes.length === 0" class="text-gray-400">
        No showtimes available for this movie.
      </div>

      <div v-else class="grid grid-cols-1 gap-4">
        <div
          v-for="showtime in showtimes"
          :key="showtime.id"
          class="bg-gray-800 rounded-xl p-6 flex justify-between items-center"
        >
          <div>
            <p class="text-gray-400 text-sm">Hall: {{ showtime.hall }}</p>
            <p class="text-white font-semibold mt-1">{{ formatDate(showtime.start_time) }}</p>
            <p class="text-gray-400 text-sm mt-1">
              Available:
              {{ showtime.seats.filter(s => s.status === 'available').length }}
              / {{ showtime.seats.length }} seats
            </p>
          </div>
          <RouterLink
            :to="`/showtimes/${showtime.id}`"
            class="ml-6 shrink-0 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold px-4 py-2 rounded-lg transition"
          >
            Select Seats
          </RouterLink>
        </div>
      </div>
    </main>
  </div>
</template>