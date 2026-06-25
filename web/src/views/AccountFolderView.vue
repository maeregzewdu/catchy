<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { RefreshCw, Loader2 } from '@lucide/vue'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import MessageList from '@/components/messages/MessageList.vue'
import MessageDetail from '@/components/message/MessageDetail.vue'
import { useMailStore } from '@/stores/mail'
import { useAccountStore } from '@/stores/accounts'

const props = defineProps<{
  accountId: string
  folder: string
  messageId?: string
}>()

const router = useRouter()
const mailStore = useMailStore()
const accountStore = useAccountStore()

const folderKey = computed(() => `${props.accountId}:${props.folder}`)
const messages = computed(() => mailStore.folderMessages[folderKey.value] ?? [])
const loading = computed(() => mailStore.folderLoading[folderKey.value] ?? false)
const syncLoading = computed(() => accountStore.syncLoading[props.accountId])
const account = computed(() => accountStore.accounts.find(a => a.id === props.accountId))
const selectedId = computed(() => props.messageId ?? null)

const breadcrumbs = computed(() => [
  {
    label: account.value?.name ?? 'Account',
    to: `/dashboard/accounts/${props.accountId}/INBOX`,
  },
  { label: props.folder },
])

watch(
  [() => props.accountId, () => props.folder],
  () => {
    mailStore.fetchFolderMessages(props.accountId, props.folder)
  },
  { immediate: true },
)

watch(
  () => props.messageId,
  (id) => {
    if (id) {
      mailStore.selectMessage(id, 'imap', props.accountId)
    }
  },
  { immediate: true },
)

function onSelect(id: string) {
  router.push(`/dashboard/accounts/${props.accountId}/${props.folder}/${id}`)
}

function onStar(id: string, starred: boolean) {
  mailStore.starMessage(id, props.accountId, starred)
}

async function sync() {
  await accountStore.syncAccount(props.accountId)
  await mailStore.fetchFolderMessages(props.accountId, props.folder)
}
</script>

<template>
  <DashboardLayout :breadcrumbs="breadcrumbs">
    <template #actions>
      <span v-if="messages.length" class="text-xs text-muted-foreground mr-1">{{ messages.length }} messages</span>
      <Button variant="ghost" size="icon-sm" :disabled="syncLoading" @click="sync">
        <Loader2 v-if="syncLoading" class="h-4 w-4 animate-spin" />
        <RefreshCw v-else class="h-4 w-4" />
      </Button>
    </template>

    <div class="flex flex-1 min-h-0">
      <div class="w-80 shrink-0 border-r border-border flex flex-col min-h-0">
        <MessageList
          :messages="messages"
          :loading="loading"
          :selected-id="selectedId"
          :empty-title="`${folder} is empty`"
          empty-description="No messages in this folder yet."
          @select="onSelect"
          @star="onStar"
        />
      </div>
      <div class="flex-1 min-w-0 min-h-0">
        <MessageDetail />
      </div>
    </div>
  </DashboardLayout>
</template>
