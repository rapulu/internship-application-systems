package main

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"fmt"
	"net"
	"time"
	"log"
	"os"
)

func main() {

	payloadV4, _ := createICMP(ipv4.ICMPTypeEcho)
	addr := "google.com"
	ttl := 52
	for {
		ploss, dur, err := PingIpv6(addr, ttl, payloadV4)
		if err != nil {
			log.Printf("host: %s, Loss: %d, error: %s\n", addr, ploss, err)
			return
		}
		log.Printf("host: %s, Loss: %d, RTT: %s\n", addr, ploss, dur)
		time.Sleep(1 * time.Second)
	}
}


func PingIpv4(address string, ttl int, payload []byte)(int, time.Duration, error){
	netaddr, err := net.ResolveIPAddr("ip4", address)

	conn, err := net.Dial("ip4:icmp", netaddr.String())

	if ttl == 0 {
		ttl = 60
	}

	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	opts := ipv4.NewConn(conn)

	if err := opts.SetTTL(ttl); err != nil {
		return 0, 0, fmt.Errorf("set TTL %d: %s", ttl, err)
	}

	return ping(conn, payload)
}

func PingIpv6(address string, ttl int, payload []byte) (int, time.Duration, error){
	netaddr, err := net.ResolveIPAddr("ip", address)

	conn, err := net.Dial("ip:ipv6-icmp", netaddr.String())

	if err != nil {
		return 0, 0, fmt.Errorf("set HopLimit %d: %s", ttl, err)
	}

	if ttl == 0 {
		ttl = 60
	}

	defer conn.Close()

	opts := ipv6.NewConn(conn)
	if err := opts.SetHopLimit(ttl); err != nil {
		return 0, 0, fmt.Errorf("set HopLimit %d: %s", ttl, err)
	}

	return ping(conn, payload)
}

func ping(conn net.Conn, data []byte) (int, time.Duration, error) {

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Send it
	start := time.Now()

	wd, err := conn.Write(data);
	if err != nil {
		return 0, 0, err
	}

	rb := make([]byte, 1500)

	rd, err := conn.Read(rb)
	if err != nil {
		return 0, 0, err
	}
	ploss := len(data) - (wd+rd)

	duration := time.Since(start)

	return ploss, duration, nil
}

func createICMP(t icmp.Type) ([]byte, error) {
	m := icmp.Message{
	  Type: t,
	  Code: 0,
	  Body: &icmp.Echo{
	     ID: os.Getpid() & 0xffff, Seq: 1, //<< uint(seq), // TODO
	     Data: []byte("echo requests"),
	  },
	}
	return m.Marshal(nil)
}