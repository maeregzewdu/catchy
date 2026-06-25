<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";
import { useClipboard } from "@vueuse/core";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import ThemeToggle from "@/components/layout/ThemeToggle.vue";
import { ArrowRight, Zap, Mail, Server, Inbox, Check, Copy } from "@lucide/vue";
import { useAppStore } from "@/stores/app";

const appStore = useAppStore();

const isOnline = computed(() => appStore.health?.status === "ok");
const trapAddr = computed(() => appStore.health?.trap_addr ?? "localhost:1025");
const version = computed(() => appStore.health?.version ?? "");

const { copy } = useClipboard();
const copiedKey = ref<string | null>(null);
function copyCode(key: string, text: string) {
    copy(text);
    copiedKey.value = key;
    setTimeout(() => {
        if (copiedKey.value === key) copiedKey.value = null;
    }, 1500);
}

const features = [
    {
        icon: Server,
        title: "SMTP trap",
        body: "Point any mailer at the local SMTP server. Every message is caught instantly — nothing ever leaves your machine.",
    },
    {
        icon: Inbox,
        title: "IMAP accounts",
        body: "Connect real inboxes to read, search and inspect alongside your trapped mail in one place.",
    },
    {
        icon: Mail,
        title: "Instant web UI",
        body: "Preview rendered HTML, raw source, headers and attachments the moment an email arrives.",
    },
];

const examples = [
    {
        key: "laravel",
        label: "Laravel",
        caption: "Add to your .env file",
        code: `MAIL_MAILER=smtp
MAIL_HOST=127.0.0.1
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null`,
    },
    {
        key: "node",
        label: "Node.js",
        caption: "Using nodemailer",
        code: `const nodemailer = require('nodemailer')
const transport = nodemailer.createTransport({
  host: '127.0.0.1',
  port: 1025,
  auth: false,
})
await transport.sendMail({
  from: 'app@example.com',
  to: 'test@example.com',
  subject: 'Hello from catchy!',
  text: 'Caught!',
})`,
    },
    {
        key: "swaks",
        label: "swaks",
        caption: "Send a test email from the terminal",
        code: `swaks --to test@example.com \\
      --server 127.0.0.1:1025 \\
      --header "Subject: Test from swaks"`,
    },
];
</script>

<template>
    <div class="min-h-screen bg-background">
        <!-- Nav -->
        <div class="sticky top-6 z-40 mt-6">
            <header class="max-w-5xl mx-auto px-6">
                <div
                    class="px-4 h-14 flex items-center justify-between rounded-full border border-border/60 bg-background/60 backdrop-blur-xl shadow-lg shadow-black/5 supports-[backdrop-filter]:bg-background/50"
                >
                    <div class="flex items-center gap-2.5">
                        <div class="flex items-center gap-2 px-1.5 cursor-default">
                            <Zap class="h-4 w-4" />
                            <span class="font-bold tracking-tight">Catchy</span>
                        </div>
                        <Badge
                            v-if="version"
                            variant="outline"
                            class="text-xs font-normal"
                            >v{{ version }}</Badge
                        >
                    </div>
                    <div class="flex items-center gap-1.5">
                        <ThemeToggle />
                        <Button as-child size="sm">
                            <RouterLink to="/dashboard">
                                Open Dashboard
                                <ArrowRight class="h-4 w-4" />
                            </RouterLink>
                        </Button>
                    </div>
                </div>
            </header>
        </div>

        <!-- Hero -->
        <section class="relative overflow-hidden">
            <!-- Background glow -->
            <div
                aria-hidden="true"
                class="pointer-events-none absolute inset-x-0 -top-40 flex justify-center"
            >
                <div
                    class="h-112 w-112 rounded-full bg-primary/10 blur-3xl"
                />
            </div>

            <div
                class="relative max-w-3xl mx-auto px-6 pt-24 pb-20 text-center"
            >
                <div
                    class="inline-flex items-center gap-2 rounded-full border border-border/60 bg-muted/40 px-3 py-1 text-xs text-muted-foreground"
                >
                    <span class="relative flex h-2 w-2">
                        <span
                            v-if="isOnline"
                            class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500/60"
                        />
                        <span
                            class="relative inline-flex h-2 w-2 rounded-full"
                            :class="isOnline ? 'bg-green-500' : 'bg-red-500'"
                        />
                    </span>
                    {{
                        isOnline
                            ? `Listening on ${trapAddr}`
                            : "Server unreachable"
                    }}
                </div>

                <h1
                    class="mt-6 text-5xl sm:text-6xl font-bold tracking-tight text-balance leading-[1.05]"
                >
                    Catch every email
                    <br class="hidden sm:block" />
                    your app sends
                </h1>

                <p
                    class="mt-6 text-lg text-muted-foreground max-w-xl mx-auto text-balance leading-relaxed"
                >
                    A local SMTP trap and IMAP client for developers. Point your
                    app at the trap and every outgoing email lands here — no
                    inbox clutter, no real sends.
                </p>

                <div
                    class="mt-10 flex flex-col sm:flex-row items-center justify-center gap-3"
                >
                    <Button as-child size="lg" class="w-full sm:w-auto">
                        <RouterLink to="/dashboard">
                            Go to Dashboard
                            <ArrowRight class="h-4 w-4" />
                        </RouterLink>
                    </Button>
                    <Button
                        as-child
                        variant="outline"
                        size="lg"
                        class="w-full sm:w-auto"
                    >
                        <RouterLink to="/dashboard/settings"
                            >Configure Accounts</RouterLink
                        >
                    </Button>
                </div>
            </div>
        </section>

        <!-- Features -->
        <section class="max-w-5xl mx-auto px-6 py-20">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
                <Card
                    v-for="feature in features"
                    :key="feature.title"
                    class="border-border/60 transition-colors hover:border-border hover:bg-muted/30"
                >
                    <CardContent>
                        <div
                            class="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary"
                        >
                            <component :is="feature.icon" class="h-5 w-5" />
                        </div>
                        <h3 class="mt-5 font-semibold tracking-tight">
                            {{ feature.title }}
                        </h3>
                        <p
                            class="mt-2 text-sm text-muted-foreground leading-relaxed"
                        >
                            {{ feature.body }}
                        </p>
                    </CardContent>
                </Card>
            </div>
        </section>

        <!-- Connect your app -->
        <section class="max-w-3xl mx-auto px-6 pb-28">
            <div class="text-center mb-10">
                <h2 class="text-2xl font-bold tracking-tight">
                    Connect your app
                </h2>
                <p class="mt-2 text-muted-foreground">
                    Drop these settings into your project and start catching
                    mail.
                </p>
            </div>

            <Tabs default-value="laravel" class="flex flex-col gap-4">
                <div class="flex justify-center">
                    <TabsList>
                        <TabsTrigger
                            v-for="ex in examples"
                            class="rounded-full px-4"
                            :key="ex.key"
                            :value="ex.key"
                            >{{ ex.label }}</TabsTrigger
                        >
                    </TabsList>
                </div>

                <TabsContent
                    v-for="ex in examples"
                    :key="ex.key"
                    :value="ex.key"
                    class="mt-0"
                >
                    <Card class="border-border/60 overflow-hidden p-0">
                        <CardHeader
                            class="flex items-center justify-between px-6 py-2 bg-sidebar"
                        >
                            <span class="text-xs text-muted-foreground">{{
                                ex.caption
                            }}</span>
                            <Button
                                variant="ghost"
                                size="icon-sm"
                                class="h-7 w-7"
                                @click="copyCode(ex.key, ex.code)"
                            >
                                <Check
                                    v-if="copiedKey === ex.key"
                                    class="h-3.5 w-3.5 text-green-500"
                                />
                                <Copy v-else class="h-3.5 w-3.5" />
                            </Button>
                        </CardHeader>
                        <CardContent class="pb-6">
                            <pre
                                class="text-sm font-mono leading-relaxed overflow-x-auto"
                                >{{ ex.code }}</pre
                            >
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </section>

        <!-- Footer -->
        <footer class="border-t border-border/60">
            <div
                class="max-w-5xl mx-auto px-6 py-8 flex items-center justify-between text-sm text-muted-foreground"
            >
                <div class="flex items-center gap-2.5">
                    <div class="flex items-center gap-1.5">
                        <Zap class="h-4 w-4" />
                        <span class="font-bold tracking-tight">Catchy</span>
                    </div>
                    <Badge
                        v-if="version"
                        variant="outline"
                        class="text-xs font-normal"
                        >v{{ version }}</Badge
                    >
                </div>
                <span class="font-mono text-xs">{{ trapAddr }}</span>
            </div>
        </footer>
    </div>
</template>
