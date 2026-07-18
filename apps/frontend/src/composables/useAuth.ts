import { onAuthStateChanged, signInWithPopup, signOut } from 'firebase/auth'
import { auth, googleProvider } from '../firebase'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

export function useAuth() {
  const authStore = useAuthStore()
  const router = useRouter()

  async function login(): Promise<void> {
    try {
      const result = await signInWithPopup(auth, googleProvider)
      const token = await result.user.getIdToken()

      authStore.setUser(result.user, token)
      router.push('/')

    } catch (error) {
      console.error('Login failed: ', error)
    }
  }

  async function logout(): Promise<void> {
    await signOut(auth)
    authStore.clearUser()
    router.push('/login')
  }

  // called when app started
  function initAuth(): void {
    onAuthStateChanged(auth, async (firebaseUser) => {
      if (firebaseUser) {
        const token = await firebaseUser.getIdToken()
        authStore.setUser(firebaseUser, token)
      } else {
        authStore.clearUser()
      }
    })
  }

  return { login, logout, initAuth }
}