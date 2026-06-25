<script setup lang="ts">
import { RouterLink } from "vue-router";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import AppSidebar from "@/components/layout/AppSidebar.vue";
import ThemeToggle from "@/components/layout/ThemeToggle.vue";

interface Crumb {
  label: string;
  to?: string;
}

defineProps<{ breadcrumbs?: Crumb[] }>();
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset class="flex flex-col min-h-0">
      <!-- Permanent top bar -->
      <header
        class="flex items-center h-12 px-3 gap-4 border-b border-border shrink-0"
      >
        <SidebarTrigger />

        <Breadcrumb v-if="breadcrumbs?.length">
          <BreadcrumbList>
            <template v-for="(crumb, i) in breadcrumbs" :key="i">
              <BreadcrumbItem>
                <BreadcrumbLink
                  v-if="crumb.to && i < breadcrumbs.length - 1"
                  as-child
                >
                  <RouterLink :to="crumb.to">{{ crumb.label }}</RouterLink>
                </BreadcrumbLink>
                <BreadcrumbPage v-else>{{ crumb.label }}</BreadcrumbPage>
              </BreadcrumbItem>
              <BreadcrumbSeparator v-if="i < breadcrumbs.length - 1" />
            </template>
          </BreadcrumbList>
        </Breadcrumb>

        <div class="ml-auto flex items-center gap-1">
          <slot name="actions" />
          <ThemeToggle />
        </div>
      </header>

      <slot />
    </SidebarInset>
  </SidebarProvider>
</template>
