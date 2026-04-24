package tcpio

import (
	"net"

	"github.com/rabbit-backend/tcpio/epoll"
)

type Listener struct {
	ln           *net.TCPListener
	epoller      *epoll.Epoll
	onConnEvent  func(conn net.Conn) error
	onErrorEvent func(error)
}

type ListenerOptions struct {
	Addr         string
	OnConnEvent  func(conn net.Conn) error
	OnErrorEvent func(error)
}

func Listen(options *ListenerOptions) (*Listener, error) {
	ln, err := net.Listen("tcp", options.Addr)
	if err != nil {
		return nil, err
	}

	epoller, err := epoll.New()
	if err != nil {
		return nil, err
	}

	return &Listener{
		ln:      ln.(*net.TCPListener),
		epoller: epoller,
		onConnEvent: func(conn net.Conn) error {
			if options.OnConnEvent != nil {
				return options.OnConnEvent(conn)
			}

			return nil
		},
		onErrorEvent: func(err error) {
			if options.OnErrorEvent != nil {
				options.OnErrorEvent(err)
			}
		},
	}, nil
}

func (l *Listener) GetListener() *net.TCPListener {
	return l.ln
}
