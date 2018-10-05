package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	// "log"
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
	IntroducerIP     = "10.194.16.24"
	Port             = ":6666"
	DetectPeriod     = 500 * time.Millisecond
)

type Header struct {
	Type uint8
	Seq  uint16
	Zero uint8
}

type Update struct {
	UpdateID        uint64
	TTL             uint8
	UpdateType      uint8
	MemberTimeStamp uint64
	MemberIP        uint32
	MemberState     uint8
}

var init_timer *time.Timer
var PingAckTimeout map[uint16]*time.Timer
var FailureTimeout map[[2]uint64]*time.Timer
var CurrentMember *Member
var CurrentList *MemberList
var LocalIP string

var DuplicateUpdateCaches map[uint64]uint8
var TTLCaches *TtlCache

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
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	} else {
		return binary.BigEndian.Uint32(ip)
	}
}

// Convert uint32 to net.IP
func int2ip(binip uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, binip)
	return ip
}

// Helper function to print the err in process
func printError(err error) {
	if err != nil {
		fmt.Println("[ERROR]", err.Error())
	}
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

	// Seperate header and payload
	const HeaderLength = 4 // Header Length 4 bytes

	// Read header
	headerBinData := buffer[:HeaderLength]
	var header Header
	buf := bytes.NewReader(headerBinData)
	err = binary.Read(buf, binary.BigEndian, &header)
	printError(err)

	// Read payload
	payload := buffer[HeaderLength:n]

	// Resume detection

	if header.Type&Ping != 0 {

		// Check whether this ping carries Init Request
		if header.Type&MemInitRequest != 0 {
			// Handle Init Request
			fmt.Printf("RECEIVE INIT REQUEST FROM [%s]: %d\n", addr.IP.String(), header.Seq)
			initReply(addr.IP.String(), header.Seq, payload)

		} else if header.Type&MemUpdateSuspect != 0 {
			fmt.Printf("handle suspect update\n")
			handleSuspect(payload)
			// Get update entry from TTL Cache
			update, flag, err := getUpdate()
			// if no update there, do pure ping
			if err != nil {
				ack(addr.IP.String(), header.Seq)
			} else {
				// Send update as payload of ping
				ackWithPayload(addr.IP.String(), header.Seq, update, flag)
			}


		} else if header.Type&MemUpdateResume != 0 {
			fmt.Printf("handle resume update\n")
			handleResume(payload)
			// Get update entry from TTL Cache
			update, flag, err := getUpdate()
			// if no update there, do pure ping
			if err != nil {
				ack(addr.IP.String(), header.Seq)
			} else {
				// Send update as payload of ping
				ackWithPayload(addr.IP.String(), header.Seq, update, flag)
			}


		} else if header.Type&MemUpdateLeave != 0 {
			fmt.Printf("handle leave update\n")
			handleLeave(payload)
			// Get update entry from TTL Cache
			update, flag, err := getUpdate()
			// if no update there, do pure ping
			if err != nil {
				ack(addr.IP.String(), header.Seq)
			} else {
				// Send update as payload of ping
				ackWithPayload(addr.IP.String(), header.Seq, update, flag)
			}


		} else if header.Type&MemUpdateJoin != 0 {
			fmt.Printf("handle join update\n")
			handleJoin(payload)
			// Get update entry from TTL Cache
			update, flag, err := getUpdate()
			// if no update there, do pure ping
			if err != nil {
				ack(addr.IP.String(), header.Seq)
			} else {
				// Send update as payload of ping
				ackWithPayload(addr.IP.String(), header.Seq, update, flag)
			}


		} else {
			// Ping with no payload,
			// No handling payload needed
			// Check whether update sending needed
			// If no, simply reply with ack
			ack(addr.IP.String(), header.Seq)
		}

	} else if header.Type&Ack != 0 {

		// Receive Ack, stop ping timer
		timer, ok := PingAckTimeout[header.Seq-1]
		if ok {
			timer.Stop()
			fmt.Printf("RECEIVE ACK FROM [%s]: %d\n", addr.IP.String(), header.Seq)
			delete(PingAckTimeout, header.Seq-1)
		}

		// Read payload
		payload := buffer[HeaderLength:n]

		if header.Type&MemInitReply != 0 {
			// Ack carries Init Reply, stop init timer
			stop := init_timer.Stop()
			if stop {
				fmt.Printf("RECEIVE INIT REPLY FROM [%s]: %d\n", addr.IP.String(), header.Seq)
			}
			handleInitReply(payload)

		} else if header.Type&MemUpdateSuspect != 0 {
			fmt.Printf("handle suspect update\n")
			handleSuspect(payload)

		} else if header.Type&MemUpdateResume != 0 {
			fmt.Printf("handle resume update\n")
			handleResume(payload)

		} else if header.Type&MemUpdateLeave != 0 {
			fmt.Printf("handle leave update\n")
			handleLeave(payload)

		} else if header.Type&MemUpdateJoin != 0 {
			fmt.Printf("handle join update\n")
			handleJoin(payload)

		} else {
			fmt.Printf("receive pure ack\n")
		}
	}
}

// Check whether the update is duplicated
// If duplicated, return false, else, return true and start a timer
func isUpdateDuplicate(id uint64) bool {
	_, ok := DuplicateUpdateCaches[id]
	if ok {
		fmt.Printf("[INFO]: Receive duplicated update %d\n", id)
		return true
	} else {
		DuplicateUpdateCaches[id] = 1 // add to cache
		fmt.Printf("[INFO]: Add update %d to duplicated cache table \n", id)
		recent_update_timer := time.NewTimer(16 * time.Second) // set a delete timer
		go func() {
			<-recent_update_timer.C
			_, ok := DuplicateUpdateCaches[id]
			if ok {
				delete(DuplicateUpdateCaches, id) // delete from cache
				fmt.Printf("[INFO]: Delete update %d from duplicated cache table \n", id)
			}
		}()
		return false
	}
}

func getUpdate() ([]byte, uint8, error) {
	var binBuffer bytes.Buffer

	update, err := TTLCaches.Get()
	if err != nil {
		return binBuffer.Bytes(), 0, err
	}

	binary.Write(&binBuffer, binary.BigEndian, update)
	return binBuffer.Bytes(), update.UpdateType, nil
}

func handleSuspect(payload []byte) {
	buf := bytes.NewReader(payload)
	var update Update
	err := binary.Read(buf, binary.BigEndian, &update)
	printError(err)

	// Retrieve update ID
	updateID := update.UpdateID
	if !isUpdateDuplicate(updateID) {
		// If find someone sends suspect update which
		// suspect self, tell them I am alvie
		if CurrentMember.TimeStamp == update.MemberTimeStamp && CurrentMember.IP == update.MemberIP {
			addUpdate2Cache(CurrentMember, MemUpdateResume)
			return
		}

		// Receive new update, handle it
		CurrentList.Update(update.MemberTimeStamp, update.MemberIP,
			update.MemberState)
		TTLCaches.Set(&update)
		timer := time.NewTimer(time.Second)
		FailureTimeout[[2]uint64{update.MemberTimeStamp, uint64(update.MemberIP)}] = timer
		go func() {
			<-timer.C
			fmt.Printf("[Failure Detected][%s] %xTIMEOUT\n", int2ip(update.MemberIP).String(), update.MemberTimeStamp)
			err := CurrentList.Delete(update.MemberTimeStamp, update.MemberIP)
			printError(err)
			delete(FailureTimeout, [2]uint64{update.MemberTimeStamp, uint64(update.MemberIP)})
		}()

	}
}

func handleResume(payload []byte) {
	buf := bytes.NewReader(payload)
	var update Update
	err := binary.Read(buf, binary.BigEndian, &update)
	printError(err)

	// Retrieve update ID
	updateID := update.UpdateID
	if !isUpdateDuplicate(updateID) {
		// Receive new update, handle it
		timer, ok := FailureTimeout[[2]uint64{update.MemberTimeStamp, uint64(update.MemberIP)}]
		if ok {
			timer.Stop()
			delete(FailureTimeout, [2]uint64{update.MemberTimeStamp, uint64(update.MemberIP)})
		}
		CurrentList.Update(update.MemberTimeStamp, update.MemberIP, update.MemberState)
		TTLCaches.Set(&update)
	}
}

func handleLeave(payload []byte) {
	buf := bytes.NewReader(payload)
	var update Update
	err := binary.Read(buf, binary.BigEndian, &update)
	printError(err)

	// Retrieve update ID
	updateID := update.UpdateID
	if !isUpdateDuplicate(updateID) {
		// Receive new update, handle it
		CurrentList.Delete(update.MemberTimeStamp, update.MemberIP)
		TTLCaches.Set(&update)
	}
}

func handleJoin(payload []byte) {
	buf := bytes.NewReader(payload)
	var update Update
	err := binary.Read(buf, binary.BigEndian, &update)
	printError(err)

	// Retrieve update ID
	updateID := update.UpdateID
	if !isUpdateDuplicate(updateID) {
		// Receive new update, handle it
		CurrentList.Insert(&Member{update.MemberTimeStamp, update.MemberIP,
			update.MemberState})
		TTLCaches.Set(&update)
	}
}

// Generate a new update and set it in TTL Cache
func addUpdate2Cache(member *Member, updateType uint8) {
	key := TTLCaches.RandGen.Uint64()
	update := Update{key, 3, updateType, member.TimeStamp, member.IP, member.State}
	TTLCaches.Set(&update)
}

// Handle the full membership list(InitReply) received from introducer
func handleInitReply(payload []byte) {
	num := len(payload) / 13 // 13 bytes per member
	buf := bytes.NewReader(payload)
	for idx := 0; idx < num; idx++ {
		var member Member
		err := binary.Read(buf, binary.BigEndian, &member)
		printError(err)
		// Insert existing member to the new member's list
		CurrentList.Insert(&member)
	}
}

// Introducer replies new node join init request and
// send the new node join updates to others in membership
func initReply(addr string, seq uint16, payload []byte) {
	// Read and insert new member to the memberlist
	var member Member
	buf := bytes.NewReader(payload)
	err := binary.Read(buf, binary.BigEndian, &member)
	printError(err)
	// Update state of the new member
	// ...
	CurrentList.Insert(&member)
	addUpdate2Cache(&member, MemUpdateJoin)

	// Put the entire memberlist to the Init Reply's payload
	var memBuffer bytes.Buffer // Temp buf to store member's binary value
	var binBuffer bytes.Buffer

	for i := 0; i < CurrentList.Size(); i += 1 {
		member_, _ := CurrentList.RetrieveByIdx(i)

		binary.Write(&memBuffer, binary.BigEndian, member_)
		binBuffer.Write(memBuffer.Bytes())
		memBuffer.Reset() // Clear buffer
	}

	// DEBUG PRINTLIST
	fmt.Printf("len of payload before send reply: %d\n", len(binBuffer.Bytes()))
	CurrentList.PrintMemberList()

	// Send pigggback Init Reply
	ackWithPayload(addr, seq, binBuffer.Bytes(), MemInitReply)
}

func initRequest(member *Member) {
	// Construct Init Request payload
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, member)

	// Send piggyback Init Request
	pingWithPayload(&Member{0, ip2int(net.ParseIP(IntroducerIP)), 0}, binBuffer.Bytes(), MemInitRequest)

	// Start Init timer, if expires, exit process
	init_timer = time.NewTimer(2 * time.Second)
	go func() {
		<-init_timer.C
		fmt.Printf("INIT %s TIMEOUT, PROCESS EXIT.\n", IntroducerIP)
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

func pingWithPayload(member *Member, payload []byte, flag uint8) {
	// Source for genearting random number
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)
	seq := randGen.Intn(0x01<<15 - 2)
	addr := int2ip(member.IP).String() + Port

	packet := Header{Ping | flag, uint16(seq), 0}
	var binBuffer bytes.Buffer
	binary.Write(&binBuffer, binary.BigEndian, packet)

	if payload != nil {
		binBuffer.Write(payload) // Append payload
		udpSend(addr, binBuffer.Bytes())
	} else {
		udpSend(addr, binBuffer.Bytes())
	}
	fmt.Printf("PING [%s]: %d\n", addr, seq)

	timer := time.NewTimer(time.Second)
	PingAckTimeout[uint16(seq)] = timer
	go func() {
		<-timer.C
		fmt.Printf("PING [%s]: %d TIMEOUT\n", addr, seq)
		err := CurrentList.Update(member.TimeStamp, member.IP, StateSuspect)
		if err == nil {
			addUpdate2Cache(member, MemUpdateSuspect)
		}
		delete(PingAckTimeout, uint16(seq))
	}()
}

func ping(member *Member) {
	pingWithPayload(member, nil, 0x00)
}

// Start the membership service and join in the group
func startService() bool {

	// Create a new log file or append to exit file
	// file, err := os.OpenFile(LocalIP+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	// if err != nil {
	// 	fmt.Println("[Error] Open or Create log file failed")
	// 	return false
	// }
	// logPrefix := "[" + LocalIP + "]: "
	// fmt = log.New(file, logPrefix, log.Ldate|log.Lmicroseconds|log.Lshortfile)

	// Create self entry
	LocalIP = getLocalIP().String()
	timestamp := time.Now().UnixNano()
	state := StateAlive
	CurrentMember = &Member{uint64(timestamp), ip2int(getLocalIP()), uint8(state)}

	// Create member list
	CurrentList = NewMemberList(10)

	// Make necessary tables
	PingAckTimeout = make(map[uint16]*time.Timer)
	FailureTimeout = make(map[[2]uint64]*time.Timer)
	DuplicateUpdateCaches = make(map[uint64]uint8)
	TTLCaches = NewTtlCache()

	if LocalIP == IntroducerIP {
		CurrentMember.State |= (StateIntro | StateMonit)
		CurrentList.Insert(CurrentMember)
	} else {
		// New member, send Init Request to the introducer
		initRequest(CurrentMember)
	}

	return true
}

// Main func
func main() {

	// Init
	if startService() == true {
		fmt.Printf("START SERVICE\n")
	}

	// Start daemon
	udpDaemon()

	for {
		// Shuffle membership list and get a member IP
		if CurrentList.Size() > 0 {
			member := CurrentList.Shuffle()
			// Do not pick itself as the ping target
			if member.TimeStamp == CurrentMember.TimeStamp && member.IP == CurrentMember.IP {
				time.Sleep(DetectPeriod)
				continue
			}
			// Get update entry from TTL Cache
			update, flag, err := getUpdate()
			// if no update there, do pure ping
			if err != nil {
				ping(member)
			} else {
				// Send update as payload of ping
				pingWithPayload(member, update, flag)
			}
			time.Sleep(DetectPeriod)
		}
	}

}
