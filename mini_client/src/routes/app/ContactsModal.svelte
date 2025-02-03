<script lang="ts">
    // svelte
    import { fade } from "svelte/transition";
    import { Button, ButtonGroup, Input, Spinner, Toast } from "flowbite-svelte";
    // custom components
    import ContactsList from "./ContactsList.svelte";
    import DivModal from "./DivModal.svelte";
    // custom logic/api
    import { getChatApp } from "$lib/chat.client";
    import { validateUsername } from "$lib/validation.client";

    let app = getChatApp();

    let contactsList = $derived( Array.from(app.contacts.values()) );

    // props
    let { showModal = $bindable(), closeMenu, selectedChatID=$bindable() } = $props();

    // state
    let contactUsername = $state("");
    let loading = $state(false);
    let errorMsg = $state<Array<string>>([]);
    let oldChatID = selectedChatID;
    let preventModalClose = $state(false);

    // effects
    $effect(()=>{ if (oldChatID !== selectedChatID) { showModal = false;}})
    $effect(()=>{ if (showModal) {
        if (contactsList.length === 0) { // TODO: ???
            app.updateContacts();
        }
    }})

    // functions
    function closeModal() {
        showModal = false;
    }

    async function onsubmit(e: any) {
        e.preventDefault();

        if (!validateUsername(contactUsername)){
            errorMsg.push("Wrong username format");
            return;
        }
        loading = true;

        let app = getChatApp();
        try {
            await app.addContact(contactUsername);
        } catch (e: any) {
            errorMsg.push(e.message);
            loading = false;
        }

        loading = false;
        contactUsername = "";
    }
    async function oncontactremove(username:string) {
        try {
            await getChatApp().removeContact(username);
        } catch (e: any) {
            errorMsg.push(e.message);
        }
    }
    function onContactClick() {
        closeModal();
        closeMenu();
    }
    function onInputFocus() {
        // restore?
    }
</script>

<DivModal {preventModalClose} bind:showModal>
    <div class="absolute top-0 w-full !z-50">
    {#each errorMsg as msg}
        <Toast transition={fade} color="dark" class="rounded bg-red-100 !mx-auto mt-2 !max-w-[100%] !p-2 !gap-0 !w-5/6" on:close={function () { console.log("toast closed"); console.log(errorMsg); console.log(errorMsg.indexOf(msg)); errorMsg.splice(errorMsg.indexOf(msg), 1);  }} align={false}>
            <span class="text-red-700 my-auto">{msg}</span>
        </Toast>    
    {/each}
    </div>
    {#snippet header()}
       <span class="text-lg">Contacts</span>
    {/snippet}
    <div class="flex flex-col h-full p-8 overflow-hidden">
        <form {onsubmit}>
            <div class="flex flex-row justify-center items-center">
                <ButtonGroup class="w-4/5">
                    <Input class="!focus:ring-0 !focus:border-transparent focus:border-gray-600 !focus:outline-none focus:ring-0 focus:ring-offset-0" bind:value={contactUsername} size="md" onfocus={onInputFocus} autocapitalize="off" placeholder="Contact's username..." />
                    <Button type="submit" class="!z-0">
                        {#if loading}
                            <Spinner size='5'/>
                        {:else}
                            <svg class="!fill-primary-800 !stroke-primary-700" width="15px" height="15px" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                <path d="M4 12H20M12 4V20"  stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                            </svg>
                        {/if}
                    </Button>
                </ButtonGroup>
            </div>
        </form>
        <div class="flex flex-row flex-1 overflow-y-auto overflow-x-hidden">
            <ContactsList oncontactclick={onContactClick} bind:preventModalClose {oncontactremove} bind:selectedChatID />
        </div>
    </div>
</DivModal>