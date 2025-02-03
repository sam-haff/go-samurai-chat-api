<script lang='ts'>
    import { fade } from "svelte/transition";
    import { goto } from "$app/navigation";
    import { Label, Input } from "flowbite-svelte";
    import { EnvelopeSolid, LockSolid } from  'flowbite-svelte-icons'
    import { getChatApp } from "$lib/chat.client";
    import ChatSubmit from "./SubmitButton.svelte";

    // init

    // props
    let {onfade, errorMessages=$bindable(), loading=$bindable()}:{
        onfade: any,
        errorMessages: string[],
        loading: boolean,
    } = $props();

    // state
    let errorMsg = $state(""); 

    const handleSubmit = async (e: any) => {
        e.preventDefault();

        loading = true;        
        errorMsg = "";

		const formData = new FormData(e.target);
        const email = formData.get('email')?.toString();
        const pwd = formData.get('pwd')?.toString();

        if (email === null || pwd === null) return;

        try {
            await getChatApp().getConn().singin(email!, pwd!);
        } catch (e:any) {
            errorMsg = e.message;
            errorMessages.push(errorMsg);
            loading = false;
            return;
        }
        loading = false;

        // TODO: should also check if registered completely
        goto("/app")
	}
</script>


<form onsubmit={handleSubmit}>
    <div class="flex flex-col mt-8"  transition:fade={{duration: 500}} onoutroend={onfade}>
        <Label for="input-group-1" class="block mb-2">Email</Label>
        <Input id="email" name="email" type="email" placeholder="">
            <EnvelopeSolid slot="left" class="w-5 h-5 text-gray-500 dark:text-gray-400" />
        </Input>
        <Label for="input-group-3" class="block mb-2">Password</Label>
        <Input id="pwd" name="pwd" type="password" placeholder="">
            <LockSolid slot="left" class="w-5 h-5 text-gray-500 dark:text-gray-400" />
        </Input>

        <ChatSubmit label="Login" bind:loading={loading}/>
    </div>
</form>
