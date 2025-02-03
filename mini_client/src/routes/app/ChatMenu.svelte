<script lang="ts">
    // svelte
    import { Button, Sidebar, SidebarWrapper } from "flowbite-svelte";
    import { ArrowRightOutline } from "flowbite-svelte-icons";
    import { onMount } from "svelte";
    // custom components
    import ChatsList from "./ChatsList.svelte";

    // props
    let { open = $bindable(), openContactsModal = $bindable(), selectedChatID = $bindable(), hiddenMenu } : {
        selectedChatID: string,
        openContactsModal: boolean,
        open: boolean,
        hiddenMenu: boolean,
    } = $props();

    function closeMenu() {
        open = false;
    }

    let oldSelectedChatID = selectedChatID;

    // effects
    $effect(
        function () {
            if (selectedChatID !== oldSelectedChatID) {
                open = false;
            }
        }
    )

    // mount
    onMount(
        function() {
            let el = document.getElementById("slider");
            let contEl = document.getElementById("menu-container");
            if (!el || !contEl) {return;}
            el.onclick = function(ev: MouseEvent) {
                let r = contEl.getBoundingClientRect();
                if (ev.clientX < r.left || ev.clientX > r.right || ev.clientY > r.bottom || ev.clientY < r.top) {
                    open = false;
                    return;
                }
            }
        }
    )
</script>

<div id="slider" class="flex flex-col flex-1 z-10  h-full w-screen " class:slide-in={open} class:slide-out={!open} >
    <div id="menu-container" class="flex-1 w-fit p-2 border-r-4 z-10 backdrop-blur-lg bg-white/30">
        <Sidebar class="h-full w-72 text-sm mt-7 pt-1 overflow-hidden">
            <SidebarWrapper divClass="p-2 pr-0 rounded" class="py-0 ">
                <Button onclick={()=>{openContactsModal=true;}} color="none" outline class="group bg-[rgb(194,231,255)] w-44 border-2 hover:border-2 border-transparent hover:shadow-[1px_1px_6px_0px_rgba(60,60,60,0.5)] !text-gray-950">
                    <span class="mr-auto text-gray-950 group-hover:text-gray-950">Contacts</span>
                <ArrowRightOutline class="hidden ml-auto group-hover:block"/>
                </Button>
                <ChatsList {closeMenu} bind:selectedChatID/>
                
            </SidebarWrapper>
        </Sidebar>
    </div>
</div>

<style>
#slider {
    position: absolute;
    transform: translateX(-100%);
    -webkit-transform: translateX(-100%);
}

.slide-in {
    animation: slide-in 0.5s forwards;
    -webkit-animation: slide-in 0.5s forwards;
}

.slide-out {
    animation: slide-out 0.5s forwards;
    -webkit-animation: slide-out 0.5s forwards;
}
    
@keyframes slide-in {
    100% { transform: translateX(0%); }
}

@-webkit-keyframes slide-in {
    100% { -webkit-transform: translateX(0%); }
}
    
@keyframes slide-out {
    0% { transform: translateX(0%); }
    100% { transform: translateX(-100%); }
}

@-webkit-keyframes slide-out {
    0% { -webkit-transform: translateX(0%); }
    100% { -webkit-transform: translateX(-100%); }
}
</style>