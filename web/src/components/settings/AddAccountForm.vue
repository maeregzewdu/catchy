<script setup lang="ts">
import { reactive, ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Loader2, PlusCircle } from '@lucide/vue'
import { useAccountStore } from '@/stores/accounts'

const accountStore = useAccountStore()

const loading = ref(false)
const error = ref<string | null>(null)
const success = ref(false)

const form = reactive({
  name: '',
  email: '',
  smtp_host: '',
  smtp_port: 587,
  imap_host: '',
  imap_port: 993,
  username: '',
  password: '',
})

async function submit() {
  loading.value = true
  error.value = null
  success.value = false
  try {
    await accountStore.createAccount({ ...form })
    success.value = true
    Object.assign(form, {
      name: '', email: '', smtp_host: '', smtp_port: 587,
      imap_host: '', imap_port: 993, username: '', password: '',
    })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to add account'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <form class="space-y-4" @submit.prevent="submit">
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div class="space-y-1.5">
        <Label for="name">Account name</Label>
        <Input id="name" v-model="form.name" placeholder="My Gmail" required />
      </div>
      <div class="space-y-1.5">
        <Label for="email">Email address</Label>
        <Input id="email" v-model="form.email" type="email" placeholder="you@example.com" required />
      </div>
      <div class="space-y-1.5">
        <Label for="smtp_host">SMTP host</Label>
        <Input id="smtp_host" v-model="form.smtp_host" placeholder="smtp.gmail.com" required />
      </div>
      <div class="space-y-1.5">
        <Label for="smtp_port">SMTP port</Label>
        <Input id="smtp_port" v-model.number="form.smtp_port" type="number" required />
      </div>
      <div class="space-y-1.5">
        <Label for="imap_host">IMAP host</Label>
        <Input id="imap_host" v-model="form.imap_host" placeholder="imap.gmail.com" required />
      </div>
      <div class="space-y-1.5">
        <Label for="imap_port">IMAP port</Label>
        <Input id="imap_port" v-model.number="form.imap_port" type="number" required />
      </div>
      <div class="space-y-1.5">
        <Label for="username">Username</Label>
        <Input id="username" v-model="form.username" placeholder="you@example.com" required />
      </div>
      <div class="space-y-1.5">
        <Label for="password">Password / App password</Label>
        <Input id="password" v-model="form.password" type="password" required />
      </div>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <p v-if="success" class="text-sm text-green-600 dark:text-green-400">Account added successfully.</p>

    <Button type="submit" :disabled="loading">
      <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
      <PlusCircle v-else class="h-4 w-4" />
      Add Account
    </Button>
  </form>
</template>
