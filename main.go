package main

import (
	"bufio"
	"time"
	"os"
	"log"
	"fmt"
	"net"
	"syscall"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	_"golang.org/x/net/ipv6"
	"errors"
)

const (
	ProtocolICMP = 1
	ProtocolIPv6ICMP = 58
)

// Default to listen on all IPv4 interfaces
var ListenAddr = "0.0.0.0"

var PacketsSent int
var	PacketsRecv int

type Address struct{
	name string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Please enter an ip or an address: ")

	scanner.Scan()

	text := scanner.Text()

	addr := Address{text}

	for{

		ploss, dur, err := addr.Ping("ipv4")

		if err != nil {
			log.Printf("host: %s, Loss: %f, error: %s\n", addr.name, ploss, err)
			return
		}
		log.Printf("host: %s, Loss: %f, RTT: %s\n", addr.name, ploss, dur)

		time.Sleep(1 * time.Second)
	}
}

func (addr Address) Ping(ipv string) (float64, time.Duration, error) {
	switch ipv{
		case "ipv4":
			return PingIp4(addr)
		// case "ipv6":
		// 	return PingIp6(addr)
		default:
			return 0, 0, errors.New("No ip version specified")
	}
}

func PingIp4(addr Address)(float64, time.Duration, error){
	// Start listening for icmp replies
	c, err := icmp.ListenPacket("ip4:icmp", ListenAddr)
	if err != nil {
			return 0, 0, err
	}

	defer c.Close()

	// Resolve any DNS (if used) and get the real IP of the target
	hostip, err := net.ResolveIPAddr("ip4", addr.name)

	if err != nil {
			panic(err)
			return 0, 0, err
	}

	// create ICMP message
	m := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
					ID: os.Getpid() & 0xffff, Seq: 1, //<< uint(seq), // TODO
					Data: []byte("echo requests"),
			},
	}
	b, err := m.Marshal(nil)
	if err != nil {
			return 0, 0, err
	}
	PacketsRecv++

	// Send it
	start := time.Now()
	for {
		if _, err := c.WriteTo(b, hostip); err != nil{
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
					//return 0, 0, err
				}
			}
		}
		PacketsSent++
		break
	}
	//calculate Packet loss
	ploss := float64(PacketsSent-PacketsRecv) / float64(PacketsSent) * 100

	// Wait for a reply
	reply := make([]byte, 1500)

	err = c.SetReadDeadline(time.Now().Add(10 * time.Second))

	if err != nil {
			return ploss, 0, err
	}
	n, peer, err := c.ReadFrom(reply)
	if err != nil {
			return ploss, 0, err
	}
	duration := time.Since(start)

	// done here
	rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
	if err != nil {
			return ploss, 0, err
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
			return ploss, duration, nil
	default:
			return ploss, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}

// func PingIp6(addr Address)(float32, time.Duration, error){

// 	c, err := icmp.ListenPacket("udp6", ListenAddr)

// 	if err != nil {
// 		return 0, 0, err
// 	}
// 	defer c.Close()

// 	// Resolve any DNS (if used) and get the real IP of the target
// 	hostip, err := net.ResolveIPAddr("ip", addr.name)



// 	wm := icmp.Message{
// 		Type: ipv6.ICMPTypeEchoRequest, Code: 0,
// 		Body: &icmp.Echo{
// 			ID: os.Getpid() & 0xffff, Seq: 1,
// 			Data: []byte("echo requests"),
// 		},
// 	}

// 	wb, err := wm.Marshal(nil)

// 	if err != nil {
// 		return 0, 0, err
// 	}

// 	// Send it
// 	start := time.Now()

// 	n, err := c.WriteTo(wb, hostip)

// 	if err != nil {
// 		return 0, 0, err
// 	}

// 	//calculate Packet loss
// 	ploss := float32(len(wb)) - float32(n)

// 	rb := make([]byte, 1500)

// 	err = c.SetReadDeadline(time.Now().Add(10 * time.Second))
// 	if err != nil {
// 			return ploss, 0, err
// 	}

// 	n, peer, err := c.ReadFrom(rb)
// 	if err != nil {
// 		return 0, 0, err
// 	}

// 	duration := time.Since(start)

// 	rm, err := icmp.ParseMessage(ProtocolIPv6ICMP, rb[:n])

// 	if err != nil {
// 		return 0, 0, err
// 	}

// 	switch rm.Type {
// 		case ipv6.ICMPTypeEchoReply:
// 			return ploss, duration, nil
// 		default:
// 			return ploss, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
// 	}
// }