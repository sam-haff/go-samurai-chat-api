<script lang="ts">
    // svelte or svelte libs
    import {swipeable} from "@svelte-put/swipeable"
    import type { SwipeEndEventDetail, SwipeSingleDirection, SwipeStartEventDetail } from "@svelte-put/swipeable";
    import { fade } from "svelte/transition";
    import { Avatar, Button } from "flowbite-svelte";
    import { MessagesSolid} from "flowbite-svelte-icons";
    import { SvelteSet } from "svelte/reactivity";
    // custom logic or api
    import { getChatApp } from "$lib/chat.client";
    import { createChatID } from "../../chat-api/api";
    import { disableTouchScroll, enableTouchScroll } from "$lib/scrollutils.client";

    let app = getChatApp();
    let contactsList = $derived( Array.from(app.contacts.values()) );

    // props
    let { selectedChatID = $bindable(), preventModalClose = $bindable(), oncontactclick, oncontactremove}: {
        selectedChatID: string,
        preventModalClose: boolean,
        oncontactclick: any,
        oncontactremove: (u: string)=>Promise<void>,
    } = $props();

    // state
    let removingContacts = new SvelteSet<string>();
    let direction: SwipeSingleDirection | null = $state(null);
    $effect(function() {
        preventModalClose = direction !== null;
    })

	function swipestart(e: CustomEvent<SwipeStartEventDetail>) {
        //disableTouchScroll(); // disable scroll while swiping for better UX
		direction = e.detail.direction;
        setTimeout(()=>{direction=null}, 2000);
	}

	async function swipeend(e: CustomEvent<SwipeEndEventDetail>) {
 //       enableTouchScroll();
		const { passThreshold } = e.detail;

		if (passThreshold) {
            let el = e.target as HTMLElement;
			const username = (e.target as HTMLElement).dataset.username;
            const contactIndex: number|undefined = Number(el.dataset.index);
            if (username){
                let contact = contactsList[contactIndex];
                removingContacts.add(contact.uid);
                await oncontactremove(username);
                setTimeout( ()=>removingContacts.delete(contact.uid), 800);
            }
		}
        direction = null;
	}
</script>

{#if contactsList.length === 0}
    <div class="flex flex-1 h-full items-center justify-center">
        <span>No contacts yet...</span>
    </div>
{:else}
    <div class="flex w-11/12 sm:w-3/4 p-2 flex-col max-h-[80vh] mt-2 items-center">
        {#each contactsList as contact, i (contact.uid)}
        <div class="flex mt-1 relative w-full h-fit z-10">
            <div class="absolute fill-red-500 stroke-red-500 right-0 top-5" out:fade>
                <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0,0,256,256" width="20px" height="20px" fill-rule="nonzero"><g fill="#ff3e3e" fill-rule="nonzero" stroke="none" stroke-width="1" stroke-linecap="butt" stroke-linejoin="miter" stroke-miterlimit="10" stroke-dasharray="" stroke-dashoffset="0" font-family="none" font-weight="none" font-size="none" text-anchor="none" style="mix-blend-mode: normal"><g transform="scale(8.53333,8.53333)"><path d="M13,3c-0.26757,-0.00363 -0.52543,0.10012 -0.71593,0.28805c-0.1905,0.18793 -0.29774,0.44436 -0.29774,0.71195h-5.98633c-0.36064,-0.0051 -0.69608,0.18438 -0.87789,0.49587c-0.18181,0.3115 -0.18181,0.69676 0,1.00825c0.18181,0.3115 0.51725,0.50097 0.87789,0.49587h18c0.36064,0.0051 0.69608,-0.18438 0.87789,-0.49587c0.18181,-0.3115 0.18181,-0.69676 0,-1.00825c-0.18181,-0.3115 -0.51725,-0.50097 -0.87789,-0.49587h-5.98633c0,-0.26759 -0.10724,-0.52403 -0.29774,-0.71195c-0.1905,-0.18793 -0.44836,-0.29168 -0.71593,-0.28805zM6,8v16c0,1.105 0.895,2 2,2h14c1.105,0 2,-0.895 2,-2v-16z"></path></g></g></svg>
            </div>
            {#if removingContacts.has(contact.uid)} 
                <div class="absolute left-6 top-4 text-slate-600">
                    <span>Removing...</span>
                </div>
            {/if}
            <div use:swipeable onswipestart={swipestart} onswipeend={swipeend}
				style:left="var(--swipe-distance-x)" data-index={i} data-username={contact.username} class="flex relative w-full" in:fade={{duration: 100}}>
                <Button  onclick={()=>{  if (direction !== null) return; oncontactclick(); selectedChatID = createChatID(getChatApp().getAuth().currentUser!.uid, contact.uid); console.log(selectedChatID);}} 
                    color="none"
                    outline 
                    class="relative z-0 group hover:bg-gray-200 bg-gray-100 p-4 w-full border-2 border-none hover:border-2 border-gray-400 hover:border-gray-900 outline-none focus:ring-0">
                    
                    <Avatar size="sm" src={contact.img_url}/>
                    <span class="ms-2 mr-auto text-gray-950 group-hover:text-gray-950">{contact.username}</span>
                    <MessagesSolid class="hidden ml-auto group-hover:block fill-gray-700 w-6 h-6"/>
                </Button>
            </div>
            <hr/>
        </div>
        {/each}
    </div>
{/if}


