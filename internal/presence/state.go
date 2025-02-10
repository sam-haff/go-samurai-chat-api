package presence

import cmap "go-chat-app-api/internal/concurrent-map"

type State struct {
	OnlineByUid cmap.ConcurrentMap[string, int]
}

func NewState() *State {
	return &State{
		OnlineByUid: cmap.New[int](),
	}
}

func (s *State) GetClientsConnectedNum(uid string) int {
	c, _ := s.OnlineByUid.Get(uid)
	return c
}
func (s *State) IsOnline(uid string) bool {
	c, ok := s.OnlineByUid.Get(uid)
	return ok && (c > 0)
}
