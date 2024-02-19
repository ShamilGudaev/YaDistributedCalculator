import { ref, readonly, watch } from 'vue'
import { defineStore } from 'pinia'
import * as z from 'zod'
import { useEventSource, useIntervalFn } from '@vueuse/core'

const nanoidScheme = z.string().regex(/^[A-Za-z0-9_-]{21}$/)
const uint32Scheme = z.number().int().nonnegative().lte(0xffff_ffff)
const uint64StrScheme = z
  .string()
  .regex(/^([1-9]\d{0,19}|0)$/)
  .refine((v) => v.length < 20 || v <= '18446744073709551615')

const expressionScheme = z.intersection(
  z.object({
    id: uint64StrScheme,
    text: z.string()
  }),
  z.union([
    z.object({
      result: z.null(),
      agentId: z.nullable(nanoidScheme)
    }),
    z.object({
      result: z.string(),
      agentId: z.null()
    })
  ])
)

const expressionArrScheme = z.array(expressionScheme)

export type Expression = z.infer<typeof expressionScheme>

const stringDateScheme = z
  .string()
  .datetime({ offset: false })
  .pipe(z.coerce.date())
  .transform((date) => date.valueOf())

const agentScheme = z.object({
  id: nanoidScheme,
  expressionIds: z.array(uint64StrScheme),
  lastSeen: stringDateScheme,
  deletedAt: z.nullable(stringDateScheme)
})

const agentArrScheme = z.array(agentScheme)

const agentIdsScheme = z.array(nanoidScheme)

export type Agent = z.infer<typeof agentScheme>

export const executionTimeScheme = z.object({
  opMulMS: uint32Scheme,
  opDivMS: uint32Scheme,
  opAddMS: uint32Scheme,
  opSubMS: uint32Scheme
})

export type ExecutionTime = z.infer<typeof executionTimeScheme>

const intitialDataScheme = z.object({
  expressions: expressionArrScheme,
  agents: agentArrScheme,
  executionTime: executionTimeScheme
})

type InitialData = z.infer<typeof intitialDataScheme>

export const useDataStore = defineStore('data', () => {
  const expressionMap = ref(new Map<string, Expression>())
  const agentMap = ref(new Map<string, Agent>())
  const executionTime = ref<ExecutionTime>({ opAddMS: 0, opDivMS: 0, opMulMS: 0, opSubMS: 0 })
  const time = ref(0)

  useIntervalFn(() => {
    time.value = new Date().valueOf()
  })

  const { event, data } = useEventSource('http://localhost:1323/subscribe', [
    'initial_data',
    'expressions_change',
    'agents_change',
    'exec_time_change',
    'agents_remove'
  ])

  watch(
    [event, data],
    ([event, data0]) => {
      if (event == null || data0 == null) return
      const data = JSON.parse(data0)
      console.log(data)
      switch (event) {
        case 'initial_data':
          loadInitialData(intitialDataScheme.parse(data))
          return
        case 'expressions_change':
          expressionArrScheme.parse(data).map(expressionChange)
          return
        case 'agents_change':
          agentArrScheme.parse(data).map(agentChange)
          return
        case 'exec_time_change':
          execTimeChange(executionTimeScheme.parse(data))
          return
        case 'agents_remove':
          agentIdsScheme.parse(data).map(agentRemove)
      }
    },
    { immediate: true }
  )

  function loadInitialData(data: InitialData) {
    for (const expression of data.expressions) {
      expressionMap.value.set(expression.id, expression)
    }

    for (const agent of data.agents) {
      agentMap.value.set(agent.id, agent)
    }

    Object.assign(executionTime.value, data.executionTime)
  }

  function expressionChange(expr: Expression) {
    const prev = expressionMap.value.get(expr.id)

    if (prev === undefined) {
      expressionMap.value.set(expr.id, expr)
      return
    }

    Object.assign(prev, expr)
  }

  function agentChange(agent: Agent) {
    const prev = agentMap.value.get(agent.id)

    if (prev === undefined) {
      agentMap.value.set(agent.id, agent)
      return
    }

    Object.assign(prev, agent)
  }

  function agentRemove(id: string) {
    agentMap.value.delete(id)
  }

  function execTimeChange(execTime: ExecutionTime) {
    Object.assign(executionTime.value, execTime)
  }

  return {
    expressionMap: readonly(expressionMap),
    agentMap: readonly(agentMap),
    executionTime: readonly(executionTime),
    time: readonly(time)
  }
})
