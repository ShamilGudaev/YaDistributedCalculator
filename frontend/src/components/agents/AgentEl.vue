<script setup lang="ts">
import { useDataStore } from '@/stores/data'
import { computed } from 'vue'
import XOctagonIcon from '../icons/XOctagonIcon.vue'
import ClockIcon from '../icons/ClockIcon.vue'
import CheckIcon from '../icons/CheckIcon.vue'

const props = defineProps<{ agentId: string }>()
const data = useDataStore()
const agent = computed(() => {
  const agent = data.agentMap.get(props.agentId)
  if (agent === undefined) throw new Error('missing agent')
  return agent
})

const deleted = computed(() => agent.value.deletedAt !== null)
const missing = computed(() => !deleted.value && data.time - agent.value.lastSeen >= 15 * 1000)
const ok = computed(() => !deleted.value && !missing.value)
</script>
<template>
  <div class="flex">
    <div>
      <CheckIcon class="text-green-300" v-if="ok" />
      <ClockIcon class="text-orange-200" v-else-if="missing" />
      <XOctagonIcon class="text-red-200" v-else />
    </div>
    <div class="pl-2">
      {{ agent.id }} -
      {{ agent.expressionIds.length === 0 ? 'Idle' : `${agent.expressionIds.length} tasks` }}
    </div>
  </div>
</template>
