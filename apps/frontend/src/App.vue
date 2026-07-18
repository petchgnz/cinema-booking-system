<script setup lang="ts">
import { useAuth } from './composables/useAuth'

const { initAuth } = useAuth()
initAuth() // restore sessions when refresh
</script>

<template>
  <RouterView />
</template>
