<script lang="ts">
    // svelte
    import { Avatar, Button, Card, Input, Spinner } from "flowbite-svelte";
    import { fade, scale } from "svelte/transition";
    import { cubicIn } from "svelte/easing";
    import { untrack } from "svelte";
    // custom components
    import ChatMessage from "./ChatMessage.svelte";
    // api
    import { getChatApp } from "$lib/chat.client";
    import { ApiChatMessage, ApiChatUser, ApiResp, ApiRespCode, extractToUIDFromChatID } from "../../chat-api/api";

    let app = getChatApp();
    let currentUser = app.trackCurrentUser();

    // props
    let { selectedChatID = $bindable() } = $props();
    let oldChatID = selectedChatID;

    // state
    let errorMsg = $state(""); // TODO: show msg in UI
    let loading = $state(true);
    let sendingMsg = $state(false);
    const zeroUser = new ApiChatUser("","","","");
    let withUser = $state(zeroUser);
    let messages: ApiChatMessage[] = [];
    //let quer='(max-width: 480px)'
    //let screenSupported = createMediaStore(query);

    let chatMessages = app.trackChat(extractToUIDFromChatID(app.getAuth(), selectedChatID)!)
    $effect( function() {
        if (chatMessages) {
        }
    } )
    $effect( function () {
        if (selectedChatID !== "") {
            chatMessages = app.trackChat(extractToUIDFromChatID(app.getAuth(), selectedChatID)!)
        }
        setTimeout(scrollToBottom, 150);
    } );
    $effect( function() {
        if ($chatMessages!.length > 0) {
            setTimeout(scrollToBottom, 150)
        } 
    } );

    // input
    let messageText = $state(""); 

    // effects
    $effect(()=>{
        oldChatID = selectedChatID;

        withUser = zeroUser;
        if (!untrack(()=>$currentUser)) {return; }
        loadChat(selectedChatID);
        
        setTimeout(scrollToBottom, 300)
    });

    // functions
    function scrollToBottom() {
        console.log("Scroll called");
        let elem = document.getElementById('chat-container');
        if (!elem) { return; }

        let diff = elem.scrollHeight - elem.scrollTop - elem.clientHeight;
        if (diff > 300 && diff < 400) {
            elem.scrollTo({ top: elem.scrollHeight, behavior: 'smooth' })
        } else {
            elem.scrollTop = elem.scrollHeight;
        }
    }
    async function loadUser(chatID: any) {
        if (chatID && chatID.length > 0) {
            let app = getChatApp();

            let withUid = extractToUIDFromChatID(app.getAuth(), chatID);
            if (!withUid) {
                throw new Error("Invalid chat ID. " + errorMsg);
            }

            let u = await app.getConn().getUser(withUid);
            withUser = u;
        }
    }

    async function loadChat(chatID: string) {
        loading = true;

        try {
            let app = getChatApp();

            await loadUser(chatID);
            if (withUser.uid.length == 0) {
                throw new Error("Failed to load user");
            }
        } catch (e: any) {
            errorMsg = e.message;
            console.log(e.message);
            loading = false;
        }

        loading = false;
    } 
    async function onsubmit(e: any) {
        e.preventDefault();

        let app = getChatApp();
        sendingMsg = true;
        app.getConn().sendMessage(withUser.uid, messageText, function (data:any){
            sendingMsg = false;
            let r = data as ApiResp;
            if (r.code !== ApiRespCode.Success) {
                console.log("Bad WS response. " + r.msg);
                // TODO: handle error
                return;
            }

            let msg = r.obj as ApiChatMessage;

            pinchOut();
        });       // TODO: add 'sending message' aki loading state
        messageText = "";
        /*
        let resp = await apiSendMessage(fbAuth!, withUser.uid, messageText);
        if (resp.code !== ApiRespCode.Success) {
            console.log("Failed to send message. " + resp.msg);
            // do what??
            return;
        }
        console.log("Message sent!");
        */
    }
    function pinchOut() {
        let appliedScale = 1 - Math.random()*0.01;
        document.querySelector('meta[name="viewport"]')!.setAttribute('content', "width=device-width, initial-scale=" + appliedScale);

        document.body.scrollTop = 0;
    }
</script>


{#key selectedChatID}
    <div id="cont" class="w-full h-full pb-36 chat-max-height" in:scale|global>
        <Card class="h-full w-full sm:w-4/5 !mx-auto chat-max-height max-w-[600px] mt-1 sm:mt-8 ml-0 !py-0 !px-0 !overflow-auto block">
            {#if loading}
                <Spinner class="m-4"/>
            {:else}
                <div class="flex flex-col topbar-container overflow-hidden h-full ">
                    <div class="shadow-sm sticky top-0 p-2">
                        <div class="flex flex-row items-center content-center">
                            <Avatar class="w-8 h-8" src={withUser.img_url}/> <span class="ml-4 text-gray-900"><i><b>{withUser.username}</b></i></span>
                        </div>
                        
                    </div>
                    <div  class="flex flex-row flex-1 p-2 overflow-auto ">
                        <div in:fade|global={{duration:550, easing: cubicIn}} id="chat-container" class="flex flex-col w-full overflow-auto [&::-webkit-scrollbar]:w-2
                            [&::-webkit-scrollbar-track]:rounded-full
                            [&::-webkit-scrollbar-track]:bg-transparent
                            [&::-webkit-scrollbar-thumb]:rounded-full
                            [&::-webkit-scrollbar-thumb]:bg-gray-300">
                            {#each $chatMessages! as message (message.msg_id)}
                                <ChatMessage ava_url={message.from == withUser.uid ? withUser.img_url : $currentUser!.img_url} msg={message}/>
                            {/each} 
                        </div>
                    </div>
                    <div class="shadow-sm mt-auto p-2">
                        <form onsubmit={onsubmit}>
                            <div class="flex flex-row">
                                <div class="flex flex-col flex-1">
                                    <Input class="text-lg" type="text" autocomplete="off" autofocus id="chat-input" bind:value={messageText} placeholder="Your message..."/>
                                </div>
                                <div class="flex ml-2 flex-col">
                                    <Button type="submit" outline class="my-auto">
                                        {#if sendingMsg}
                                            <Spinner size='5'/>
                                        {:else}
                                            <svg class="fill-current w-5 h-5" version="1.0" id="katman_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 30 30" style="enable-background:new 0 0 30 30;" xml:space="preserve"><polygon points="28.5,15 1.5,4.5 5.55,13.7 22.88,15 5.54,16.3 1.5,25.5 "/></svg>
                                        {/if}
                                    </Button>
                                </div>
                            </div>
                        </form>
                    </div>
                </div>
            {/if}
        </Card>
    </div>
{/key}

<style>
    .chat-max-height {
        min-height: 700px;
        max-height: 700px;
    }
    @media (max-height: 650px) {
        .chat-max-height{
            min-height: 500px;
            max-height: 500px;
        }
    }
    .topbar-container {
        position: relative;
    }
</style>