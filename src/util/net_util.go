package util

import (
	"net"
	"time"
	"errors"
)

const (
	RECV_TIMEOUT_MS int = 3000
)

func SetTimeout(conn net.Conn, timeout *time.Time) {
	if timeout != nil {
		conn.SetReadDeadline(*timeout)
	}
}

func NetRecv(conn net.Conn, buf []byte, size int, timeout_ms int) (int, error) {
	var timeout *time.Time = nil
	if timeout_ms > 0 {
		duration := time.Duration(timeout_ms)
		timeout = new(time.Time)
		*timeout = time.Now().Add(duration * time.Millisecond)
	}
	return NetRecvTimeout(conn, buf, size, timeout)
}

func NetRecvTimeout(conn net.Conn, buf []byte, size int, timeout *time.Time) (int, error) {
	if len(buf) < size {
		return -1, errors.New("NetRecv buf not enough")
	}
	read_size := 0
	for {
		SetTimeout(conn, timeout)
		if n, err := conn.Read(buf[read_size:]); err == nil {
			read_size += n
			if n == 0 {
				break
			} else if read_size >= size {
				break
			}
		} else {
			return read_size, err
		}
	}
	return read_size, nil
}
