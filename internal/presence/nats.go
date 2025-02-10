package presence

import (
	"log"

	"github.com/nats-io/nats.go"
)

const (
	NATSNewChatUserConn  = "new-chat-user-conn"
	NATSLostChatUserConn = "lost-chat-user-conn"
)

func GetNATSOnlineStatusChangeSubject(uid string) string {
	return uid + "-online-change"
}

type NATSOnlineStatusChange struct {
	Uid    string
	Online bool
}

func (s *State) RegisterNatsListeners(conn *nats.Conn) func() {
	newConnSub, _ := conn.Subscribe(NATSNewChatUserConn, func(msg *nats.Msg) {
		log.Println("Conn event")
		uid := string(msg.Data)

		shard := s.OnlineByUid.GetShard(uid)

		shard.Lock()
		cnt, ok := shard.UnsafeGet(uid)
		if !ok || cnt == 0 {
			shard.UnsafeSet(uid, 1)
		} else {
			shard.UnsafeSet(uid, cnt+1)
		}
		shard.Unlock()
		if !ok || cnt == 0 {
			// do this ouside of lock block cause want to spend least time in the lock
			conn.Publish(GetNATSOnlineStatusChangeSubject(uid), []byte{1})
		}
	})

	lostConnSub, _ := conn.Subscribe(NATSLostChatUserConn, func(msg *nats.Msg) {
		uid := string(msg.Data)

		shard := s.OnlineByUid.GetShard(uid)

		shard.Lock()
		cnt, ok := shard.UnsafeGet(uid)
		if ok && cnt > 0 {
			shard.UnsafeSet(uid, cnt-1)
		}
		shard.Unlock()
		if cnt == 1 { // do this ouside of lock block cause want to spend least time in the lock
			conn.Publish(GetNATSOnlineStatusChangeSubject(uid), []byte{0})
		}
	})

	return func() {
		newConnSub.Unsubscribe()
		lostConnSub.Unsubscribe()
	}
}
