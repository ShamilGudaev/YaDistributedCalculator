<script setup lang="ts">
import { useForm } from 'vee-validate'
import { TabsContent } from '../ui/tabs'
import { executionTimeScheme, useDataStore } from '@/stores/data'
import { z } from 'zod'
import { Input } from '../ui/input'
import { FormControl, FormField, FormItem, FormLabel } from '../ui/form'
import { Button } from '../ui/button'
import { toTypedSchema } from '@vee-validate/zod'
import { ref } from 'vue'
import { Separator } from '../ui/separator'
import { whenever } from '@vueuse/core'

const data = useDataStore()

const formSchema = toTypedSchema(executionTimeScheme)
const form = useForm({ validationSchema: formSchema })

const responseSchema = z.object({
  ok: z.boolean().refine((v) => v)
})

const editMode = ref(false)

whenever(
  () => editMode.value,
  () => {
    console.log('editmode')

    form.setValues({
      opMulMS: data.executionTime.opMulMS,
      opDivMS: data.executionTime.opDivMS,
      opAddMS: data.executionTime.opAddMS,
      opSubMS: data.executionTime.opSubMS
    })
  },
  {
    immediate: false
  }
)

const onSubmit = form.handleSubmit(async (values) => {
  const res = await fetch('http://localhost:1323/apply_execution_time', {
    method: 'POST',
    cache: 'no-cache',

    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values)
  })

  if (!res.ok) {
    console.error(res.statusText)
    return
  }

  responseSchema.parse(await res.json())
  editMode.value = false
})

function onEditClick() {
  console.log('clicked')
  editMode.value = true
}

function onCancelClick() {
  console.log('canc')
  editMode.value = false
}
</script>
<template>
  <TabsContent value="settings">
    <form v-if="editMode" @submit="onSubmit" class="space-y-6">
      <FormField v-slot="{ componentField }" name="opMulMS">
        <FormItem>
          <FormLabel>Multiplication</FormLabel>
          <FormControl>
            <Input
              type="number"
              placeholder="5000"
              v-bind="componentField"
              :default-value="data.executionTime.opMulMS"
            />
          </FormControl>
          <FormMessage></FormMessage>
        </FormItem>
      </FormField>
      <FormField v-slot="{ componentField }" name="opDivMS">
        <FormItem>
          <FormLabel>Division</FormLabel>
          <FormControl>
            <Input
              type="number"
              placeholder="5000"
              v-bind="componentField"
              :default-value="data.executionTime.opDivMS"
            />
          </FormControl>
          <FormMessage></FormMessage>
        </FormItem>
      </FormField>
      <FormField v-slot="{ componentField }" name="opAddMS">
        <FormItem>
          <FormLabel>Addition</FormLabel>
          <FormControl>
            <Input
              type="number"
              placeholder="5000"
              v-bind="componentField"
              :default-value="data.executionTime.opAddMS"
            />
          </FormControl>
          <FormMessage></FormMessage>
        </FormItem>
      </FormField>
      <FormField v-slot="{ componentField }" name="opSubMS">
        <FormItem>
          <FormLabel>Subtraction</FormLabel>
          <FormControl>
            <Input
              type="number"
              placeholder="5000"
              v-bind="componentField"
              :default-value="data.executionTime.opSubMS"
            />
          </FormControl>
          <FormMessage></FormMessage>
        </FormItem>
      </FormField>
      <div class="flex">
        <Button type="submit">Submit</Button>
        <Button type="button" variant="secondary" class="ml-4" @click="onCancelClick"
          >Cancel</Button
        >
      </div>
    </form>
    <div v-else class="flex flex-col">
      <div>Multiplication: {{ data.executionTime.opMulMS }}ms</div>
      <Separator class="my-4" />
      <div>Division: {{ data.executionTime.opDivMS }}ms</div>
      <Separator class="my-4" />
      <div>Addition: {{ data.executionTime.opAddMS }}ms</div>
      <Separator class="my-4" />
      <div>Subtraction: {{ data.executionTime.opSubMS }}ms</div>
      <div class="flex">
        <Button type="button" @click="onEditClick" class="mt-6">Edit</Button>
      </div>
    </div>
  </TabsContent>
</template>
