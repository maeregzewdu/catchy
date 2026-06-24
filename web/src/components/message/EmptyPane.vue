<script setup lang="ts">
import { Mail } from '@lucide/vue'

defineProps<{ text?: string }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center h-full gap-3 text-muted-foreground">
    <Mail class="h-12 w-12 opacity-20" />
    <p class="text-sm">{{ text ?? 'Select a message to read it' }}</p>
  </div>
</template>
