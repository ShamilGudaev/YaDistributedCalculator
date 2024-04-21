<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod';
import { useForm } from 'vee-validate';
import { FormField, FormItem, FormControl, FormLabel, FormMessage } from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import * as z from 'zod'
import { useRouter, RouterLink } from 'vue-router';

const formSchema = toTypedSchema(z.object({ login: z.string(), password: z.string() }))
const form = useForm({ validationSchema: formSchema })

const responseSchema = z.object({ result: z.union([z.literal("authorized"), z.literal("invalid_credentials")]) })

const router = useRouter();

const onSubmit = form.handleSubmit(async (values) => {

    const res = await fetch('http://127.0.0.1:1323/authorization', {
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

    const { result } = responseSchema.parse(await res.json())

    if (result === "authorized") {
        router.push("/")
        return
    }

    form.setErrors({ login: "Invalid credentials" })
})

</script>
<template>
    <div class="flex flex-col items-center p-4">
        <div class="flex flex-col w-full max-w-[960px]">
            <form @submit="onSubmit" class=" space-y-6">
                <FormField v-slot="{ componentField }" name="login">
                    <FormItem>
                        <FormLabel>Login</FormLabel>
                        <FormControl>
                            <Input type="text" v-bind="componentField" />
                        </FormControl>
                        <FormMessage></FormMessage>
                    </FormItem>
                </FormField>
                <FormField v-slot="{ componentField }" name="password">
                    <FormItem>
                        <FormLabel>Password</FormLabel>
                        <FormControl>
                            <Input type="password" v-bind="componentField" />
                        </FormControl>
                        <FormMessage></FormMessage>
                    </FormItem>
                </FormField>
                <div class="flex">
                    <Button type="submit">Authorize</Button>
                </div>
            </form>
            <RouterLink to="/registration" class="mt-3">Registration</RouterLink>
        </div>
    </div>
</template>