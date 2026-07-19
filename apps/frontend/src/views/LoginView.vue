<script setup lang="ts">
import { useAuth } from '../composables/useAuth'

const { login } = useAuth()
</script>

<template>
  <div class="min-h-screen bg-gray-900 flex items-center justify-center">
    <div class="bg-gray-800 p-10 rounded-2xl text-center space-y-6">
      <h1 class="text-3xl font-bold text-white">Cinema Booking</h1>
      <p class="text-gray-400">Sign in to book your seats</p>
      <button
        @click="login"
        class="bg-white text-gray-900 font-semibold px-6 py-3 rounded-lg hover:bg-gray-100 transition"
      >
        Sign in with Google
      </button>
    </div>
  </div>
</template>