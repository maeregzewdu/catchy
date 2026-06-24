<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
} from '@/components/ui/sidebar'
import { Badge } from '@/components/ui/badge'
import { useAccountStore } from '@/stores/accounts'
import { useMailStore } from '@/stores/mail'
import { Inbox, Search, Settings, Mail, Zap, Send } from '@lucide/vue'

const route = useRoute()
const accountStore = useAccountStore()
const mailStore = useMailStore()

const trapUnread = computed(() => mailStore.trapUnreadCount)

function isActive(path: string) {
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<template>
  <Sidebar collapsible="icon">
    <SidebarHeader class="p-4">
      <div class="flex items-center gap-2">
        <div class="flex h-8 w-8 items-center justify-center rounded-sm bg-primary text-primary-foreground shrink-0">
          <Zap class="h-4 w-4" />
        </div>
        <div class="flex flex-col leading-none group-data-[collapsible=icon]:hidden">
          <span class="font-bold text-sm">Catchy</span>
          <span class="text-xs text-muted-foreground">mail dev tool</span>
        </div>
      </div>
    </SidebarHeader>

    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                as-child
                :is-active="isActive('/dashboard/trap')"
                tooltip="Trap Inbox"
              >
                <RouterLink to="/dashboard/trap" class="flex items-center gap-2 w-full">
                  <Inbox class="h-4 w-4 shrink-0" />
                  <span class="flex-1 truncate">Trap Inbox</span>
                  <Badge
                    v-if="trapUnread > 0"
                    variant="destructive"
                    class="ml-auto h-5 min-w-5 px-1 text-xs"
                  >
                    {{ trapUnread > 99 ? '99+' : trapUnread }}
                  </Badge>
                </RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>

      <template v-for="account in accountStore.accounts" :key="account.id">
        <SidebarGroup>
          <SidebarGroupLabel class="truncate">{{ account.name }}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  as-child
                  :is-active="isActive(`/dashboard/accounts/${account.id}/INBOX`)"
                  tooltip="Inbox"
                >
                  <RouterLink :to="`/dashboard/accounts/${account.id}/INBOX`" class="flex items-center gap-2 w-full">
                    <Mail class="h-4 w-4 shrink-0" />
                    <span class="truncate">Inbox</span>
                  </RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton
                  as-child
                  :is-active="isActive(`/dashboard/accounts/${account.id}/Sent`)"
                  tooltip="Sent"
                >
                  <RouterLink :to="`/dashboard/accounts/${account.id}/Sent`" class="flex items-center gap-2 w-full">
                    <Send class="h-4 w-4 shrink-0" />
                    <span class="truncate">Sent</span>
                  </RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </template>
    </SidebarContent>

    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            as-child
            :is-active="isActive('/dashboard/search')"
            tooltip="Search"
          >
            <RouterLink to="/dashboard/search" class="flex items-center gap-2 w-full">
              <Search class="h-4 w-4 shrink-0" />
              <span class="truncate">Search</span>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
        <SidebarMenuItem>
          <SidebarMenuButton
            as-child
            :is-active="isActive('/dashboard/settings')"
            tooltip="Settings"
          >
            <RouterLink to="/dashboard/settings" class="flex items-center gap-2 w-full">
              <Settings class="h-4 w-4 shrink-0" />
              <span class="truncate">Settings</span>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
  </Sidebar>
</template>
