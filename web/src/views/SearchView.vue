<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Search } from '@lucide/vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import MessageList from '@/components/messages/MessageList.vue'
import MessageDetail from '@/components/message/MessageDetail.vue'
import { useMailStore } from '@/stores/mail'
import { useDebounce } from '@/composables/useDebounce'

const props = defineProps<{ messageId?: string }>()
const router = useRouter()
const mailStore = useMailStore()

const query = ref(mailStore.searchQuery)
const debouncedQuery = useDebounce(query)

const results = computed(() => mailStore.searchResults)
const loading = computed(() => mailStore.searchLoading)
const selectedId = computed(() => props.messageId ?? null)

watch(debouncedQuery, q => mailStore.search(q))

watch(
  () => props.messageId,
  (id) => {
    if (id) {
      const msg = results.value.find(m => m.id === id)
      if (msg) mailStore.selectMessage(id, msg.source, msg.account_id ?? undefined)
    }
  },
  { immediate: true },
)

function onSelect(id: string) {
  router.push(`/dashboard/search/${id}`)
}

function onStar(id: string, starred: boolean) {
  const msg = results.value.find(m => m.id === id)
  if (msg) mailStore.starMessage(id, msg.account_id, starred)
}
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset class="flex flex-col min-h-0">
      <header class="flex items-center h-12 px-3 gap-2 border-b border-border shrink-0">
        <SidebarTrigger />
        <Separator orientation="vertical" class="h-4" />
        <span class="text-sm font-medium">Search</span>
      </header>

      <div class="flex flex-1 min-h-0">
        <div class="w-80 shrink-0 border-r border-border flex flex-col min-h-0">
          <!-- Search input -->
          <div class="p-3 border-b border-border shrink-0">
            <div class="relative">
              <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                v-model="query"
                placeholder="Search messages…"
                class="pl-8"
              />
            </div>
          </div>

          <MessageList
            :messages="results"
            :loading="loading"
            :selected-id="selectedId"
            :empty-text="query.trim() ? 'No results' : 'Type to search all messages'"
            @select="onSelect"
            @star="onStar"
          />
        </div>
        <div class="flex-1 min-w-0 min-h-0">
          <MessageDetail />
        </div>
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
