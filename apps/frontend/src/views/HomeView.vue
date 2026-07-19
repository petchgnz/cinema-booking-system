<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuth } from '../composables/useAuth';
import { useAuthStore } from '../stores/auth';
import { getMovies } from '../api/movies';
import type { Movie } from '../types';

const { logout } = useAuth();
const authStore = useAuthStore();

const movies = ref<Movie[]>([]);
const loading = ref<boolean>(true);

onMounted(async () => {
  try {
    const res = await getMovies();
    movies.value = res.data;
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="min-h-screen bg-gray-900 text-white">
    <nav class="bg-gray-800 px-6 py-4 flex justify-between items-center">
      <h1 class="text-xl font-bold">Cinema Booking</h1>
      <div class="flex items-center gap-4">
        <span class="text-gray-400 text-sm">{{
          authStore.user?.displayName
        }}</span>
        <button
          @click="logout"
          class="text-sm bg-gray-700 px-4 py-2 rounded-lg hover:bg-gray-600 transition"
        >
          Logout
        </button>
      </div>
    </nav>

    <main class="max-w-4xl mx-auto px-6 py-10">
      <h2 class="text-2xl font-semibold mb-6">Now Showing</h2>

      <div
        v-if="loading"
        class="text-gray-400"
      >
        Loading...
      </div>

      <div
        v-else
        class="grid gap-5"
      >
        <div
          v-for="movie in movies"
          :key="movie.id"
          class="flex items-center justify-between rounded-2xl border border-gray-700 bg-gray-800 p-5 shadow-lg hover:border-indigo-500 hover:shadow-xl transition"
        >
          <div class="flex items-start gap-5 flex-1">
            <!-- Poster -->
            <img
              v-if="movie.poster_url"
              :src="movie.poster_url"
              class="h-28 w-20 rounded-lg object-cover shrink-0"
            />

            <div
              v-else
              class="flex h-28 w-20 shrink-0 items-center justify-center rounded-lg bg-gray-700"
            >
              <span class="text-xs text-gray-400"> No Image </span>
            </div>

            <!-- Content -->
            <div class="flex flex-col flex-1 min-w-0">
              <h3 class="text-xl font-bold text-white">
                {{ movie.title }}
              </h3>

              <span
                class="mt-2 inline-flex w-fit rounded-full bg-indigo-600/20 px-2 py-1 text-xs text-indigo-300"
              >
                {{ movie.duration }} minutes
              </span>

              <p class="mt-4 text-sm leading-6 text-gray-300 line-clamp-3">
                {{ movie.description }}
              </p>
            </div>
          </div>

          <!-- Action -->
          <RouterLink
            :to="`/movies/${movie.id}/showtimes`"
            class="ml-6 shrink-0 rounded-xl bg-indigo-600 px-5 py-3 font-semibold text-white transition hover:bg-indigo-500"
          >
            Book Now
          </RouterLink>
        </div>
      </div>
    </main>
  </div>
</template>
