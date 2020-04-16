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
	"syscall"
	"bufio"
	"errors"
)
const ttl = 52

var PacketsRecv int
var PacketsSent int

type Address string

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Please enter an ip or an address: ")
	scanner.Scan()
	netaddr := scanner.Text()

	addr := Address(netaddr)

	for {
		ploss, dur, err := addr.Pinger("ipv4")
		if err != nil {
			log.Printf("host: %s, Loss: %f, error: %s\n", addr, ploss, err)
			return
		}
		log.Printf("host: %s, Loss: %f, RTT: %s\n", addr, ploss, dur)
		time.Sleep(1 * time.Second)
	}
}

func (addr Address) Pinger(ipv string) (float64, time.Duration, error) {
	switch ipv{
		case "ipv4":
			payloadV4, _ := createICMP(ipv4.ICMPTypeEcho)
			return PingIpv4(addr, ttl, payloadV4)
		case "ipv6":
			payloadV6, _ := createICMP(ipv6.ICMPTypeEchoRequest)
			return PingIpv6(addr, ttl, payloadV6)
		default:
			return 0, 0, errors.New("No ip version specified")
	}
}

func PingIpv4(address Address, ttl int, payload []byte)(float64, time.Duration, error){
	netaddr, err := net.ResolveIPAddr("ip4", string(address))

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

func PingIpv6(address Address, ttl int, payload []byte) (float64, time.Duration, error){
	netaddr, err := net.ResolveIPAddr("ip", string(address))

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

func ping(conn net.Conn, data []byte) (float64, time.Duration, error) {

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Send it
	start := time.Now()
	for {
		if _, err := conn.Write(data); err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
				}
			}
		}
		PacketsSent++
		break
	}

	rb := make([]byte, 1500)

	_, err := conn.Read(rb)
	if err != nil {
		return 0, 0, err
	}

	ploss := float64(PacketsSent-PacketsRecv) / float64(PacketsSent) * 100

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
	PacketsRecv++
	return m.Marshal(nil)
}