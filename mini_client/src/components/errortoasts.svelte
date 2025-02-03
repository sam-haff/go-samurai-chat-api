<script lang="ts">
    import { Toast } from "flowbite-svelte";
    import { fade } from "svelte/transition";

class ErrorMessageWithId{
        msg: string = "";
        id: string = "";

        constructor (msg: string, id: string) {
            this.msg = msg;
            this.id = id;
        }
    }

    // props
    let { messages = $bindable() }: {
        messages: string[]
    } = $props();

    // state
    let indexedMessages = $derived(messages.map( (val:string)=>new ErrorMessageWithId(val, crypto.randomUUID()) ))
    
</script>
    <div class="absolute top-0 w-full flex flex-col !z-50">

{#each indexedMessages as msg,i (msg.id)}
        <Toast transition={fade} color="dark" class="relative rounded bg-red-100 !mx-auto mt-2 !max-w-[100%] !p-2 !gap-0 !w-5/6 sm:!w-4/6 !z-50" on:close={function () { console.log("toast closed"); console.log(messages); console.log(messages.splice(i, 1)); console.log(messages);  }} align={false}>
            <span class="text-red-700 my-auto">{msg.msg}</span>
        </Toast>    
{/each}
    </div>