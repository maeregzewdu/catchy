<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { useAccountStore } from '@/stores/accounts'

const appStore = useAppStore()
const accountStore = useAccountStore()

onMounted(() => {
  appStore.fetchHealth()
  accountStore.fetchAccounts()
})
</script>

<template>
  <RouterView />
</template>
