import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '@/api'
import type { Account, CreateAccountPayload, VerifyResult } from '@/types'

export const useAccountStore = defineStore('accounts', () => {
  const accounts = ref<Account[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const verifyResults = ref<Record<string, VerifyResult>>({})
  const verifyLoading = ref<Record<string, boolean>>({})
  const syncLoading = ref<Record<string, boolean>>({})

  async function fetchAccounts() {
    loading.value = true
    error.value = null
    try {
      accounts.value = await api.listAccounts()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Failed to load accounts'
    } finally {
      loading.value = false
    }
  }

  async function createAccount(data: CreateAccountPayload): Promise<Account> {
    const account = await api.createAccount(data)
    await fetchAccounts()
    return account
  }

  async function deleteAccount(id: string) {
    if (!id) throw new Error('This account has no ID and cannot be deleted')
    await api.deleteAccount(id)
    accounts.value = accounts.value.filter(a => a.id !== id)
    delete verifyResults.value[id]
  }

  async function verifyAccount(id: string) {
    verifyLoading.value[id] = true
    try {
      verifyResults.value[id] = await api.verifyAccount(id)
    } catch (e: unknown) {
      verifyResults.value[id] = {
        smtp: e instanceof Error ? e.message : 'error',
        imap: 'skipped',
      }
    } finally {
      verifyLoading.value[id] = false
    }
  }

  async function syncAccount(id: string) {
    syncLoading.value[id] = true
    try {
      await api.syncAccount(id)
    } finally {
      syncLoading.value[id] = false
    }
  }

  return {
    accounts,
    loading,
    error,
    verifyResults,
    verifyLoading,
    syncLoading,
    fetchAccounts,
    createAccount,
    deleteAccount,
    verifyAccount,
    syncAccount,
  }
})
