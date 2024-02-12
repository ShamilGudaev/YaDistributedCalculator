<script setup lang="ts">
import { useDataStore } from '@/stores/data'
import { TabsContent } from '../ui/tabs'
import { computed } from 'vue'
import AgentEl from './AgentEl.vue'
import { Separator } from '../ui/separator';

const data = useDataStore()

const agents = computed(() => {
  return [...data.agentMap.entries()]
    .sort(([id0], [id1]) => {
      if (id0 < id1) {
        return -1
      }

      if (id0 > id1) {
        return 1
      }

      return 0
    })
    .map(([, el]) => el)
})
</script>
<template>
  <TabsContent value="agents">
    <div class="flex flex-col w-full">
      <template v-for="(agent, i) of agents" :key="agent.id">
        <AgentEl :agentId="agent.id" />
        <Separator class="my-4" v-if="i < agents.length - 1"/>
      </template>
    </div>
  </TabsContent>
</template>
