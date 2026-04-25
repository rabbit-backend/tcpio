package tcpio

import (
	"github.com/rabbit-backend/tcpio/epoll"
)

func (l *Listener) Start() {
	go l.eventLoop(l.epoller)

	for {
		conn, err := l.ln.Accept()
		if err != nil {
			l.onErrorEvent(err)
			continue
		}

		if err := l.epoller.Add(conn); err != nil {
			conn.Close()
			l.onErrorEvent(err)
		}
	}
}

func (l *Listener) eventLoop(epoller *epoll.Epoll) {
	for {
		connections, err := epoller.Wait()
		if err != nil {
			l.onErrorEvent(err)
			continue
		}

		for _, conn := range connections {
			if err := l.onConnEvent(conn); err != nil {
				if conn != nil {
					conn.Close()
				}

				if err := epoller.Remove(conn); err != nil {
					l.onErrorEvent(err)
					continue
				}

				l.onErrorEvent(err)
			}
		}
	}
}
