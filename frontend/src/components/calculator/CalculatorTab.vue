<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { TabsContent } from '@/components/ui/tabs'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { useDataStore } from '@/stores/data'
import { computed } from 'vue'
import { compareUInt64Str } from '@/utils/comparators'
import { Separator } from '@/components/ui/separator'
import CheckIcon from '../icons/CheckIcon.vue'
import ClockIcon from '../icons/ClockIcon.vue'
import RefreshIcon from '../icons/RefreshIcon.vue'

const formSchema = toTypedSchema(
  z.object({
    expression: z.string()
  })
)

const responseSchema = z.union([
  z.object({
    expressionId: z.string(),
    error: z.null()
  }),
  z.object({
    expressionId: z.null(),
    error: z.string()
  })
])

const form = useForm({ validationSchema: formSchema })

const onSubmit = form.handleSubmit(async (values) => {
  const res = await fetch('http://127.0.0.1:1323/add_expression', {
    method: 'POST',
    cache: 'no-cache',
    credentials: 'include',
    mode: 'cors',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values)
  })

  if (!res.ok) {
    console.error(res.statusText)
    return
  }

  const data = responseSchema.parse(await res.json())
  console.log(data)
  if (data.error != null) {
    form.setFieldError('expression', data.error)
    return
  }

  form.resetForm()
})

const data = useDataStore()

const expressions = computed(() => {
  return [...data.expressionMap.entries()]
    .sort(([key0], [key1]) => -compareUInt64Str(key0, key1))
    .map(([, v]) => v)
})
</script>
<template>
  <TabsContent value="calculator">
    <form @submit="onSubmit" class="space-y-6">
      <FormField v-slot="{ componentField }" name="expression">
        <FormItem>
          <FormLabel>Expression</FormLabel>
          <FormControl>
            <Input type="text" placeholder="" v-bind="componentField" />
          </FormControl>
          <FormMessage></FormMessage>
        </FormItem>
      </FormField>
      <Button type="submit">Submit</Button>
    </form>
    <div class="pt-6 w-full"></div>
    <div class="flex flex-col w-full">
      <template v-for="(expression, i) of expressions" :key="expression.id">
        <div class="flex">
          <div>
            <CheckIcon class="text-green-300" v-if="expression.result !== null" />
            <RefreshIcon class="text-yellow-200" v-else-if="expression.agentId !== null" />
            <ClockIcon class="text-orange-200" v-else />
          </div>
          <div class="pl-2">
            {{ expression.text + (expression.result === null ? '' : ` = ${expression.result}`) }}
          </div>
        </div>
        <Separator class="my-2" v-if="i < expressions.length - 1"></Separator>
      </template>
    </div>
  </TabsContent>
</template>
