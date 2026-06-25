<script setup lang="ts">
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import {
    Empty,
    EmptyContent,
    EmptyDescription,
    EmptyHeader,
    EmptyMedia,
    EmptyTitle,
} from "@/components/ui/empty";
import { Inbox } from "@lucide/vue";
import MessageRow from "./MessageRow.vue";
import type { Message } from "@/types";

defineProps<{
    messages: Message[];
    loading: boolean;
    selectedId: string | null;
    emptyTitle?: string;
    emptyDescription?: string;
}>();

const emit = defineEmits<{
    select: [id: string];
    star: [id: string, starred: boolean];
}>();
</script>

<template>
    <div class="flex flex-col h-full">
        <Empty
            v-if="messages.length === 0"
            class="border-none flex-1"
        >
            <EmptyMedia variant="icon">
                <Inbox />
            </EmptyMedia>
            <EmptyHeader>
                <EmptyTitle>{{ emptyTitle ?? "No messages" }}</EmptyTitle>
                <EmptyDescription>{{
                    emptyDescription ??
                    "Messages will appear here once received."
                }}</EmptyDescription>
            </EmptyHeader>
            <EmptyContent />
        </Empty>

        <ScrollArea v-else class="flex-1">
            <MessageRow
                v-for="msg in messages"
                :key="msg.id"
                :message="msg"
                :is-selected="msg.id === selectedId"
                @select="emit('select', $event)"
                @star="(id, starred) => emit('star', id, starred)"
            />
        </ScrollArea>
    </div>
</template>
