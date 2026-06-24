import type {
  Account,
  Attachment,
  CreateAccountPayload,
  HealthResponse,
  Message,
  MessageDetail,
  VerifyResult,
} from '@/types'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: 'UNKNOWN' }))
    throw Object.assign(new Error(err.error ?? res.statusText), { code: err.code, status: res.status })
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

// ── Health ──────────────────────────────────────────────────────────────────

export const getHealth = () => request<HealthResponse>('/api/v1/health')

// ── Trap messages ───────────────────────────────────────────────────────────

export const listTrapMessages = () => request<Message[]>('/api/v1/trap/messages')

export const getTrapMessage = (id: string) =>
  request<MessageDetail>(`/api/v1/trap/messages/${id}`)

export const getTrapRawMIME = (id: string) =>
  fetch(`/api/v1/trap/messages/${id}/raw`).then(r => r.text())

export const patchTrapMessage = (id: string, patch: { is_read?: boolean; is_starred?: boolean }) =>
  request<void>(`/api/v1/trap/messages/${id}`, { method: 'PATCH', body: JSON.stringify(patch) })

export const deleteTrapMessage = (id: string) =>
  request<void>(`/api/v1/trap/messages/${id}`, { method: 'DELETE' })

export const clearTrapMessages = () =>
  request<void>('/api/v1/trap/messages', { method: 'DELETE' })

// ── Accounts ────────────────────────────────────────────────────────────────

export const listAccounts = () => request<Account[]>('/api/v1/accounts')

export const createAccount = (data: CreateAccountPayload) =>
  request<Account>('/api/v1/accounts', { method: 'POST', body: JSON.stringify(data) })

export const deleteAccount = (id: string) =>
  request<void>(`/api/v1/accounts/${id}`, { method: 'DELETE' })

export const verifyAccount = (id: string) =>
  request<VerifyResult>(`/api/v1/accounts/${id}/verify`, { method: 'POST' })

export const syncAccount = (id: string) =>
  request<{ status: string }>(`/api/v1/accounts/${id}/sync`, { method: 'POST' })

// ── Account messages ────────────────────────────────────────────────────────

export const listMessages = (accountId: string, folder = 'INBOX', limit = 50) =>
  request<Message[]>(`/api/v1/accounts/${accountId}/messages?folder=${encodeURIComponent(folder)}&limit=${limit}`)

export const getMessage = (accountId: string, msgId: string) =>
  request<Message>(`/api/v1/accounts/${accountId}/messages/${msgId}`)

export const listMessageAttachments = (accountId: string, msgId: string) =>
  request<Attachment[]>(`/api/v1/accounts/${accountId}/messages/${msgId}/attachments`)

export const getMessageRawMIME = (accountId: string, msgId: string) =>
  fetch(`/api/v1/accounts/${accountId}/messages/${msgId}/raw`).then(r => r.text())

export const patchMessage = (accountId: string, msgId: string, patch: { is_read?: boolean; is_starred?: boolean }) =>
  request<void>(`/api/v1/accounts/${accountId}/messages/${msgId}`, { method: 'PATCH', body: JSON.stringify(patch) })

export const deleteMessage = (accountId: string, msgId: string) =>
  request<void>(`/api/v1/accounts/${accountId}/messages/${msgId}`, { method: 'DELETE' })

// ── Search ──────────────────────────────────────────────────────────────────

export const searchMessages = (q: string, accountId?: string, source?: string) => {
  const params = new URLSearchParams({ q })
  if (accountId) params.set('account', accountId)
  if (source) params.set('source', source)
  return request<Message[]>(`/api/v1/search?${params}`)
}
