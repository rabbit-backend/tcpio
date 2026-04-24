package epoll

import (
	"log"
	"net"
	"reflect"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type Epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

func New() (*Epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		fd:          fd,
		connections: make(map[int]net.Conn),
		lock:        &sync.RWMutex{},
	}, nil
}

func (e *Epoll) Add(conn net.Conn) error {
	fd := getFD(conn)

	err := unix.EpollCtl(
		e.fd,
		syscall.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{
			Fd:     int32(fd),
			Events: unix.POLLIN | unix.POLLHUP,
		},
	)

	if err != nil {
		return err
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	e.connections[fd] = conn

	return nil
}

func (e *Epoll) Remove(conn net.Conn) error {
	fd := getFD(conn)

	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	delete(e.connections, fd)

	if len(e.connections)%100 == 0 {
		log.Printf("total number of connections: %v", len(e.connections))
	}

	return nil
}

func (e *Epoll) Wait() ([]net.Conn, error) {
	events := make([]unix.EpollEvent, 100)
retry:
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connections []net.Conn
	for i := range n {
		conn := e.connections[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}

func getFD(conn net.Conn) int {
	fd := reflect.Indirect(reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").FieldByName("fd")).FieldByName("pfd")
	return int(fd.FieldByName("Sysfd").Int())
}
