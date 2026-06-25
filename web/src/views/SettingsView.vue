<script setup lang="ts">
import { computed } from 'vue'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Server } from '@lucide/vue'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import AccountCard from '@/components/settings/AccountCard.vue'
import AddAccountForm from '@/components/settings/AddAccountForm.vue'
import { useAccountStore } from '@/stores/accounts'
import { useAppStore } from '@/stores/app'

const accountStore = useAccountStore()
const appStore = useAppStore()
const trapAddr = computed(() => appStore.health?.trap_addr ?? 'localhost:1025')

const smtpPort = computed(() => {
  const [, port] = trapAddr.value.split(':')
  return port ?? '1025'
})
</script>

<template>
  <DashboardLayout :breadcrumbs="[{ label: 'Settings' }]">
      <ScrollArea class="flex-1">
        <div class="max-w-2xl mx-auto p-6 space-y-8">

          <!-- SMTP Trap info -->
          <section class="space-y-3">
            <h2 class="text-sm font-semibold">SMTP Trap</h2>
            <Card>
              <CardHeader class="pb-2">
                <CardTitle class="text-sm flex items-center gap-2">
                  <Server class="h-4 w-4" />
                  {{ trapAddr }}
                </CardTitle>
                <CardDescription>No authentication required. Accepts all email unconditionally.</CardDescription>
              </CardHeader>
              <CardContent class="text-xs space-y-2">
                <p class="text-muted-foreground">Point your app's SMTP settings here to catch outgoing mail:</p>
                <pre class="bg-muted rounded p-3 font-mono overflow-x-auto">MAIL_MAILER=smtp
MAIL_HOST=localhost
MAIL_PORT={{ smtpPort }}
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null</pre>
              </CardContent>
            </Card>
          </section>

          <Separator />

          <!-- Accounts list -->
          <section class="space-y-3">
            <h2 class="text-sm font-semibold">IMAP Accounts</h2>
            <p v-if="accountStore.loading" class="text-sm text-muted-foreground">Loading…</p>
            <p v-else-if="accountStore.accounts.length === 0" class="text-sm text-muted-foreground">
              No accounts configured yet.
            </p>
            <div v-else class="space-y-3">
              <AccountCard
                v-for="account in accountStore.accounts"
                :key="account.id"
                :account="account"
              />
            </div>
          </section>

          <Separator />

          <!-- Add account form -->
          <section class="space-y-3">
            <h2 class="text-sm font-semibold">Add IMAP Account</h2>
            <AddAccountForm />
          </section>

        </div>
      </ScrollArea>
  </DashboardLayout>
</template>
