<script lang="ts">
    import { getChatApp } from "$lib/chat.client";
    import {validateUsername, validateEmail, validatePassword} from "$lib/validation.client"
    import { fade } from "svelte/transition";
    import { goto } from "$app/navigation";
    import { Label, Input, Dropzone, Tooltip } from "flowbite-svelte";
    import { EnvelopeSolid, LockSolid, UserSolid } from  'flowbite-svelte-icons'
    import ChatSubmit from "./SubmitButton.svelte";

    // props
    let {onfade, errorMessages=$bindable(), loading=$bindable()} : {
        onfade: any,
        errorMessages: string[],
        loading: boolean,
    } = $props()

    // state
    let selectedFile: File | undefined;
    let previewURL = $state("");

    // todo: to use
    const dropHandle = (event: any) => {
        event.preventDefault();

        if (event.dataTransfer.items) {
            let file = event.dataTransfer.items[0];
            if (file.kind === 'file'){
                selectedFile = file;
            }
        }
    };

  const handleChange = (e: any) => {
    e.preventDefault();

    const files = e.target.files;
    console.log(files);

    if (files.length > 0) {
        selectedFile = files[0];
        previewURL = URL.createObjectURL(selectedFile!);
    }
  };

    
    const handleSubmit = async (e: any) => {
        e.preventDefault();

        loading = true;
        try {
            const formData = new FormData(e.target);
            const email = formData.get('email')?.toString();
            const pwd = formData.get('pwd')?.toString();
            const username = formData.get('username')?.toString();

            if (!email || !pwd || !username || !selectedFile) {
                throw new Error("Please, enter all info and upload an avatar");
            }

            if (!validateEmail(email)) {
                throw new Error("Please, enter correct email");
            }
            if (!validateUsername(username)) {
                throw new Error("Username should only contain characters and numbers, and be at least 4 characters long");
            }
            if (!validatePassword(pwd)) {
                throw new Error("Password should be at least 6 characters long");
            }

            let app = getChatApp();

            await app.getConn().signup(email!, username!, pwd!)

            // login

            try {
                await app.getConn().singin(email!, pwd!);
            } catch (e:any){
                throw new Error(e.message); // TODO: do human readable errors
            }

            await app.getConn().updateAvatar(selectedFile);

            await app.updateCurrentUser(await app.getConn().getUser(app.getAuth().currentUser!.uid));

            loading = false; 

            goto("/app").catch((e: any) => {
                console.log("Failed to navigate to app");
            })
        } catch (e: any) {
            errorMessages.push(e.message)
            loading = false;
            return;
        }
	}
</script>

<form onsubmit={handleSubmit} >
    <!--mt-8-->
    <div class="flex flex-col mt-8 " transition:fade={{duration:500}} onoutroend={onfade}>
        <div id="avadiv" class="self-center">
            <Dropzone onchange={handleChange} accept=".png,.jpg" id="ava" class="self-center rounded w-20 h-20 z-0">
                {#if previewURL}
                    <img class="w-full h-full object-cover" src={previewURL} alt="Avatar preview"/>
                {:else}
                    <svg aria-hidden="true" class="m-auto w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" /></svg>
                {/if}
            </Dropzone>
        </div>
        <Tooltip placement="right" triggeredBy="#avadiv">Upload your avatar!</Tooltip>

        <Label for="input-group-1" class="block mb-2">Email</Label>
        <Input id="email" name="email" type="email" placeholder="">
            <EnvelopeSolid slot="left" class="w-5 h-5 text-gray-500 dark:text-gray-400" />
        </Input>
        <Label for="input-group-2" class="block mb-2">Username</Label>
        <Input autocapitalize="off" id="username" name="username" type="text" placeholder="">
            <UserSolid slot="left" class="w-5 h-5 text-gray-500 dark:text-gray-400" />
        </Input>
        <Label for="input-group-3" class="block mb-2">Password</Label>
        <Input id="pwd" name="pwd" type="password" placeholder="">
            <LockSolid slot="left" class="w-5 h-5 text-gray-500 dark:text-gray-400" />
        </Input>
        <ChatSubmit label="Sign Up!" bind:loading={loading}/>
    </div>
</form>
