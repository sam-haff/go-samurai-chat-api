<script lang="ts">
    // svelte
    import { Spinner } from "flowbite-svelte";
    import { onMount } from "svelte";
    import { createMediaStore } from "svelte-media-queries";
    // custom components
    import ChatPreview from "./ChatPreview.svelte";
    // api
    import { getChatApp } from "$lib/chat.client";

    let app = getChatApp();

    // props
    let {closeMenu, selectedChatID = $bindable()}: {
        closeMenu: any,
        selectedChatID: string
    } = $props()

    // state
    let chatsList = $state([] as string[]);
    let chatsHiddenTop = $state(false); // 0 no hidden els, 1 hidden els on top, 2 hidden els on bot, 3 hidden els on top and bot
    let chatsHiddenBot = $state(false); // 0 no hidden els, 1 hidden els on top, 2 hidden els on bot, 3 hidden els on top and bot

    let query='(max-width: 480px)'
    let isMobile = createMediaStore(query);

    // effects
    $effect(
        function () {
            let newChatsList = app.chatsList();
            console.log("New chats list!");
            console.log(newChatsList);
            if (chatsList.length !== newChatsList.length) { // TODO: foot shooting
                chatsList = newChatsList;
            }

            onscroll();
            setTimeout(onscroll, 1000);
        }
    )
    
    onMount( function () {
            onscroll();
    } )
    function onscroll() {
        if (!isMobile) { return; }
        let el = document.getElementById("chats-list");

        if (!el) return;

        let scrollable = el.scrollHeight > el.clientHeight;
        let atBot = el.scrollHeight - el.scrollTop - el.clientHeight < 1
        let atTop = el.scrollTop <= 0;
        if (!scrollable) {
            return;
        }

        if (!atTop && !atBot) {
            console.log("both sides hidden");
            el.classList.remove("list-hidden-top");
            el.classList.remove("list-hidden-bot");
            el.classList.add("list-hidden-both");
            return;
        }
        if (el.scrollTop <= 0) {
            el.classList.remove("list-hidden-both");
            el.classList.add("list-hidden-bot");
        } else {
            el.classList.remove("list-hidden-bot");
        }
        if (atBot) {
            el.classList.remove("list-hidden-both");
            el.classList.add("list-hidden-top");
        } else {
            el.classList.remove("list-hidden-top");
        }
    }

</script>
<!-- no relative -->
<div class="flex flex-col relative h-full w-full mt-6 p-0 overflow-hidden">
    {#if false}
        <Spinner class="m-auto"/>
    {:else} 
        {#if chatsList.length === 0}
            <span class="ml-2 font-semibold text-gray-700"><i>No chats yet...</i></span>
        {:else}
            <div class="list-smooth-extra w-full"></div>
            <div {onscroll} onresize={onscroll} id="chats-list" class:list-hidden-bot={chatsHiddenBot} class:list-hidden-top={chatsHiddenTop} class:list-hidden-both={chatsHiddenTop} class="chats-list-max-height list-smooth-extra pb-1 overflow-auto h-80 [&::-webkit-scrollbar]:w-2
                        [&::-webkit-scrollbar-track]:rounded-full
                        [&::-webkit-scrollbar-track]:bg-transparent
                        [&::-webkit-scrollbar-thumb]:rounded-full
                      [&::-webkit-scrollbar-thumb]:bg-gray-300">
                {#each chatsList as chatID (chatID)}
                    <ChatPreview {closeMenu} bind:selectedChatID chatID={chatID}/>
                {/each}
            </div>
        {/if}
    {/if}
</div>


<style>
    .list-hidden-top
    {
        mask: linear-gradient(to top, rgba(0, 0, 0, 255) 95%, rgba(0,0,0,0) 100%);
        mask-composite: intersect;
    }
    .list-hidden-bot
    {
        mask-composite: intersect;
        mask: linear-gradient(to bottom, rgba(0, 0, 0, 255) 95%, rgba(0,0,0,0) 100%);
    }
    .list-hidden-both
    {
        mask-composite: intersect;
        mask: linear-gradient(to bottom, rgba(0,0,0,0) 0%, rgba(0,0,0,1.0) 5%, rgba(0, 0, 0, 1.0) 95%, rgba(0,0,0,0) 100%);
    }
    @media (max-height: 500px) {
        .chats-list-max-height{
            max-height: 240px;
        }
    }
</style>