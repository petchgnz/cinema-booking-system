import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { User } from 'firebase/auth';

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(null);

  const isLoggedIn = computed(() => !!user.value);

  function setUser(newUser: User, newToken: string): void {
    user.value = newUser;
    token.value = newToken;
  }

  function clearUser(): void {
    user.value = null;
    token.value = null;
  }

  return { user, token, isLoggedIn, setUser, clearUser }
});
