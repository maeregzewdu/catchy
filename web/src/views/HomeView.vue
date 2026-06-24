<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    CardDescription,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import {
    ArrowRight,
    Zap,
    Mail,
    Server,
    CheckCircle2,
    XCircle,
} from "@lucide/vue";
import { useAppStore } from "@/stores/app";

const appStore = useAppStore();

const isOnline = computed(() => appStore.health?.status === "ok");
const trapAddr = computed(() => appStore.health?.trap_addr ?? "localhost:1025");
const version = computed(() => appStore.health?.version ?? "");

const laravelCode = `MAIL_MAILER=smtp
MAIL_HOST=127.0.0.1
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null`;

const nodeCode = `const nodemailer = require('nodemailer')
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
})`;

const swaksCode = `swaks --to test@example.com \\
      --server 127.0.0.1:1025 \\
      --header "Subject: Test from swaks"`;
</script>

<template>
    <div class="min-h-screen bg-background">
        <!-- Nav -->
        <header
            class="border-b border-border px-6 py-3 flex items-center justify-between"
        >
            <div class="flex items-center gap-2">
                <div
                    class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
                >
                    <Zap class="h-4 w-4" />
                </div>
                <span class="font-semibold">catchy</span>
                <Badge v-if="version" variant="outline" class="text-xs"
                    >v{{ version }}</Badge
                >
            </div>
            <Button as-child size="sm">
                <RouterLink to="/dashboard">
                    Open Dashboard
                    <ArrowRight class="h-4 w-4" />
                </RouterLink>
            </Button>
        </header>

        <!-- Hero -->
        <section class="max-w-4xl mx-auto px-6 py-16 text-center">
            <Badge class="mb-4" variant="secondary">Local dev tool</Badge>
            <h1 class="text-4xl font-bold tracking-tight mb-4">
                Catch every email your app sends
            </h1>
            <p class="text-muted-foreground text-lg mb-8 max-w-2xl mx-auto">
                catchy is a local SMTP trap and IMAP client for developers.
                Point your app at
                <code
                    class="font-mono bg-muted px-1.5 py-0.5 rounded text-sm"
                    >{{ trapAddr }}</code
                >
                and every outgoing email lands here — no inbox clutter, no real
                sends.
            </p>
            <div class="flex items-center justify-center gap-3">
                <Button as-child size="lg">
                    <RouterLink to="/dashboard">
                        Go to Dashboard
                        <ArrowRight class="h-4 w-4" />
                    </RouterLink>
                </Button>
                <Button as-child variant="outline" size="lg">
                    <RouterLink to="/dashboard/settings"
                        >Configure Accounts</RouterLink
                    >
                </Button>
            </div>
        </section>

        <Separator />

        <!-- Status + quick info -->
        <section
            class="max-w-4xl mx-auto px-6 py-12 grid grid-cols-1 md:grid-cols-3 gap-4"
        >
            <Card>
                <CardHeader class="pb-2">
                    <CardTitle class="text-sm flex items-center gap-2">
                        <div
                            :class="isOnline ? 'bg-green-500' : 'bg-red-500'"
                            class="h-2 w-2 rounded-full"
                        />
                        Server Status
                    </CardTitle>
                </CardHeader>
                <CardContent class="text-sm text-muted-foreground">
                    <div class="flex items-center gap-1.5">
                        <CheckCircle2
                            v-if="isOnline"
                            class="h-4 w-4 text-green-500"
                        />
                        <XCircle v-else class="h-4 w-4 text-destructive" />
                        {{ isOnline ? "Running" : "Unreachable" }}
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader class="pb-2">
                    <CardTitle class="text-sm flex items-center gap-2">
                        <Server class="h-4 w-4" />
                        SMTP Trap
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <code
                        class="text-sm font-mono bg-muted px-2 py-1 rounded"
                        >{{ trapAddr }}</code
                    >
                    <p class="text-xs text-muted-foreground mt-1">
                        No auth required
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader class="pb-2">
                    <CardTitle class="text-sm flex items-center gap-2">
                        <Mail class="h-4 w-4" />
                        Web
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <code class="text-sm font-mono bg-muted px-2 py-1 rounded"
                        >localhost:8080</code
                    >
                    <p class="text-xs text-muted-foreground mt-1">
                        This interface
                    </p>
                </CardContent>
            </Card>
        </section>

        <!-- Code examples -->
        <section class="max-w-4xl mx-auto px-6 pb-16">
            <h2 class="text-xl font-semibold mb-4">Connect your app</h2>
            <Tabs default-value="laravel">
                <TabsList>
                    <TabsTrigger value="laravel">Laravel</TabsTrigger>
                    <TabsTrigger value="node">Node.js</TabsTrigger>
                    <TabsTrigger value="swaks">swaks</TabsTrigger>
                </TabsList>
                <TabsContent value="laravel" class="mt-0">
                    <Card class="rounded-t-none border-t-0">
                        <CardContent class="pt-4">
                            <CardDescription class="mb-3"
                                >Add to your
                                <code class="font-mono text-xs">.env</code>
                                file:</CardDescription
                            >
                            <pre
                                class="bg-muted rounded-md p-4 text-xs font-mono overflow-x-auto"
                                >{{ laravelCode }}</pre
                            >
                        </CardContent>
                    </Card>
                </TabsContent>
                <TabsContent value="node" class="mt-0">
                    <Card class="rounded-t-none border-t-0">
                        <CardContent class="pt-4">
                            <CardDescription class="mb-3"
                                >Using nodemailer:</CardDescription
                            >
                            <pre
                                class="bg-muted rounded-md p-4 text-xs font-mono overflow-x-auto"
                                >{{ nodeCode }}</pre
                            >
                        </CardContent>
                    </Card>
                </TabsContent>
                <TabsContent value="swaks" class="mt-0">
                    <Card class="rounded-t-none border-t-0">
                        <CardContent class="pt-4">
                            <CardDescription class="mb-3"
                                >Send a test email from the
                                terminal:</CardDescription
                            >
                            <pre
                                class="bg-muted rounded-md p-4 text-xs font-mono overflow-x-auto"
                                >{{ swaksCode }}</pre
                            >
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </section>
    </div>
</template>
