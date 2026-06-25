<script setup lang="ts">
import {
    Empty,
    EmptyContent,
    EmptyDescription,
    EmptyHeader,
    EmptyMedia,
    EmptyTitle,
} from "@/components/ui/empty";
import { Mail } from "@lucide/vue";

defineProps<{ title?: string; description?: string }>();
</script>

<template>
    <div class="flex items-center justify-center h-full">
        <Empty class="border-none max-w-xs">
            <EmptyMedia variant="icon">
                <Mail />
            </EmptyMedia>
            <EmptyHeader>
                <EmptyTitle>{{ title ?? "No message selected" }}</EmptyTitle>
                <EmptyDescription>{{
                    description ??
                    "Choose a message from the list to read it here."
                }}</EmptyDescription>
            </EmptyHeader>
            <EmptyContent />
        </Empty>
    </div>
</template>
