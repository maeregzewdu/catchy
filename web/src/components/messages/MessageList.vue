<script setup lang="ts">
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import MessageRow from './MessageRow.vue'
import type { Message } from '@/types'

defineProps<{
  messages: Message[]
  loading: boolean
  selectedId: string | null
  emptyText?: string
}>()

const emit = defineEmits<{
  select: [id: string]
  star: [id: string, starred: boolean]
}>()
</script>

<template>
  <div class="flex flex-col h-full">
    <ScrollArea class="flex-1">
      <div v-if="loading && messages.length === 0" class="p-3 space-y-2">
        <Skeleton v-for="i in 8" :key="i" class="h-16 w-full rounded-md" />
      </div>

      <div v-else-if="messages.length === 0" class="flex flex-col items-center justify-center h-64 text-muted-foreground text-sm gap-2">
        <span>{{ emptyText ?? 'No messages' }}</span>
      </div>

      <template v-else>
        <MessageRow
          v-for="msg in messages"
          :key="msg.id"
          :message="msg"
          :is-selected="msg.id === selectedId"
          @select="emit('select', $event)"
          @star="(id, starred) => emit('star', id, starred)"
        />
      </template>
    </ScrollArea>
  </div>
</template>
