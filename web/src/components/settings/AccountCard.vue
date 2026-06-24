<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Loader2, Trash2, ShieldCheck, RefreshCw, CheckCircle2, XCircle } from '@lucide/vue'
import { useAccountStore } from '@/stores/accounts'
import type { Account } from '@/types'

const props = defineProps<{ account: Account }>()

const accountStore = useAccountStore()

const hasId = computed(() => !!props.account.id)
const verifyLoading = computed(() => accountStore.verifyLoading[props.account.id])
const syncLoading = computed(() => accountStore.syncLoading[props.account.id])
const verifyResult = computed(() => accountStore.verifyResults[props.account.id])

function smtpOk(result: string) {
  return result === 'ok'
}
</script>

<template>
  <Card>
    <CardHeader class="pb-2">
      <div class="flex items-start justify-between">
        <div>
          <CardTitle class="text-sm">{{ account.name }}</CardTitle>
          <CardDescription class="text-xs">{{ account.email }}</CardDescription>
        </div>
        <Badge variant="outline" class="text-xs shrink-0">{{ account.username }}</Badge>
      </div>
    </CardHeader>

    <CardContent class="text-xs text-muted-foreground space-y-1">
      <div class="grid grid-cols-2 gap-x-4 gap-y-1">
        <span class="font-medium text-foreground">SMTP</span>
        <span>{{ account.smtp_host }}:{{ account.smtp_port }}</span>
        <span class="font-medium text-foreground">IMAP</span>
        <span>{{ account.imap_host }}:{{ account.imap_port }}</span>
      </div>

      <!-- Verify result -->
      <template v-if="verifyResult">
        <Separator class="my-2" />
        <div class="space-y-1">
          <div class="flex items-center gap-1.5">
            <CheckCircle2 v-if="smtpOk(verifyResult.smtp)" class="h-3.5 w-3.5 text-green-500" />
            <XCircle v-else class="h-3.5 w-3.5 text-destructive" />
            <span>SMTP: {{ verifyResult.smtp }}</span>
          </div>
          <div class="flex items-center gap-1.5">
            <CheckCircle2 v-if="smtpOk(verifyResult.imap)" class="h-3.5 w-3.5 text-green-500" />
            <XCircle v-else class="h-3.5 w-3.5 text-destructive" />
            <span>IMAP: {{ verifyResult.imap }}</span>
          </div>
        </div>
      </template>
    </CardContent>

    <CardFooter class="gap-2 flex-wrap">
      <Button
        variant="outline"
        size="sm"
        :disabled="verifyLoading || !hasId"
        @click="accountStore.verifyAccount(account.id)"
      >
        <Loader2 v-if="verifyLoading" class="h-3.5 w-3.5 animate-spin" />
        <ShieldCheck v-else class="h-3.5 w-3.5" />
        Verify
      </Button>
      <Button
        variant="outline"
        size="sm"
        :disabled="syncLoading || !hasId"
        @click="accountStore.syncAccount(account.id)"
      >
        <Loader2 v-if="syncLoading" class="h-3.5 w-3.5 animate-spin" />
        <RefreshCw v-else class="h-3.5 w-3.5" />
        Sync
      </Button>
      <Button
        variant="destructive"
        size="sm"
        class="ml-auto"
        :disabled="!hasId"
        @click="accountStore.deleteAccount(account.id)"
      >
        <Trash2 class="h-3.5 w-3.5" />
        Delete
      </Button>
    </CardFooter>
  </Card>
</template>
