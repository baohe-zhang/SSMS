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
	MemInitRequest   = 0x01 << 2
	MemInitReply     = 0x01 << 3
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

var init_timer time.Timer
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

// Convert net.IP to uint32
func ip2int(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}
// Convert uint32 to net.IP 
func int2ip(binip uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, binip)
	return ip
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
	CurrentEntry = &Member{uint64(timestamp), ip2int(getLocalIP()), uint8(state)}
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
	_, addr, err := connect.ReadFromUDP(buffer)
	printError(err)

	// Seperate header and payload
	const HeaderLength = 4  // Header Length 4 bytes

	// Read header
	headerBinData := buffer[:HeaderLength]
	var header Header
	buf := bytes.NewReader(headerBinData)
	err = binary.Read(buf, binary.BigEndian, &header)
	printError(err)

	// Read payload
	payload := buffer[HeaderLength:] 


	if header.Type&Ping != 0 {
		// Receive Ping, check whether this pring carrie flags

		// Check whether this ping carries Init Request
		if header.Type&MemInitRequest != 0 {
			// Handle Init Request
			initReply(addr.IP.String(), header.Seq, payload)

		} else if header.Type&MemUpdateSuspect != 0 {
			fmt.Printf("handle suspect\n")
		} else if header.Type&MemUpdateResume != 0 {
			fmt.Printf("handle resume\n")
		} else if header.Type&MemUpdateLeave != 0 {
			fmt.Printf("handle leave\n")
		} else if header.Type&MemUpdateJoin != 0 {
			fmt.Printf("handle join\n")
		} else {
			// Ping with no payload, 
			// No handling payload needed
			// Check whether update sending needed
			// If no, simply reply with ack
			ack(addr.IP.String(), header.Seq)
		}


	} else if header.Type&Ack != 0 {

		// Receive Ack, stop ping timer
		stop := PingAckTimeout[header.Seq-1].Stop()
		if stop {
			Logger.Printf("RECEIVE ACK [%s]: %d\n", addr.IP.String(), header.Seq)
			delete(PingAckTimeout, header.Seq-1)
		}

		// Ack carries Init Reply, stop init timer
		if header.Type&MemInitReply != 0 {
			stop := init_timer.Stop()
			if stop {
				Logger.Printf("RECEIVE INIT REPLY FROM [%s]: %d\n", addr.IP.String(), header.Seq)
			}

			// Retrive data from Init Reply and store them into the memberlist

		} else if header.Type&MemUpdateSuspect != 0 {
			fmt.Printf("handle suspect\n")
		} else if header.Type&MemUpdateResume != 0 {
			fmt.Printf("handle resume\n")
		} else if header.Type&MemUpdateLeave != 0 {
			fmt.Printf("handle leave\n")
		} else if header.Type&MemUpdateJoin != 0 {
			fmt.Printf("handle join\n")
		} else {
			fmt.Printf("receive pure ack\n")
		} 
	}
}

func initReply(addr string, seq uint16, payload []byte) {
	// Read and insert new member to the memberlist
	var member Member
	buf := bytes.NewReader(payload)
	err := binary.Read(buf, binary.BigEndian, &member)
	printError(err)
	CurrentList.Insert(&member)

	// Put the entire memberlist to the Init Reply's payload
	var memBuffer bytes.Buffer  // Temp buf to store member's binary value
	var binBuffer bytes.Buffer
	for i := 0; i < CurrentList.Size(); i++ {
		member = CurrentList.RetrieveByIdx(i)
		binary.Write(&memBuffer, binary.BigEndian, member)
		binBuffer.Write(memBuffer.Bytes())
	}

	// Send pigggback Init Reply
	ackWithPayload(addr, seq, binBuffer.Bytes(), MemInitReply)
}

func initRequest(member *Member) {
	// Construct Init Request payload
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, member)

	// Send piggyback Init Request
	pingWithPayload(IntroducerIP, binBuffer.Bytes(), MemInitRequest)

	// Start Init timer, if expires, exit process
	init_timer := time.NewTimer(5 * time.Second)
	go func() {
		<-init_timer.C
		Logger.Printf("INIT %s TIMEOUT, PROCESS EXIT.", IntroducerIP)
		os.Exit(1)
	}()
}

func ackWithPayload(addr string, seq uint16, payload []byte, flag uint8) {
	packet := Header{Ack | flag, seq + 1, 0}
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, packet)

	if payload != nil {
		binBuffer.Write(payload) // Append payload
		udpSend(addr+Port, binBuffer.Bytes())
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
		binBuffer.Write(payload) // Append payload
		udpSend(addr, binBuffer.Bytes())
	} else {
		udpSend(addr, binBuffer.Bytes())
	}
	Logger.Printf("PING [%s]: %d\n", addr, seq)

	timer := time.NewTimer(time.Second)
	PingAckTimeout[uint16(seq)] = timer
	go func() {
		<-PingAckTimeout[uint16(seq)].C
		Logger.Printf("PING [%s]: %d TIMEOUT\n", addr, seq)
		delete(PingAckTimeout, uint16(seq))
	}()
}

func ping(addr string) {
	pingWithPayload(addr, nil, 0x00)
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
