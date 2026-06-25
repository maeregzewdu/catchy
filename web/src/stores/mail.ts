import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as api from '@/api'
import type { Attachment, Message, MessageDetail } from '@/types'

export const useMailStore = defineStore('mail', () => {
  // Trap inbox
  const trapMessages = ref<Message[]>([])
  const trapLoading = ref(false)
  const trapInitialized = ref(false)

  // Per-account per-folder messages; key = `${accountId}:${folder}`
  const folderMessages = ref<Record<string, Message[]>>({})
  const folderLoading = ref<Record<string, boolean>>({})

  // Currently selected message detail
  const selectedMessageId = ref<string | null>(null)
  const messageDetail = ref<MessageDetail | null>(null)
  const messageDetailLoading = ref(false)
  const messageDetailError = ref<string | null>(null)

  // Search
  const searchResults = ref<Message[]>([])
  const searchLoading = ref(false)
  const searchQuery = ref('')

  const trapUnreadCount = computed(
    () => trapMessages.value.filter(m => !m.is_read).length,
  )

  // ── Trap ──────────────────────────────────────────────────────────────────

  async function fetchTrapMessages() {
    if (!trapInitialized.value) trapLoading.value = true
    try {
      trapMessages.value = await api.listTrapMessages()
      trapInitialized.value = true
    } catch {
      // Silently ignore polling errors
    } finally {
      trapLoading.value = false
    }
  }

  async function clearTrapMessages() {
    await api.clearTrapMessages()
    trapMessages.value = []
    if (messageDetail.value?.account_id === null) {
      messageDetail.value = null
      selectedMessageId.value = null
    }
  }

  // ── Account folder ────────────────────────────────────────────────────────

  async function fetchFolderMessages(accountId: string, folder: string) {
    const key = `${accountId}:${folder}`
    folderLoading.value[key] = true
    try {
      folderMessages.value[key] = await api.listMessages(accountId, folder)
    } finally {
      folderLoading.value[key] = false
    }
  }

  // ── Message detail ────────────────────────────────────────────────────────

  async function selectMessage(
    id: string,
    source: 'trap' | 'imap' | 'sent',
    accountId?: string,
  ) {
    if (selectedMessageId.value === id) return
    selectedMessageId.value = id
    messageDetailLoading.value = true
    messageDetailError.value = null
    messageDetail.value = null

    try {
      let msg: Message
      let attachments: Attachment[] = []

      if (source === 'trap') {
        const full = await api.getTrapMessage(id)
        msg = full
        attachments = full.attachments ?? []
      } else {
        if (!accountId) throw new Error('accountId required for imap messages')
        msg = await api.getMessage(accountId, id)
        attachments = await api.listMessageAttachments(accountId, id).catch(() => [])
      }

      messageDetail.value = { ...msg, attachments }

      // Mark as read (optimistic)
      if (!msg.is_read) {
        _updateMessageInLists(id, { is_read: true })
        if (source === 'trap') {
          api.patchTrapMessage(id, { is_read: true }).catch(() => {})
        } else if (accountId) {
          api.patchMessage(accountId, id, { is_read: true }).catch(() => {})
        }
      }
    } catch (e: unknown) {
      messageDetailError.value = e instanceof Error ? e.message : 'Failed to load message'
      if (selectedMessageId.value === id) selectedMessageId.value = null
    } finally {
      messageDetailLoading.value = false
    }
  }

  async function starMessage(id: string, accountId: string | null, starred: boolean) {
    _updateMessageInLists(id, { is_starred: starred })
    if (messageDetail.value?.id === id) {
      messageDetail.value = { ...messageDetail.value, is_starred: starred }
    }
    try {
      if (accountId) {
        await api.patchMessage(accountId, id, { is_starred: starred })
      } else {
        await api.patchTrapMessage(id, { is_starred: starred })
      }
    } catch {
      // Revert on error
      _updateMessageInLists(id, { is_starred: !starred })
      if (messageDetail.value?.id === id) {
        messageDetail.value = { ...messageDetail.value, is_starred: !starred }
      }
    }
  }

  async function deleteSelectedMessage(id: string, accountId: string | null) {
    if (accountId) {
      await api.deleteMessage(accountId, id)
    } else {
      await api.deleteTrapMessage(id)
    }
    _removeMessageFromLists(id)
    if (selectedMessageId.value === id) {
      selectedMessageId.value = null
      messageDetail.value = null
    }
  }

  // ── Search ────────────────────────────────────────────────────────────────

  async function search(q: string, accountId?: string, source?: string) {
    searchQuery.value = q
    if (!q.trim()) {
      searchResults.value = []
      return
    }
    searchLoading.value = true
    try {
      searchResults.value = await api.searchMessages(q, accountId, source)
    } catch {
      searchResults.value = []
    } finally {
      searchLoading.value = false
    }
  }

  // ── Helpers ───────────────────────────────────────────────────────────────

  function _updateMessageInLists(id: string, patch: Partial<Message>) {
    const updateList = (list: Message[]) => {
      const idx = list.findIndex(m => m.id === id)
      if (idx !== -1) list[idx] = { ...list[idx], ...patch }
    }
    updateList(trapMessages.value)
    updateList(searchResults.value)
    for (const key of Object.keys(folderMessages.value)) {
      updateList(folderMessages.value[key])
    }
  }

  function _removeMessageFromLists(id: string) {
    trapMessages.value = trapMessages.value.filter(m => m.id !== id)
    searchResults.value = searchResults.value.filter(m => m.id !== id)
    for (const key of Object.keys(folderMessages.value)) {
      folderMessages.value[key] = folderMessages.value[key].filter(m => m.id !== id)
    }
  }

  return {
    trapMessages,
    trapLoading,
    folderMessages,
    folderLoading,
    selectedMessageId,
    messageDetail,
    messageDetailLoading,
    messageDetailError,
    searchResults,
    searchLoading,
    searchQuery,
    trapUnreadCount,
    fetchTrapMessages,
    clearTrapMessages,
    fetchFolderMessages,
    selectMessage,
    starMessage,
    deleteSelectedMessage,
    search,
  }
})
