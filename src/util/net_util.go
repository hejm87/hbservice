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

type NetInterfaceInfo struct {
	Name			string		// 网卡接口名
	IpAddr			string		// ip地址
	MacAddr			string		// mac地址
	Mtu				int
}

func GetIfaceIpAddr(iface_name string) (ip string, err error) {
	ifaces, err := GetActiveNetInterfaces()
	if err != nil {
		return ip, err
	}
	for _, x := range ifaces {
		if iface_name == x.Name {
			return x.IpAddr, nil
		}
	}
	return "", errors.New("not exists iface")
}

func GetActiveNetInterfaces() (infos []NetInterfaceInfo, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return infos, err
	}
	for _, iface := range ifaces {
		if iface.Flags & net.FlagUp == 0 {
			continue
		}
		if iface.Flags & net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := get_ip_from_addr(addr)
			if ip != nil {
				info := NetInterfaceInfo {
					Name:		iface.Name,
					IpAddr:		ip.String(),
					MacAddr:	iface.HardwareAddr.String(),
					Mtu:		iface.MTU,
				}
				infos = append(infos, info)
				break
			}
		}
	}
	return infos, nil
}

func get_ip_from_addr(addr net.Addr) net.IP {
    var ip net.IP
    switch v := addr.(type) {
    case *net.IPNet:
        ip = v.IP
    case *net.IPAddr:
        ip = v.IP
    }   
    if ip == nil || ip.IsLoopback() {
        return nil 
    }
    ip = ip.To4()
    if ip == nil {
        return nil
    }
    return ip
}
