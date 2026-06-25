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
      'group w-full text-left flex gap-3 px-4 py-4 border-b border-border/60 transition-colors hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-inset',
      isSelected ? 'bg-accent' : '',
    )"
    @click="emit('select', message.id)"
  >
    <div class="flex-1 min-w-0 space-y-1">
      <!-- Top line: sender + time -->
      <div class="flex items-baseline gap-2 min-w-0">
        <span
          class="flex-1 truncate text-sm text-foreground"
          :class="!message.is_read ? 'font-semibold' : 'font-normal'"
        >
          {{ sender }}
        </span>
        <span class="text-xs text-muted-foreground shrink-0 tabular-nums">{{ time }}</span>
      </div>

      <!-- Subject -->
      <p
        class="truncate text-sm"
        :class="!message.is_read ? 'text-foreground font-medium' : 'text-muted-foreground'"
      >
        {{ message.subject || '(no subject)' }}
      </p>

      <!-- Preview -->
      <p class="truncate text-xs text-muted-foreground/80 line-clamp-1">
        {{ message.body_text?.slice(0, 100) }}
      </p>
    </div>

    <!-- Star -->
    <button
      type="button"
      class="shrink-0 self-start p-1 -m-1 rounded-md transition-colors hover:text-yellow-400"
      :class="message.is_starred ? 'text-yellow-400' : 'text-muted-foreground/30 opacity-0 group-hover:opacity-100 focus-visible:opacity-100'"
      @click.stop="emit('star', message.id, !message.is_starred)"
    >
      <Star class="h-4 w-4" :fill="message.is_starred ? 'currentColor' : 'none'" />
    </button>
  </button>
</template>
