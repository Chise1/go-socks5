package socks5

import (
	"errors"
	"fmt"
	"github.com/Chise1/go-socks5/net/proxy"
	"net"
	"regexp"
	"strings"
)

type SockClient struct {
	address string
}

var notParse = errors.New("not parse")

func to(req *Request, conn conn) error {
	for _, d := range dialers {
		var flag bool
		for _, f := range d.Include {
			if f(req.DestAddr.String()) {
				flag = true
				break
			}
		}
		if !flag {
			continue
		}
		for _, f := range d.Exclude {
			if f(req.DestAddr.String()) {
				flag = false
				break
			}
		}
		if !flag {
			continue
		}
		dialer := d.Dialer
		// 使用代理拨号器建立TCP连接
		target, err := dialer.Dial("tcp", req.DestAddr.String())
		if err != nil {
			msg := err.Error()
			resp := hostUnreachable
			if strings.Contains(msg, "refused") {
				resp = connectionRefused
			} else if strings.Contains(msg, "network is unreachable") {
				resp = networkUnreachable
			}
			if err := sendReply(conn, resp, nil); err != nil {
				return fmt.Errorf("Failed to send reply: %v", err)
			}
			return fmt.Errorf("Connect to %v failed: %v", req.DestAddr.String(), err)
		}
		defer target.Close()

		// Send success
		local := target.LocalAddr().(*net.TCPAddr)
		bind := AddrSpec{IP: local.IP, Port: local.Port}
		if err := sendReply(conn, successReply, &bind); err != nil {
			return fmt.Errorf("Failed to send reply: %v", err)
		}

		// Start proxying
		errCh := make(chan error, 2)
		go proxyData(target, req.bufConn, errCh)
		go proxyData(conn, target, errCh)

		// Wait
		for i := 0; i < 2; i++ {
			e := <-errCh
			if e != nil {
				// return from this function closes target (and conn).
				return e
			}
		}
		return nil
	}
	return notParse
}

var dialers []struct {
	Dialer  proxy.Dialer
	Include []func(string) bool
	Exclude []func(string) bool
}

func Init(socs []Socks) error {
	var newdialers []struct {
		Dialer  proxy.Dialer
		Include []func(string) bool
		Exclude []func(string) bool
	}

	for _, soc := range socs {
		var auth *proxy.Auth
		if soc.User != "" {
			auth = &proxy.Auth{
				User:     soc.User,
				Password: soc.Password,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", soc.Addr, auth, proxy.Direct)
		if err != nil {
			return errors.New("无法连接到代理" + soc.Addr + ": " + err.Error())
		}
		var includes []func(string) bool
		var excludes []func(string) bool

		for _, include := range soc.Include {
			if include.Type == "regexp" {
				compile, err := regexp.Compile(include.Value)
				if err != nil {
					return err
				}
				includes = append(includes, func(fqdn string) bool {
					return compile.MatchString(fqdn)
				})
			} else if include.Type == "cidr" {
				_, i, err := net.ParseCIDR(include.Value)
				if err != nil {
					return err
				}
				includes = append(includes, func(fqdn string) bool {
					ip := net.ParseIP(fqdn)
					if ip == nil {
						return false
					}
					return i.Contains(ip)
				})
			}
		}
		for _, exclude := range soc.Exclude {
			if exclude.Type == "regexp" {
				compile, err := regexp.Compile(exclude.Value)
				if err != nil {
					return err
				}
				excludes = append(excludes, func(fqdn string) bool {
					return compile.MatchString(fqdn)
				})
			} else if exclude.Type == "cidr" {
				_, i, err := net.ParseCIDR(exclude.Value)
				if err != nil {
					return err
				}
				excludes = append(excludes, func(fqdn string) bool {
					ip := net.ParseIP(fqdn)
					if ip == nil {
						return false
					}
					return i.Contains(ip)
				})
			}
		}
		newdialers = append(newdialers, struct {
			Dialer  proxy.Dialer
			Include []func(string) bool
			Exclude []func(string) bool
		}{Dialer: dialer, Include: includes, Exclude: excludes})
	}
	dialers = newdialers
	return nil
}
