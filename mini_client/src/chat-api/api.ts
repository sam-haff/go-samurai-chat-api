import { initializeApp, type FirebaseApp } from "firebase/app";
import { getAuth, onAuthStateChanged, signInWithEmailAndPassword, type Auth } from "firebase/auth";
import type firebase from "firebase/compat/app";
import { initWSChatClient, WsOnlineStatusChange, type WSCallback, type WSChatClient } from "./websocket_api";
import Chat from "../routes/app/Chat.svelte";

export enum ApiRespCode{
	Success                         = 0,
	UserAlreadyRegistered           = 1,
	UsernameTaken                   = 2,
	UsernameFormatNotValid          = 3,
	ReceiverDoesNotExist            = 4,
	EmailProfileAuthInvariantBroken = 5,
	CantCreateAuthUser              = 6,
	UserNotRegistered               = 7,
	NotAuthenticated                = 8,
	MaximumTokensNumberReached      = 9,
	DeviceNameTooLong               = 10,
	InvalidArgs                     = 11,
	Failure                         = 12 //next code //cant be returned from a server yet
};
const apiUrl = "http://127.0.0.1:8080";

export class ApiChatUser {
    email: string;
    username: string;
    uid: string;
    img_url: string;
    contacts: any;

    constructor (email: string, username: string, uid: string, imgUrl: string) {
        this.email = email;
        this.username = username;
        this.uid = uid;
        this.img_url = imgUrl;
    }
}

export class ApiChatMessage {
    msg_id: string = "";
    to: string = "";
    from : string = "";
    username: string = "";
    img_url: string = "";
    conv_id: string = "";
    created_at: string = "";
    msg: string = "";


}
export function extractToUIDFromChatID(auth: Auth, chatID: string ): string|null {
    if (!auth.currentUser) {
        console.trace("No auth!");
        //throw new Error("Auth is needed to extract id from Chat ID");
        return null;
    }

    let i = chatID.indexOf(auth.currentUser!.uid);
    if (i === -1) {
        return null;
    }
    if (i === 0) {
        return chatID.substring(i+auth.currentUser.uid.length);
    }
    return chatID.substring(0, i);
}
function extractUserFromResp(resp: ApiResp) : ApiChatUser|null {
    if (resp.obj === null) { return null; }

    let userObj = resp.obj;
    return resp.obj;
}
function extractChatsIDs(resp: ApiResp) : string[] | null{
    if (resp.obj === null) { return null; }

    let chatsObj = resp.obj;
    return chatsObj;
}
async function extractChatsIDsAsync(resp: Promise<ApiResp>) : Promise<string[] | null>{
    let r = await resp;

    return extractChatsIDs(r)
}
function extractChatMessages(resp: ApiResp) : ApiChatMessage[] | null {
    if (resp.obj === null) {
        return null;
    }
    return (resp.obj as ApiChatMessage[]).reverse();
}
export function createChatID(uid1: string, uid2: string): string {
    let uids = [uid1, uid2];
    uids.sort() 

    return uids[0] + uids[1];
}
export class ApiResp {
    code: ApiRespCode;
    msg: string;
    obj: any;

    constructor (code: ApiRespCode, msg: string, obj: any) {
       this.code = code;
       this.msg = msg;
       this.obj = obj; 
    }
}

const generalFailureResp = new ApiResp(ApiRespCode.Failure, "General error", null);

function makeGeneralErrorResp(msg: string) : ApiResp{
    return new ApiResp(ApiRespCode.Failure, msg, null);
}

function makeHeaders() : HeadersInit {
    return {
        'Content-Type':'application/json'
    };
}
function makeHeadersWithAuth(token: string) : HeadersInit {
    return {
        'Content-Type':'application/json',
        'Authorization': 'Bearer ' + token,
    };
}

async function getApiRespResult(resp: Response) : Promise<ApiResp> {
    let respJson : any;
    try {
        respJson = await resp.json();
    } catch (e: any) {
        return makeGeneralErrorResp(e.message);
    }
    let apiRes = respJson.result;

    return new ApiResp(apiRes.code, apiRes.msg, apiRes.obj);
}

async function apiRegister(email: string, username: string, pwd: string) : Promise<ApiResp> {
    const resp = await fetch(
        apiUrl + '/register',
        {
            method: "POST",
            body: JSON.stringify({
                'username':username,
                'email':email,
                'pwd':pwd,
            }),
            headers: makeHeaders(),
        }
    );

    return await getApiRespResult(resp);
}
async function apiRemoveContact(auth: Auth, username: string) : Promise<ApiResp> {
    const resp = await fetch(
        apiUrl + '/removecontact',
        {
            method: "POST",
            body: JSON.stringify({
                'username': username,
            }),
            headers: makeHeadersWithAuth(await auth.currentUser!.getIdToken())
        }
    );

    return await getApiRespResult(resp);
}
async function apiAddContact(auth: Auth, username: string) : Promise<ApiResp> {
    const resp = await fetch(
        apiUrl + '/addcontact',
        {
            method: "POST",
            body: JSON.stringify({
                'username': username,
            }),
            headers: makeHeadersWithAuth(await auth.currentUser!.getIdToken())
        }
    );

    return await getApiRespResult(resp);
}

async function apiUpdateAvatarFile(auth: Auth, file: File) : Promise<ApiResp> {
    if (!auth.currentUser) {
        return makeGeneralErrorResp("Failed: not authorized");
    }

    return new Promise(function(resolve, reject) {
        let r = new FileReader()
        r.readAsArrayBuffer(file)
        r.onload = async function() {
            resolve(await getApiRespResult(await fetch(
                apiUrl + '/updateavatarfile',
                {
                    method: "POST",
                    body: r.result,
                    headers: makeHeadersWithAuth(await auth.currentUser!.getIdToken())
                }
            )))
        }
        r.onerror = async function(ev: ProgressEvent<FileReader>) {
            reject(new ApiResp(ApiRespCode.Failure, "Failed to load image client side", null));
        }
    });
}
async function apiUpdateAvatar(auth: Auth, avaUrl: string) : Promise<ApiResp> {
    if (!auth.currentUser) {
        return makeGeneralErrorResp("Failed: not authorized");
    }
    
    const resp = await fetch(
        apiUrl + '/updateavatar',
        {
            method: "POST",
            body: JSON.stringify({
                'img_url':avaUrl,
            }),
            headers:  makeHeadersWithAuth(await auth.currentUser.getIdToken())
        }
    );

    return await getApiRespResult(resp);
}
async function apiSendMessage(auth: Auth, toUid: string, msg: string) : Promise<ApiResp> {
    if (!auth.currentUser) {
        return makeGeneralErrorResp("Failed: not authorized");
    }
    const resp = await fetch(
        apiUrl + "/addmessage",
        {
            method: "POST",
            body: JSON.stringify({
                'to':toUid,
                'text':msg,
            }),
            headers: makeHeadersWithAuth(await auth.currentUser.getIdToken())
        }
    );

    return await getApiRespResult(resp);
}
async function apiGetChats(auth: Auth) : Promise<ApiResp> {
    if (!auth.currentUser) {
        return makeGeneralErrorResp("Failed: not authorized");
    }
    const resp = await fetch(
        apiUrl + "/chats",
        {
            method: "GET",
            headers: makeHeadersWithAuth(await auth.currentUser.getIdToken())
        }
    );

    return await getApiRespResult(resp);
}
async function apiGetChatMessages(auth: Auth, withUID: string, count: number, before: number) : Promise<ApiResp> {
    if (!auth.currentUser) {
        return makeGeneralErrorResp("Failed: not authorized");
    }
    const resp = await fetch(
        apiUrl + "/chat",
        {
            method: "POST",
            body: JSON.stringify({
                'with':withUID,
                'limit':count,
                'before_timestamp': before,
                'inverse': false, // TODO: pass it as argument
            }),
            headers: makeHeadersWithAuth(await auth.currentUser.getIdToken())
        }
    );

    return await getApiRespResult(resp);
}
async function apiGetUser(auth: Auth, uid: string) : Promise<ApiResp> {
    if (!auth.currentUser) {
        return makeGeneralErrorResp("Failed: not authorized");
    }
    const resp = await fetch(
        apiUrl + "/users/id/" + uid,
        {
            method: "GET",
            headers: makeHeadersWithAuth(await auth.currentUser.getIdToken())
        }
    );

    return await getApiRespResult(resp);
}

async function apiGetUserByUsername(auth: Auth, username: string) : Promise<ApiResp> {
    if (!auth.currentUser) {
        return makeGeneralErrorResp("Failed: not authorized");
    }
    const resp = await fetch(
        apiUrl + "/users/username/" + username,
        {
            method: "GET",
            headers: makeHeadersWithAuth(await auth.currentUser.getIdToken())
        }
    );

    let apiResp = await getApiRespResult(resp);

    return apiResp;
}

// TODO: make custom error type with human readable error
function connNotInitializedErr(method: string): Error {
    return new Error("Connection should be initialized/" + method);
}
function illFormedResponseErr(method: string): Error {
    return new Error("Ill formed response/" + method);
}
function websocketClosedErr(method: string): Error {
    return new Error("Websocket connection should be open, retry after reconnection/" + method);
}
function throwOnFailedRequest(resp: ApiResp, prefix?: string) {
    if (resp.code != ApiRespCode.Success) {
        if (prefix) throw new Error(prefix + resp.msg);
        throw new Error(resp.msg);
    } 
}
export class ChatConn {
    wsChatConn: WSChatClient | null = null;
    fbAuth: Auth | null = null;
    fbApp: FirebaseApp | null = null;
    apiEndpoint: string = "";
    wsEndpoint: string = "";

    finalize() {
        //this.fbAuth = null;
        //this.fbApp = null;
        this.wsChatConn?.finalize();
        //this.wsChatConn = null;
    }
    alive(): boolean {
        return this.fbApp != null && this.fbAuth != null && this.fbAuth.currentUser != null;// && this.wsChatConn != null;
    }
    async addContact(username: string): Promise<ApiChatUser> {
        if (!this.alive()) {
            throw connNotInitializedErr("addContact")
        }

        let resp = await apiAddContact(this.fbAuth!, username);
        throwOnFailedRequest(resp, "Failed to add contact. ");
        
        let user = extractUserFromResp(resp)
        if (!user) {
            throw illFormedResponseErr("addContact")
        }
        return user!;
    }
    async removeContact(username: string): Promise<ApiChatUser> {
        if (!this.alive) {
            throw connNotInitializedErr("removeContact");
        }

        let resp = await apiRemoveContact(this.fbAuth!, username);
        throwOnFailedRequest(resp, "removeContact");

        let u = extractUserFromResp(resp);
        if (!u) {
            throw illFormedResponseErr("removeContact");
        }
        return u!;
    }
    async sendMessage(to: string, msg: string, cb: (data: any)=>void) {
        if (!this.alive()) {
            throw connNotInitializedErr("sendMessage")
        }
        if (!this.wsChatConn!.connected) {
            throw websocketClosedErr("sendMessage");
        }
        await this.wsChatConn!.sendMessage(to, msg, cb);
    }
    async getUser(uid: string): Promise<ApiChatUser> {
        console.trace("Get User");
        if (!this.alive()) {
            throw connNotInitializedErr("getUser");
        }
        let resp = await apiGetUser(this.fbAuth!, uid);
        throwOnFailedRequest(resp, "Failed to get user. ")
        
        let u = extractUserFromResp(resp);
        if (!u) {
            throw illFormedResponseErr("getUser");
        }
        return u
    }
    async getUserByUsername(username: string): Promise<ApiChatUser> {
        if (!this.alive()) {
            throw connNotInitializedErr("getUserByUsername");
        }
        let resp = await apiGetUserByUsername(this.fbAuth!, username);
        throwOnFailedRequest(resp, "Failed to get user by username. ")

        let u = extractUserFromResp(resp)
        if (!u) {
            throw illFormedResponseErr("getUserByUsername");
        }
        return u;
    }
    async updateAvatar(img: File) {
        if (!this.alive()) {
            throw connNotInitializedErr("updateAvatar");
        }
        let resp = await apiUpdateAvatarFile(this.fbAuth!, img);
        throwOnFailedRequest(resp, "Failed to update avatar. ")
    }
    async signup(email: string, username: string, pwd: string) {
        let resp = await apiRegister(email, username, pwd);
        throwOnFailedRequest(resp, "Failed to register new user. ")
    }
    async singin(email: string, pwd: string, cb?: (uid?: string)=>void) {
        if (!this.fbAuth) {
            throw connNotInitializedErr("signin");
        }
        
        let u = await signInWithEmailAndPassword(this.fbAuth, email, pwd);
        if (cb) {
            cb(u.user.uid);
        }
    }
    async getChats(): Promise<string[]>{
        if (!this.alive()) {
            throw connNotInitializedErr("getChats");
        }
        let resp = await apiGetChats(this.fbAuth!);
        throwOnFailedRequest(resp, "Failed to get chats. ");

        let chatIDs = extractChatsIDs(resp)
        if (!chatIDs) {
            throw illFormedResponseErr("getChats");
        }
        return chatIDs;
    }
    async getChatMessages(withUid: string, count: number, before: number): Promise<ApiChatMessage[]> {
        if (!this.alive()) {
            throw connNotInitializedErr("getChatMessages");
        }
        let resp = await apiGetChatMessages(this.fbAuth!, withUid, count, before);
        throwOnFailedRequest(resp, "Failed to retrieve chat messages. ");
        let m = extractChatMessages(resp);
        if (!m) {
            throw illFormedResponseErr("getChatMessages");
        }

        return m;
    }
    async subscribeOnlineStatusChange(uid: string, cb: WSCallback) {
        if (!this.alive()) {
            throw connNotInitializedErr("subscribeOnlineStatusChange")
        }
        await this.wsChatConn!.subsribeOnUserStatusChange(uid, cb)
    }
}

export async function chatConn(
    apiEndpoint: string,
    wsEndpoint: string, 
    onauth: (uid: ApiChatUser|null)=>void, 
    onconnect: ()=>void, 
    onnewmessage: (msg:ApiChatMessage)=>void,
    ononlinestatuschange:(status: WsOnlineStatusChange)=>void): Promise<ChatConn> {

    let conn = new ChatConn()
    const firebaseConfig = {
            apiKey: "AIzaSyC0XbFWCwvHTRUoflcMjnmWAQ_ypQ6AzaU",
            authDomain: "chat-app-67931.firebaseapp.com",
            projectId: "chat-app-67931",
            storageBucket: "chat-app-67931.appspot.com",
            messagingSenderId: "1004659951088",
            appId: "1:1004659951088:web:9bbccd0daae58e164c4935"
        };
 
    conn.apiEndpoint = apiEndpoint;
    conn.wsEndpoint = wsEndpoint;

    conn.fbApp = initializeApp(firebaseConfig, "MyApp");
    conn.fbAuth = getAuth(conn.fbApp);
    onAuthStateChanged(conn.fbAuth,  async (user) => {
        if (user) {
            const uid = user.uid;
            try {
                let u = await conn.getUser(uid);
                if (conn.wsChatConn && conn.wsChatConn.connected()) {
                    conn.wsChatConn.wsConn!.close();
                    conn.wsChatConn.finalize();
                }
                conn.wsChatConn = await initWSChatClient(conn.fbAuth!, onconnect);
                conn.wsChatConn!.onOnlineStatusChange = ononlinestatuschange;
                conn.wsChatConn!.onNewChatMessage = onnewmessage;
                onauth(u);
                return;
            } catch (e:any) {
                console.log("Error while auth. " + e.message);
                console.log(`${e}`)
                console.log(e.stack)
                console.trace("Auth error")
            }
            //TODO: listen somewhere for username updates(from ws?)
        } else {
            onauth(null);
        }
        }
    );
    
    return conn;
}