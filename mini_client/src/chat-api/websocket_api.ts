import type { Auth } from "firebase/auth";
import type { ApiChatMessage, ApiResp } from "./api";

const endpoint = "ws://127.0.0.1:80/ws";

enum WSChatEventType{
    NewMessage = 0,
    SendMessageRequest  = 1,
    WsEvent_CheckOnlineRequest           = 2,
	WsEvent_OnlineStatusChangeEvent      = 3,
	WsEvent_OnlineStatusSubscribeRequest = 4,
}

class WSEvent {
    event_type: Number;
    is_response: boolean;
    id: string;
    obj: any;

    constructor (event_type: Number, is_reponse: boolean, id: string, obj: any) {
        this.event_type = event_type;
        this.is_response = is_reponse;
        this.id = id;
        this.obj = obj;
    }
}

export class WsOnlineStatusChange {
    uid: string = "";
    online: boolean = false;
}

export type WSCallback = (data: any) => void;
export type NewMessageHandler = (msg: ApiChatMessage) => void;
export type OnlineStatusChangeHandler = (change: WsOnlineStatusChange) => void;
export class WSChatClient{
    wsConn: WebSocket|null = null;
    responseCallbacks: Map<string, WSCallback> = new Map<string, WSCallback>();

    onNewChatMessage: NewMessageHandler = function(msg:ApiChatMessage){
        console.log("New message: " + msg.msg);
    };
    onOnlineStatusChange: OnlineStatusChangeHandler = function(change: WsOnlineStatusChange) {
        console.log("Status of ", change.uid, " is ", change.online);
    }
    connected(): boolean {
        if (!this.wsConn) {
            return false;
        }
        return this.wsConn!.readyState === this.wsConn!.OPEN
    }
    async sendMessage(to: string, msg: string, callback: WSCallback)  {
        let params: any = new Object();
        params.to = to;
        params.text = msg;
       
        let event = new WSEvent(WSChatEventType.SendMessageRequest, false, crypto.randomUUID(), params);

        this.wsConn!.send(JSON.stringify(event))
        this.responseCallbacks.set(event.id, callback);
    }

    async subsribeOnUserStatusChange(uid: string, callback: WSCallback) {
        let params: any = new Object();
        params.uid = uid;
        
        let event = new WSEvent(WSChatEventType.WsEvent_OnlineStatusSubscribeRequest, false, crypto.randomUUID(), params);

        this.wsConn!.send(JSON.stringify(event));
        this.responseCallbacks.set(event.id, callback); // TODO: uncomment TODO: check if callbask is undefined
    }

    finalize() {
        if (!this.wsConn) {return;}
        this.wsConn.onopen = null;
        this.wsConn.onclose= null;
        this.wsConn.onerror = null;
        this.wsConn.onmessage = null;
        this.wsConn.close();
    }
}

export async function initWSChatClient(auth: Auth, onconnect: any): Promise<WSChatClient|null> {
    console.log("Init WS chat");

    if (!auth.currentUser){
        return null
    }

    try {
        let client = new WSChatClient()

        client.wsConn = new WebSocket(endpoint + "?token=" + await auth.currentUser!.getIdToken())
        console.log("Got WS conn");
        console.log(client.wsConn);
        client.wsConn.onmessage = function (ev: MessageEvent) {
            try {
                let event = <WSEvent>JSON.parse(ev.data);

                let eventType = event.event_type;
                let eventId = event.id;
                console.log("New ws event");
                console.log(event);
                if (eventType === WSChatEventType.NewMessage) {
                    // handle new chat message
                    let res = <ApiResp>event.obj.result;
                    if (!res) {
                        throw new Error("Ill formed event");
                    }
                    let resp = <ApiResp>res;
                    client.onNewChatMessage(<ApiChatMessage>resp.obj) 
                    return;
                }
                if (eventType === WSChatEventType.WsEvent_OnlineStatusChangeEvent) {
                    let payload = event.obj;
                    console.log("New online change event. Got obj: ");
                    console.log(payload);

                    client.onOnlineStatusChange(payload);
                    return;
                }
                if (event.is_response)  {
                    console.log("Got WS response");
                    console.log(event);
                    let callback = client.responseCallbacks.get(eventId);
                    if (callback) {
                        callback(event.obj.result);
                    }
                    return;
                }
            } catch (e: any) {
                console.log("Failed to process event. " + e.message);
                return;
            }
        }
        
        let setupWSConn = function(c: WebSocket) {
            c.onopen = onconnect;
            c.onclose = function(e: any) {
                console.log('Socket is closed. Reconnect will be attempted in 1 second.', e.reason);
                setTimeout(async function() {
                    client.wsConn = new WebSocket(endpoint + "?token=" + await auth.currentUser!.getIdToken());
                    setupWSConn(client.wsConn);
                }, 1000);
            }
            c.onerror = function(err) {
                console.error('Socket encountered error: ', err, 'Closing socket');
                client.wsConn!.close();
            };
        }
        setupWSConn(client.wsConn);

        return client
    } catch (e: any) {
        console.log("Failed to init WS client. " + e.message);
        return null;
    }
}

