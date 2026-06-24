export function formatTime(ts: string | null | undefined): string {
  if (!ts) return ''
  const d = new Date(ts)
  const diff = Date.now() - d.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h`
  const days = Math.floor(hours / 24)
  if (days < 7) return d.toLocaleDateString('en', { weekday: 'short' })
  return d.toLocaleDateString('en', { month: 'short', day: 'numeric' })
}

export function formatDate(sentAt: string | null, receivedAt: string | null): string {
  const ts = sentAt ?? receivedAt
  if (!ts) return ''
  return new Date(ts).toLocaleString('en', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function attachmentUrl(accountId: string | null, messageId: string, attachmentId: string): string {
  if (accountId) {
    return `/api/v1/accounts/${accountId}/messages/${messageId}/attachments/${attachmentId}`
  }
  return `/api/v1/trap/messages/${messageId}/attachments/${attachmentId}`
}

export function senderName(fromAddr: string): string {
  if (!fromAddr) return 'Unknown'
  const match = fromAddr.match(/^"?([^"<]+)"?\s*</)
  if (match) return match[1].trim()
  return fromAddr
}
