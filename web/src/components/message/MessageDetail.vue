<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Star, Trash2, Download } from '@lucide/vue'
import EmptyPane from './EmptyPane.vue'
import { useMailStore } from '@/stores/mail'
import { formatDate, formatSize, attachmentUrl } from '@/lib/helpers'

const router = useRouter()
const mailStore = useMailStore()

const msg = computed(() => mailStore.messageDetail)
const loading = computed(() => mailStore.messageDetailLoading)
const error = computed(() => mailStore.messageDetailError)

async function toggleStar() {
  if (!msg.value) return
  await mailStore.starMessage(msg.value.id, msg.value.account_id, !msg.value.is_starred)
}

async function deleteMessage() {
  if (!msg.value) return
  await mailStore.deleteSelectedMessage(msg.value.id, msg.value.account_id)
  router.back()
}
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Loading -->
    <div v-if="loading" class="p-6 space-y-4">
      <Skeleton class="h-6 w-3/4" />
      <Skeleton class="h-4 w-1/2" />
      <Skeleton class="h-4 w-1/3" />
      <Separator />
      <Skeleton class="h-48 w-full" />
    </div>

    <!-- Error -->
    <div v-else-if="error" class="flex items-center justify-center h-full text-destructive text-sm p-4 text-center">
      {{ error }}
    </div>

    <!-- Empty -->
    <EmptyPane v-else-if="!msg" />

    <!-- Message -->
    <template v-else>
      <!-- Header -->
      <div class="p-4 border-b border-border shrink-0">
        <div class="flex items-start justify-between gap-3">
          <h2 class="text-base font-semibold leading-tight break-words flex-1">
            {{ msg.subject || '(no subject)' }}
          </h2>
          <div class="flex items-center gap-1 shrink-0">
            <Button variant="ghost" size="icon-sm" @click="toggleStar">
              <Star class="h-4 w-4" :class="msg.is_starred ? 'text-yellow-400 fill-yellow-400' : ''" />
            </Button>
            <Button variant="ghost" size="icon-sm" @click="deleteMessage">
              <Trash2 class="h-4 w-4" />
            </Button>
          </div>
        </div>
        <div class="mt-2 space-y-0.5 text-xs text-muted-foreground">
          <div><span class="font-medium text-foreground">From:</span> {{ msg.from_addr }}</div>
          <div><span class="font-medium text-foreground">To:</span> {{ msg.to_addrs?.join(', ') }}</div>
          <div v-if="msg.cc_addrs?.length"><span class="font-medium text-foreground">Cc:</span> {{ msg.cc_addrs.join(', ') }}</div>
          <div><span class="font-medium text-foreground">Date:</span> {{ formatDate(msg.sent_at, msg.received_at) }}</div>
        </div>
      </div>

      <!-- Tabs -->
      <Tabs default-value="rendered" class="flex flex-col flex-1 min-h-0">
        <TabsList class="shrink-0 px-4 justify-start rounded-none border-b border-border bg-transparent h-10 gap-1">
          <TabsTrigger value="rendered" class="text-xs h-8">Preview</TabsTrigger>
          <TabsTrigger value="text" class="text-xs h-8">Text</TabsTrigger>
          <TabsTrigger value="headers" class="text-xs h-8">Headers</TabsTrigger>
          <TabsTrigger
            v-if="msg.attachments?.length"
            value="attachments"
            class="text-xs h-8"
          >
            Attachments ({{ msg.attachments.length }})
          </TabsTrigger>
        </TabsList>

        <!-- Rendered HTML / text -->
        <TabsContent value="rendered" class="flex-1 min-h-0 m-0 p-0">
          <iframe
            v-if="msg.body_html"
            :srcdoc="msg.body_html"
            sandbox="allow-same-origin allow-popups"
            referrerpolicy="no-referrer"
            class="w-full h-full border-0"
          />
          <ScrollArea v-else class="h-full">
            <pre class="p-4 text-xs whitespace-pre-wrap font-mono text-foreground">{{ msg.body_text || '(empty)' }}</pre>
          </ScrollArea>
        </TabsContent>

        <!-- Plain text -->
        <TabsContent value="text" class="flex-1 min-h-0 m-0">
          <ScrollArea class="h-full">
            <pre class="p-4 text-xs whitespace-pre-wrap font-mono text-foreground">{{ msg.body_text || '(empty)' }}</pre>
          </ScrollArea>
        </TabsContent>

        <!-- Headers -->
        <TabsContent value="headers" class="flex-1 min-h-0 m-0">
          <ScrollArea class="h-full">
            <div class="p-4 space-y-1 text-xs font-mono">
              <div class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5">
                <span class="text-muted-foreground font-medium">Message-ID:</span>
                <span class="break-all">{{ msg.message_id }}</span>
                <span class="text-muted-foreground font-medium">From:</span>
                <span class="break-all">{{ msg.from_addr }}</span>
                <span class="text-muted-foreground font-medium">To:</span>
                <span class="break-all">{{ msg.to_addrs?.join(', ') }}</span>
                <template v-if="msg.cc_addrs?.length">
                  <span class="text-muted-foreground font-medium">Cc:</span>
                  <span class="break-all">{{ msg.cc_addrs.join(', ') }}</span>
                </template>
                <template v-if="msg.bcc_addrs?.length">
                  <span class="text-muted-foreground font-medium">Bcc:</span>
                  <span class="break-all">{{ msg.bcc_addrs.join(', ') }}</span>
                </template>
                <span class="text-muted-foreground font-medium">Subject:</span>
                <span class="break-all">{{ msg.subject }}</span>
                <span class="text-muted-foreground font-medium">Sent:</span>
                <span>{{ msg.sent_at }}</span>
                <span class="text-muted-foreground font-medium">Received:</span>
                <span>{{ msg.received_at }}</span>
                <span class="text-muted-foreground font-medium">Folder:</span>
                <span>{{ msg.folder }}</span>
                <span class="text-muted-foreground font-medium">Source:</span>
                <span>{{ msg.source }}</span>
              </div>
            </div>
          </ScrollArea>
        </TabsContent>

        <!-- Attachments -->
        <TabsContent value="attachments" class="flex-1 min-h-0 m-0">
          <ScrollArea class="h-full">
            <div class="p-4 space-y-2">
              <a
                v-for="att in msg.attachments"
                :key="att.id"
                :href="attachmentUrl(msg.account_id, msg.id, att.id)"
                download
                class="flex items-center gap-3 p-3 rounded-lg border border-border hover:bg-accent transition-colors text-sm"
              >
                <Download class="h-4 w-4 text-muted-foreground shrink-0" />
                <div class="flex-1 min-w-0">
                  <p class="truncate font-medium">{{ att.filename }}</p>
                  <p class="text-xs text-muted-foreground">{{ att.mime_type }} · {{ formatSize(att.size) }}</p>
                </div>
              </a>
            </div>
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </template>
  </div>
</template>
