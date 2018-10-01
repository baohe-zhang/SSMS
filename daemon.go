package main

import (
	"fmt"
	"net"
	"os"
)

// UDP Daemon loop task
func udpDaemon() {
	serverAddr := ":3666"
	udpAddr, err := net.ResolveUDPAddr("udp4", serverAddr)
	printError(err)
	// Listen the request from client
	listen, err := net.ListenUDP("udp", udpAddr)
	printError(err)
	for {
		go udpDaemonHandle(listen)
	}
}

func udpDaemonHandle(connect *net.UDPConn) {
	// Making a buffer to accept the grep command content from client
	buffer := make([]byte, 1024)
	n, addr, err := connect.ReadFromUDP(buffer)
	printError(err)

	data := string(buffer[:n])
	fmt.Println(data, addr.String())
}

func main() {
	go udpDaemon()
}

// Helper function to print the err in process
func printError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[ERROR]", err.Error())
	}
}
