export interface Account {
  id: string
  name: string
  email: string
  smtp_host: string
  smtp_port: number
  imap_host: string
  imap_port: number
  username: string
  created_at: string
}

export interface Message {
  id: string
  account_id: string | null
  source: 'trap' | 'imap' | 'sent'
  message_id: string
  folder: string
  subject: string
  from_addr: string
  to_addrs: string[]
  cc_addrs: string[]
  bcc_addrs: string[]
  body_text: string
  body_html: string
  is_read: boolean
  is_starred: boolean
  sent_at: string | null
  received_at: string | null
  created_at: string
}

export interface MessageDetail extends Message {
  attachments: Attachment[]
}

export interface Attachment {
  id: string
  message_id: string
  filename: string
  mime_type: string
  size: number
}

export interface HealthResponse {
  status: string
  version: string
  trap_addr: string
}

export interface ApiError {
  error: string
  code: string
}

export interface VerifyResult {
  smtp: string
  imap: string
}

export interface CreateAccountPayload {
  name: string
  email: string
  smtp_host: string
  smtp_port: number
  imap_host: string
  imap_port: number
  username: string
  password: string
}
