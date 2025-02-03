<script lang="ts">
    // svelte
    import { Avatar } from "flowbite-svelte";
    import { fade } from "svelte/transition";
    // api
    import { getChatApp } from "$lib/chat.client";
    import type { ApiChatMessage, ApiChatUser } from "../../chat-api/api";

    // props
    let { msg, ava_url}: {
        ava_url: string,
        msg: ApiChatMessage,
    } = $props();

    let app = getChatApp()
    const isMy = msg.from == app.getAuth().currentUser!.uid;
</script>

<div class="flex flex-row mt-1" in:fade>
    <div class="flex flex-col">
         <Avatar class="w-6 h-6" src={ava_url}/>
    </div>
    <div class="flex flex-col flex-1">
        <div class:bg-blue-50={isMy} class:bg-gray-50={!isMy} class="ml-2 mt-2 w-fit max-w-64 text-wrap break-words p-4 rounded-xl rounded-tl-none shadow-sm border-gray-200 border-2 " >
            <span class="text-black text-wrap">{msg.msg}</span>
        </div>
    </div>
</div>