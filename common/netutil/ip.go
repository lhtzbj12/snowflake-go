package netutil

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

func getFirstIP(ip string) (string, error) {
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}
	return "", errors.New("no valid ip found")
}

func GetClientIP(r *http.Request) (string, error) {
	ipInHeaders := [...]string{"X-Forwarded-For", "X-Real-IP", "Proxy-Client-IP", "WL-Proxy-Client-IP", "HTTP_CLIENT_IP", "HTTP_X_FORWARDED_FOR"}
	for _, ipInHeader := range ipInHeaders {
		ipStr := r.Header.Get(ipInHeader)
		ip, _ := getFirstIP(ipStr)
		if ip != "" {
			return ip, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}

func GetFirstNonLoopbackIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return "127.0.0.1"
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
