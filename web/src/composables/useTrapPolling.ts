import { onMounted, onUnmounted } from 'vue'
import { useMailStore } from '@/stores/mail'

export function useTrapPolling(intervalMs = 3000) {
  const mailStore = useMailStore()
  let timer: ReturnType<typeof setInterval>

  onMounted(() => {
    mailStore.fetchTrapMessages()
    timer = setInterval(() => mailStore.fetchTrapMessages(), intervalMs)
  })

  onUnmounted(() => clearInterval(timer))
}
