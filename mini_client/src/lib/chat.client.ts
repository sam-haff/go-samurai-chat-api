import { writable, type Writable } from "svelte/store";
import { ApiChatMessage, ApiChatUser, chatConn, createChatID, extractToUIDFromChatID, type ChatConn } from "../chat-api/api";
import { WsOnlineStatusChange } from "../chat-api/websocket_api";
import type { Auth } from "firebase/auth";
import { SvelteMap } from "svelte/reactivity";

// svelte api integration
export class ChatApp {
    conn: ChatConn | null = null;   
    userStatuses: Map<string, Writable<boolean>> = new Map(); 
    chats: SvelteMap<string, Writable<ApiChatMessage[]>> = new SvelteMap();
    contacts: SvelteMap<string, ApiChatUser> = new SvelteMap();
    currentUserStore: Writable<ApiChatUser | null> = writable(null);
    currentUser: ApiChatUser | null = null;

    chatsList(): string[] {
        return Array.from(this.chats.keys());
    }
    async initChats() {
        let chats = await this.conn!.getChats();

        for (let chatID of chats) {
            this.trackChat(extractToUIDFromChatID(this.getAuth(), chatID)!);
        }
    }
    async updateCurrentUser(u: ApiChatUser | null) {
        this.currentUserStore.set(u);
        this.currentUser = u;
        await this.initContacts();
    }
    async initContacts() { // TODO: modify in accordance with type change
        let curr: ApiChatUser|null = this.currentUser;
        if (!curr) {
            // throw ?
            return;
        }
        this.contacts.clear();
        for (const [key, value] of Object.entries(curr!.contacts)) {
            console.log(this.contacts);
            try {
                let contact = await this.getConn().getUser(key);

                this.contacts.set(contact.uid, contact); 
            } catch (e: any) {
                console.log("Got invalid contact data. " + e.message);
                continue;
            }
        }
    }
    async addContact(username: string): Promise<ApiChatUser> {
        let u = await this.getConn().addContact(username);
        this.contacts.set(u.uid, u);

        return u;
    }
    async removeContact(username: string): Promise<ApiChatUser> {
        let u = await this.getConn().removeContact(username); // remove should only be by uid(because user could change username)
        this.contacts.delete(u.uid);

        return u;
    }
    getConn(): ChatConn {
        if (!this.conn) {
            console.trace("Expected chat conn, none found");
            throw new Error("No chat conn");
        }
        return this.conn!;
    }
    getAuth(): Auth {
        if (!this.getConn().fbAuth) {
            throw new Error("No auth");
        }
        return this.getConn().fbAuth!;
    }
    trackCurrentUser(): Writable<ApiChatUser|null>{
        return this.currentUserStore;
    }
    trackContacts(): SvelteMap<string, ApiChatUser> {
        return this.contacts;
    }
    trackContactsList(): ApiChatUser[] {
        return Array.from(this.contacts.values());
    }
    trackChat(withID: string): Writable<ApiChatMessage[]> {
        console.trace("Track chat with " + withID);
        if (withID === "" || withID === null) { throw new Error("Chat cannot be with empty UID"); }
        let chatID = createChatID(withID, this.getAuth().currentUser!.uid);

        if (!this.chats.has(chatID)){
            let w = writable<ApiChatMessage[]>([]);
            this.chats.set(chatID, w);
            this.getConn().getChatMessages(withID, 1024, Number.MAX_SAFE_INTEGER).then(
                (val: ApiChatMessage[]) =>{
                    if (!this.chats.has(chatID)) {
                        throw new Error("Chats stores should be initialized");
                    }
                    let w = this.chats.get(chatID);
                    w!.set(val);
                }
            );
            return w;
        }
        return this.chats.get(chatID)!;
    }
    trackOnlineStatus(uid: string): Writable<boolean> {
        if (!this.conn) {
            throw new Error("No conn");
        }
        if (this.userStatuses.has(uid)) {
            return this.userStatuses.get(uid)!;
        }
        let w = writable(false);
        this.userStatuses.set(uid, w)

        this.conn.subscribeOnlineStatusChange(uid, function(){});

        return w;
    }
    async updateContacts() {
        await this.updateCurrentUser(await this.getConn().getUser(this.getAuth().currentUser!.uid));
    }
    alive(): boolean {
        return this.conn != null && this.conn.alive();
    }
    finalize() {
        this.contacts.clear();
        this.chats.clear();
        if (this.conn) {
            this.conn.finalize();
        }
    }
    async singout() {
        await this.getAuth().signOut();
        this.finalize();
    }
}

let chatApp: ChatApp | null;

export function getChatApp(): ChatApp {
    if (!chatApp) {
        console.trace("Expected chat app instance but none found");
        throw new Error("No chat app instance present!");
    }
    return chatApp!;
}

export async function initChatApp() {
    if (chatApp) { return; }
    
    chatApp = new ChatApp();
    chatApp.conn = await chatConn(
        "", 
        "",
        function(u: ApiChatUser | null) { // onauth
            chatApp!.updateCurrentUser(u);
        },
        async function() { // onconnect
            if (!chatApp!.conn) {
                console.log("Failure on invariant: no WS connection/onconnect");
                return;
            }
            await chatApp!.initChats();
            for (let uid of chatApp!.userStatuses.keys()) {
                chatApp!.conn!.subscribeOnlineStatusChange(uid, function(){});
            }
        },
        function (msg: ApiChatMessage) { // onnewmessage
            if (!chatApp) {return;}

            let chatID = msg.conv_id;
            let w = chatApp!.trackChat(extractToUIDFromChatID(chatApp.getAuth(), chatID)!);
            let v: ApiChatMessage[] = []; 
            w.subscribe((val: ApiChatMessage[])=>{ v = val; });
            v.push(msg);
            w.set(v);
        },
        function (change: WsOnlineStatusChange) { //ononlinestatuschange
            if (chatApp!.userStatuses.has(change.uid)) {
                let s = chatApp!.userStatuses.get(change.uid);
                s!.set(change.online);
                return;
            }
            chatApp!.userStatuses.set(change.uid, writable(false));
        }
        
    ); 
}
function finalizeChatApp() {
    if (chatApp) {
        chatApp.finalize();
        chatApp = null;
    }
}
export async function resetChatApp() {
    finalizeChatApp();
    await initChatApp();
}
