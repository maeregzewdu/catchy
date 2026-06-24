<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Trash2 } from '@lucide/vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import MessageList from '@/components/messages/MessageList.vue'
import MessageDetail from '@/components/message/MessageDetail.vue'
import { useMailStore } from '@/stores/mail'
import { useTrapPolling } from '@/composables/useTrapPolling'

const props = defineProps<{ messageId?: string }>()
const router = useRouter()
const mailStore = useMailStore()

useTrapPolling()

const messages = computed(() => mailStore.trapMessages)
const loading = computed(() => mailStore.trapLoading)
const selectedId = computed(() => props.messageId ?? null)

watch(
  () => props.messageId,
  (id) => {
    if (id) {
      const msg = messages.value.find(m => m.id === id)
      if (msg) mailStore.selectMessage(id, 'trap')
    }
  },
  { immediate: true },
)

function onSelect(id: string) {
  router.push(`/dashboard/trap/${id}`)
}

function onStar(id: string, starred: boolean) {
  mailStore.starMessage(id, null, starred)
}

async function clearAll() {
  await mailStore.clearTrapMessages()
  router.replace('/dashboard/trap')
}
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset class="flex flex-col min-h-0">
      <!-- Top bar -->
      <header class="flex items-center h-12 px-3 gap-2 border-b border-border shrink-0">
        <SidebarTrigger />
        <Separator orientation="vertical" class="h-4" />
        <span class="text-sm font-medium">Trap Inbox</span>
        <span v-if="messages.length" class="text-xs text-muted-foreground">({{ messages.length }})</span>
        <div class="ml-auto">
          <Dialog>
            <DialogTrigger as-child>
              <Button variant="ghost" size="icon-sm" :disabled="messages.length === 0">
                <Trash2 class="h-4 w-4" />
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Clear all messages?</DialogTitle>
                <DialogDescription>
                  This will permanently delete all {{ messages.length }} trapped messages.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button variant="destructive" @click="clearAll">Clear All</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </header>

      <!-- Two-pane -->
      <div class="flex flex-1 min-h-0">
        <div class="w-80 shrink-0 border-r border-border flex flex-col min-h-0">
          <MessageList
            :messages="messages"
            :loading="loading"
            :selected-id="selectedId"
            empty-text="No trapped emails yet. Point your app at localhost:1025."
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
