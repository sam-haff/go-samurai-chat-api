<script lang='ts'>
    // svelte
    import { fade } from "svelte/transition";
    import { goto } from "$app/navigation";
    import { Avatar, Button, Heading, Indicator, Navbar, NavBrand, NavUl, Spinner } from "flowbite-svelte";
    import { createMediaStore } from "svelte-media-queries";
    // custom components
    import ChatMenu from "./ChatMenu.svelte";
    import ContactsModal from "./ContactsModal.svelte";
    import Chat from "./Chat.svelte";
    import ChatPlaceholder from "./ChatPlaceholder.svelte";
    // api
    import { getChatApp } from "$lib/chat.client";

    let app = getChatApp();

    let user = app.trackCurrentUser();

    let query='(min-height: 450px)'
    let screenSupported = createMediaStore(query);

    // props    
    let {data} = $props();

    // state
    let openContactsModal = $state(false);
    let openChatsModal = $state(false);
    let selectedChatID = $state("");
    let openMenu = $state(true);
    let chatMenuBtnCls = $derived(openMenu ? "ml-5 mr-5 !p-1 bg-transparent border-transparent focus:ring-0 hover:bg-slate-20 bg-slate-100" : "ml-5 mr-5 !p-1 bg-transparent border-transparent focus:ring-0 hover:bg-slate-10")

    // functions
    async function logout() {
        await getChatApp().singout()

        goto("/");
    }

    let screenHeight = $state(0)
    let screenWidth = $state(0)
</script>
<svelte:window bind:innerWidth={screenWidth} bind:innerHeight={screenHeight} />

{#if $screenSupported}
{#if !$user}
    <div class="flex h-screen w-screen justify-center items-center align-middle pb-96">
        <Spinner size="12"/>
    </div>
{:else}
    <div class="flex flex-col h-screen bg-gray-50">
        <div class="flex flex-row">
            <div class="flex flex-col w-screen">
                   <Navbar fluid class="border-b !ml-0 !pl-0"> 
                    <Button onclick={()=>{ openMenu = !openMenu; }} outline  size="sm" class={chatMenuBtnCls}>
                        {#if !openMenu}
                            <svg class="h-6" viewBox="0 0 24 24" fill="#FFF" stroke="#FFF" xmlns="http://www.w3.org/2000/svg"><path d="M4 18L20 18" stroke="#475569" stroke-width="1.5" stroke-linecap="round"/><path d="M4 12L20 12" stroke="#475569" stroke-width="1.5" stroke-linecap="round"/><path d="M4 6L20 6" stroke="#475569" stroke-width="1.5" stroke-linecap="round"/></svg>
                        {:else}
                            <svg class="h-6" viewBox="0 0 24 24" fill="#FFF" stroke="#FFF" xmlns="http://www.w3.org/2000/svg"><path d="M4 18L20 18" stroke="#475569" stroke-width="1.5" stroke-linecap="round"/><path d="M4 12L20 12" stroke="#475569" stroke-width="1.5" stroke-linecap="round"/><path d="M4 6L20 6" stroke="#475569" stroke-width="1.5" stroke-linecap="round"/></svg>
                            <!--<div class="bg-opacity-50">
                                <svg class="h-6" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M6.99486 7.00636C6.60433 7.39689 6.60433 8.03005 6.99486 8.42058L10.58 12.0057L6.99486 15.5909C6.60433 15.9814 6.60433 16.6146 6.99486 17.0051C7.38538 17.3956 8.01855 17.3956 8.40907 17.0051L11.9942 13.4199L15.5794 17.0051C15.9699 17.3956 16.6031 17.3956 16.9936 17.0051C17.3841 16.6146 17.3841 15.9814 16.9936 15.5909L13.4084 12.0057L16.9936 8.42059C17.3841 8.03007 17.3841 7.3969 16.9936 7.00638C16.603 6.61585 15.9699 6.61585 15.5794 7.00638L11.9942 10.5915L8.40907 7.00636C8.01855 6.61584 7.38538 6.61584 6.99486 7.00636Z" fill="#475569"></path> </g></svg>
                            </div> -->
                        {/if}
                    </Button> <NavBrand class="left-0 mb-0">
                        <img src="chat-icon.svg" class="me-3 mb-0 h-6" alt="Logo" />
                        <Heading tag="h6" class="self-center kaushan-script-regular">
                            Samurai
                        </Heading>
                    </NavBrand>
                    
                    <NavUl>
                    </NavUl>

                    {#if ($user)}
                        <div transition:fade class="ms-auto flex items-center">
                            <div class="relative">
                                <Indicator class="absolute top-7 left-7 w-2 h-2" color={getChatApp().alive() ? "green" : "red"} />
                                <Avatar class="mr-2 w-9 h-9"  src={$user.img_url}/>
                            </div>
                            <span class="mr-4 text-sm text-stone-900 font-semibold">{$user.username}</span>
                            <Button onclick={logout} outline={true} class="group !p-0 border-slate-400 border-transparent hover:bg-white w-6 h-6" size="md">
                                <svg class="group-hover:fill-red-600 hover:cursor-pointer w-3 h-3" fill="#000000" viewBox="0 0 32 32" version="1.1" xmlns="http://www.w3.org/2000/svg"><path d="M10 28.75h-6.75v-25.5h6.75c0.69 0 1.25-0.56 1.25-1.25s-0.56-1.25-1.25-1.25v0h-8c-0.69 0-1.25 0.56-1.25 1.25v0 28c0 0.69 0.56 1.25 1.25 1.25h8c0.69 0 1.25-0.56 1.25-1.25s-0.56-1.25-1.25-1.25v0zM31.218 16.162c0.010-0.060 0.016-0.13 0.016-0.201 0-0.157-0.029-0.308-0.083-0.446l0.003 0.008-0-0.002c-0.062-0.141-0.143-0.261-0.243-0.364l0 0c-0.012-0.013-0.015-0.029-0.027-0.041l-5-5c-0.226-0.227-0.539-0.367-0.885-0.367-0.691 0-1.251 0.56-1.251 1.251 0 0.345 0.14 0.658 0.366 0.884v0l2.866 2.866h-18.981c-0.69 0-1.25 0.56-1.25 1.25s0.56 1.25 1.25 1.25v0h18.982l-2.867 2.865c-0.226 0.226-0.366 0.539-0.366 0.884 0 0.691 0.56 1.251 1.251 1.251 0.345 0 0.658-0.14 0.884-0.366l5-5.001c0.146-0.154 0.253-0.347 0.302-0.562l0.002-0.008c0.012-0.042 0.022-0.093 0.029-0.146l0.001-0.006z"></path></svg>
                            </Button>
                        </div>
                    {/if}
                </Navbar>
            </div>
        </div>
        <div class="flex flex-row flex-1 relative">
            <ChatMenu bind:open={openMenu} bind:openContactsModal bind:selectedChatID hiddenMenu={true}/>
            <div class="flex flex-col flex-1 pb-20 sm:pb-6">
                {#if selectedChatID !== ""}
                    <Chat bind:selectedChatID/>
                {:else}
                    <ChatPlaceholder/>
                {/if}
            </div>
        </div>
        
    </div>

    <ContactsModal closeMenu={function() { openMenu= false; }} bind:showModal={openContactsModal} bind:selectedChatID={selectedChatID}/>
{/if}
{:else}
    <div class="flex w-screen h-screen justify-center items-center align-middle">
        <span>Screen size is not supported </span>
    </div>
{/if}

<style>
</style>