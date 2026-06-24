<script setup lang="ts">
import { computed } from 'vue'
import { Star } from '@lucide/vue'
import { cn } from '@/lib/utils'
import { formatTime, senderName } from '@/lib/helpers'
import type { Message } from '@/types'

const props = defineProps<{
  message: Message
  isSelected: boolean
}>()

const emit = defineEmits<{
  select: [id: string]
  star: [id: string, starred: boolean]
}>()

const sender = computed(() => senderName(props.message.from_addr))
const time = computed(() => formatTime(props.message.sent_at ?? props.message.received_at))
</script>

<template>
  <button
    type="button"
    :class="cn(
      'w-full text-left px-3 py-3 flex flex-col gap-1 border-b border-border transition-colors hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
      isSelected ? 'bg-accent' : '',
      !message.is_read ? 'font-medium' : 'text-muted-foreground',
    )"
    @click="emit('select', message.id)"
  >
    <div class="flex items-center gap-2 min-w-0">
      <span
        v-if="!message.is_read"
        class="h-2 w-2 rounded-full bg-primary shrink-0"
      />
      <span v-else class="h-2 w-2 shrink-0" />
      <span class="flex-1 truncate text-sm text-foreground">{{ sender }}</span>
      <span class="text-xs text-muted-foreground shrink-0">{{ time }}</span>
      <button
        type="button"
        class="shrink-0 -mr-1 p-0.5 rounded hover:text-yellow-400 transition-colors"
        :class="message.is_starred ? 'text-yellow-400' : 'text-muted-foreground/40'"
        @click.stop="emit('star', message.id, !message.is_starred)"
      >
        <Star class="h-3.5 w-3.5" :fill="message.is_starred ? 'currentColor' : 'none'" />
      </button>
    </div>
    <p class="truncate text-sm pl-4" :class="isSelected ? 'text-foreground' : ''">
      {{ message.subject || '(no subject)' }}
    </p>
    <p class="truncate text-xs text-muted-foreground pl-4 line-clamp-1">
      {{ message.body_text?.slice(0, 80) }}
    </p>
  </button>
</template>
