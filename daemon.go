package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	Ping             = 0x01
	Ack              = 0x01 << 1
	MemUpdate        = 0x01 << 2
	MemInit          = 0x01 << 3
	MemUpdateSuspect = 0x01 << 4
	MemUpdateAlive   = 0x01 << 5
	MemUpdateLeave   = 0x01 << 6
	MemUpdateJoin    = 0x01 << 7
	IntroducerIP     = " "
	Port             = ":6666"
	DetectPeriod     = 500 * time.Millisecond
)

type SSMSHeader struct {
	SType uint8
	SSeq  uint16
	SZero uint8
}

//type SSMSEntry struct {

// A trick to simply get local IP address
func getLocalIP() net.IP {
	dial, err := net.Dial("udp", "8.8.8.8:80")
	printError(err)
	localAddr := dial.LocalAddr().(*net.UDPAddr)
	dial.Close()

	return localAddr.IP
}

// UDP send
func udpSend(addr string, packet []byte) {
	conn, err := net.Dial("udp", addr)
	printError(err)
	defer conn.Close()

	conn.Write(packet)
}

// UDP Daemon loop task
func udpDaemon() {
	serverAddr := Port
	udpAddr, err := net.ResolveUDPAddr("udp4", serverAddr)
	printError(err)
	// Listen the request from client
	listen, err := net.ListenUDP("udp", udpAddr)
	printError(err)

	go func() {
		for {
			udpDaemonHandle(listen)
		}
	}()
}

func udpDaemonHandle(connect *net.UDPConn) {
	// Making a buffer to accept the grep command content from client
	buffer := make([]byte, 1024)
	n, addr, err := connect.ReadFromUDP(buffer)
	printError(err)

	data := buffer[:n]
	var header SSMSHeader
	buf := bytes.NewReader(data)
	err = binary.Read(buf, binary.BigEndian, &header)
	printError(err)

	if header.SType&Ping == 0x01 {
		ack(addr.IP.String(), header.SSeq)
	} else if header.SType&Ack == 0x01 {
		fmt.Println(addr.IP)
		fmt.Printf("ACK: %d", header.SSeq)
	}

}

func ack(addr string, seq uint16) {
	packet := SSMSHeader{Ack, seq + 1, 0}
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, packet)

	udpSend(addr+Port, binBuffer.Bytes())
}

func ping(addr string) {
	// Source for genearting random number
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)
	seq := randGen.Intn(0x01<<15 - 2)

	packet := SSMSHeader{Ping, uint16(seq), 0}
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, packet)

	udpSend(addr, binBuffer.Bytes())
}

func main() {
	fmt.Println(getLocalIP())
	udpDaemon()
	for {
		ping("10.193.185.82" + Port)
		time.Sleep(time.Second)
	}
}

// Helper function to print the err in process
func printError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[ERROR]", err.Error())
	}
}
