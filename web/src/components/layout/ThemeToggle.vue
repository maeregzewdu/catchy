<script setup lang="ts">
import { useDark, useToggle } from "@vueuse/core";
import { Button } from "@/components/ui/button";
import { Sun, Moon } from "@lucide/vue";

const isDark = useDark();
const toggleDark = useToggle(isDark);
</script>

<template>
  <Button
    variant="ghost"
    size="icon-sm"
    :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    @click="toggleDark()"
  >
    <Sun v-if="isDark" class="h-4 w-4" />
    <Moon v-else class="h-4 w-4" />
  </Button>
</template>
