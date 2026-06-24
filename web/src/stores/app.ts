import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getHealth } from '@/api'
import type { HealthResponse } from '@/types'

export const useAppStore = defineStore('app', () => {
  const health = ref<HealthResponse | null>(null)
  const healthLoading = ref(false)
  const healthError = ref<string | null>(null)

  async function fetchHealth() {
    healthLoading.value = true
    healthError.value = null
    try {
      health.value = await getHealth()
    } catch (e: unknown) {
      healthError.value = e instanceof Error ? e.message : 'Unreachable'
    } finally {
      healthLoading.value = false
    }
  }

  return { health, healthLoading, healthError, fetchHealth }
})
