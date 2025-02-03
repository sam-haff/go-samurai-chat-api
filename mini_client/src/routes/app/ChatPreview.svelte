<script lang="ts">
    // svelte
    import { Avatar, Indicator, Spinner } from "flowbite-svelte";
    import { ArrowRightOutline } from "flowbite-svelte-icons";
    // api
    import { getChatApp } from "$lib/chat.client";
    import { ApiChatMessage, ApiChatUser, extractToUIDFromChatID } from "../../chat-api/api";

    // props
    let { chatID, closeMenu, selectedChatID= $bindable() }: {
        closeMenu: any,
        chatID: string,
        selectedChatID: string,
    } = $props();

    let app = getChatApp();

    // state
    let loading = $state(false);

    let currentUser = app.trackCurrentUser()
    let withUID = extractToUIDFromChatID(app.getAuth(), chatID)!;
    let chatMessages = app.trackChat(withUID);
    let lastMessage = $derived($chatMessages.length ? $chatMessages[$chatMessages.length-1] : new ApiChatMessage());

    let withUser = $state(new ApiChatUser("", "", "", ""));
    let onlineStatus = app.trackOnlineStatus(withUID!);

    loadUser();
    async function loadUser() {
        loading = true;
        try {
            loading = true;

            let withUID = extractToUIDFromChatID(app.getAuth(), chatID);

            if (!withUID) {
                throw new Error("Failed to load last message. Wrong chat id.");
            }

            withUser = await app.getConn().getUser(withUID);
        } catch (e: any) {
            console.log("Failed to load preview chat user. " + e.message);
            loading = false;
        }
        loading = false;
    }

    // functions
    function truncate(input: string) {
        if (input.length > 13) {
            return input.substring(0, 13) + '...';
        }

        return input;
    };

    let indicatorClass = $derived("absolute top-8 left-8 w-2 h-2 z-10 " + ($onlineStatus ? "bg-green-500" : "bg-slate-400"));
</script>

<button onclick={()=>{ closeMenu(); selectedChatID = chatID; }} class="flex flex-row mt-2 group !shadow w-64 h-fit bg-gray-200 hover:bg-gray-300 p-2 rounded-2xl">
    {#if loading}
        <Spinner/>
    {:else}
        <div class="flex flex-col">
            <div class="relative">
                <Indicator class={indicatorClass} />
                <Avatar style="mask:radial-gradient(circle at 90% 90%, rgba(0,0,0,0) 5px, #fff 6px);mask-composite:exclude;"
                src={withUser.img_url} />
            </div>
        </div>
        <div class="flex flex-col ml-2 flex-1">
            <div class="flex flex-row items-start justify-start content-start">
                <span class="text-gray-600"><i><b>{withUser.username}</b></i></span>
            </div>
            <div class="flex flex-row">
                <div class="flex flex-col justify-center">
                    <Avatar src={lastMessage.from == withUser.uid ? withUser.img_url : $currentUser!.img_url} class="w-5 h-5"/>
                </div>
                <div class="flex flex-col">
                    <span class="ml-1 text-gray-950">{truncate(lastMessage.msg)}</span>
                </div>
            </div>
        </div>
        <div class="flex flex-col w-fit">
            <ArrowRightOutline class="hidden w-5 h-5 ml-auto group-hover:block"/>
        </div>
    {/if}
    </button>


