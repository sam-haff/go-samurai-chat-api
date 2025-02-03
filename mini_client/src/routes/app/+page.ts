// svelte
import { redirect } from "@sveltejs/kit";
import { browser } from "$app/environment";
// api
import { getChatApp, initChatApp } from "$lib/chat.client";

export async function load() {
    if (browser) {
        await initChatApp();
        await getChatApp().getAuth().authStateReady();
        
        let u = getChatApp().getAuth().currentUser; // for printing token to then use it with postman for testing
        if (!u) {
            redirect(307, "/"); // back to login/reg screen
        }
        
        return;
    }

    return;
}