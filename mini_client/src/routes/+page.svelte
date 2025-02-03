<script lang='ts'>
    import { getChatApp } from "$lib/chat.client";
    import { getAuth } from "firebase/auth";
    import { goto } from "$app/navigation";
    import { Button, ButtonGroup, Heading } from "flowbite-svelte";
    import Errortoasts from "../components/errortoasts.svelte";
    import Register from "./register.svelte";
    import Login from "./login.svelte";

    let app = getChatApp();

    // state
    let mode = $state("login");
    let loading = $state(false);
    let errorMessages = $state<string[]>([]);

    class PageListgroupOption{
        content: string = "";
        mode: string = "";
        get current(): boolean {
            return this.mode === mode;
        }

        constructor (content: string, mode: string) {
            this.content = content;
            this.mode = mode;
        }
    }

    let optionsList = [ 
        new PageListgroupOption("Login", "login"),
        new PageListgroupOption("Register", "register"),
    ];

    let loggedIn = $state(false)
    function updateLoggedIn() {
        let auth = getAuth();
        loggedIn = (auth.currentUser != null);
        console.log(loggedIn); // TODO: del
    }

    let currentUser = app.trackCurrentUser();

    $effect( function (){
        if ($currentUser && !loading) {
            goto('/app');
        }
    });
</script>

<div class="flex flex-col relative items-center justify-center content-center h-screen pb-36 sm:pb-20">
    <Errortoasts bind:messages={errorMessages}/>
    <div class="flex flex-row">
        <div class="flex flex-col">
            <div class="flex-row content-center mx-auto items-center justify-center">
                <Heading tag="h1" class="kaushan-script-regular mb-4">
                    Samurai
                </Heading>
            </div>
            <div class="flex-row content-center justify-center items-center">
                <div class="w-fit mx-auto">
                    <!--mt-8-->
                <ButtonGroup class="mx-auto mt-8">
                    {#each optionsList as optBtn (optBtn)}
                        <Button color={ mode===optBtn.mode ? "blue" : "light" } on:click={()=>{ mode = "transition"; }}>
                            {optBtn.content}
                        </Button>
                    {/each}
                </ButtonGroup>
                </div>
            </div>
        </div>
    </div>
    <div class="flex-row" >
        {#if mode === "register"}
            <Register bind:errorMessages bind:loading onfade={() => { mode = "login" }}/>
        {:else if mode === "login"}
            <Login bind:errorMessages bind:loading onfade={()=>{ mode = "register" }}></Login>
        {/if}
    </div>
</div>
