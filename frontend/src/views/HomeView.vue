<script setup lang="ts">
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'
import CalculatorTab from '@/components/calculator/CalculatorTab.vue'
import SettingsTab from '@/components/settings/SettingsTab.vue'
import AgentsTab from '@/components/agents/AgentsTab.vue'
</script>

<template>
  <main class="w-full flex flex-col items-center px-4 py-6">
    <Tabs default-value="calculator" class="w-full max-w-[960px]">
      <ScrollArea>
        <TabsList class="bg-transparent p-0 h-8">
          <TabsTrigger value="calculator" class="data-[state=active]:bg-muted">
            {{ 'Calculator' }}
          </TabsTrigger>
          <TabsTrigger value="settings" class="data-[state=active]:bg-muted">Settings</TabsTrigger>
          <TabsTrigger value="agents" class="data-[state=active]:bg-muted">
            {{ 'Agents' }}
          </TabsTrigger>
        </TabsList>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
      <div class="pt-4 w-full"></div>
      <CalculatorTab />
      <SettingsTab />
      <AgentsTab />
    </Tabs>
  </main>
</template>
