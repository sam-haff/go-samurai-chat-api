import { browser } from "$app/environment";
import { resetChatApp} from "$lib/chat.client";

export async function load() {
    if (browser) {
        resetChatApp();
    }
}