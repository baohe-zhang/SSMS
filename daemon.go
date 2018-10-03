package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
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
	MemUpdateResume  = 0x01 << 5
	MemUpdateLeave   = 0x01 << 6
	MemUpdateJoin    = 0x01 << 7
	StateAlive       = 0x01
	StateSuspect     = 0x01 << 1
	StateMonit       = 0x01 << 2
	StateIntro       = 0x01 << 3
	IntroducerIP     = ""
	Port             = ":6666"
	DetectPeriod     = 500 * time.Millisecond
)

type Header struct {
	Type uint8
	Seq  uint16
	Zero uint8
}

var PingAckTimeout map[uint16]*time.Timer
var CurrentEntry *Member
var CurrentList *MemberList
var Logger *log.Logger

var LocalIP string

// A trick to simply get local IP address
func getLocalIP() net.IP {
	dial, err := net.Dial("udp", "8.8.8.8:80")
	printError(err)
	localAddr := dial.LocalAddr().(*net.UDPAddr)
	dial.Close()

	return localAddr.IP
}

// Start the membership service and join in the group
func startService() bool {
	state := StateAlive

	// Create a new log file or append to exit file
	file, err := os.OpenFile(LocalIP+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		fmt.Println("[Error] Open or Create log file failed")
		return false
	}
	logPrefix := "[" + LocalIP + "]: "
	Logger = log.New(file, logPrefix, log.Ldate|log.Lmicroseconds|log.Lshortfile)

	LocalIP = getLocalIP().String()
	timestamp := time.Now().UnixNano()
	CurrentEntry = Member{uint64(timestamp), LocalIP, state}
	CurrentList = NewMemberList(10)
	CurrentList.Insert(CurrentEntry)
	if LocalIP == IntroducerIP {
		CurrentEntry.State |= (StateIntro | StateMonit)
	} else {

	}

	return true
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
	var header Header
	buf := bytes.NewReader(data)
	err = binary.Read(buf, binary.BigEndian, &header)
	printError(err)

	if header.Type&Ping != 0 {
		ack(addr.IP.String(), header.Seq)
	} else if header.Type&Ack != 0 {
		stop := PingAckTimeout[header.Seq-1].Stop()
		if stop {
			Logger.Printf("ACK [%s]: %d\n", addr.IP.String(), header.Seq)
			delete(PingAckTimeout, header.Seq-1)
		}
	}

}

func join(member *Member) {
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, member)
}

func ackWithPayload(addr string, seq uint16, payload []byte, flag uint8) {
	packet := Header{Ack | flag, seq + 1, 0}
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, packet)

	if payload != nil {
		udpSend(addr+Port, binBuffer.Bytes()+payload)
	} else {
		udpSend(addr+Port, binBuffer.Bytes())
	}
}

func ack(addr string, seq uint16) {
	ackWithPayload(addr, seq, nil, 0x00)
}

func pingWithPayload(addr string, payload []byte, flag uint8) {
	// Source for genearting random number
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)
	seq := randGen.Intn(0x01<<15 - 2)

	packet := Header{Ping | flag, uint16(seq), 0}
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, packet)

	if payload != nil {
		udpSend(addr, binBuffer.Bytes()+payload)
	} else {
		udpSend(addr, binBuffer.Bytes())
	}
	Logger.Printf("Ping [%s]: %d\n", addr, seq)

	timer := time.NewTimer(time.Second)
	PingAckTimeout[uint16(seq)] = timer
	go func() {
		<-PingAckTimeout[uint16(seq)].C
		Logger.Printf("Ping [%s]: %d timeout, no response\n", addr, seq)
		delete(PingAckTimeout, uint16(seq))
	}()
}

func ping(addr string) {
	pintWithPayload(addr, nil, 0x00)
}

// Main func
func main() {
	PingAckTimeout = make(map[uint16]*time.Timer)
	udpDaemon()
	for {
		ping("10.193.185.82" + Port)
		time.Sleep(DetectPeriod)
	}
}

// Helper function to print the err in process
func printError(err error) {
	if err != nil {
		Logger.Println("[ERROR]", err.Error())
	}
}
