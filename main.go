package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func main() { 
	port := flag.Int("port", 1080, "port to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *port, err)
	}
	defer listener.Close()

	log.Printf("SOCKS5 proxy listening on :%d", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}


// ASSIGNMENT STARTS HERE: 

// tldr: 
// check 0x05 for SOCKS5 -> check 0x00 OR 0x02 for username\pass
// if 0x02 in methods -> jump into authenticateUserPass()
// if 0x00 or after authenticateUser -> jump into handleConnection()
// after handleConnection -> jump into relay()
func negotiateAuth(conn net.Conn) error { 
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	ver, nmethods := header[0], header[1]
	if ver != 0x05 {
		return errors.New("not SOCKS5")
	}
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}

	authNotNeeded := false
	authNeeded := false

	for i := 0; i < nmethods; i++ {
		if methods[i] == 0x02 {
			authNeeded = true // found 0x02 so the connection supports the username and password
		}
		if methods[i] == 0x00 {
			authNotNeeded = true // found 0x00 so the connection supports no username and password
		}
	}
	// building the response package:
	res := make([]byte, 2)
	res[0] = 0x05

	if (authNeeded==true) {
		res[1] = 0x02
	} else if (authNotNeeded==true) {
		res[1] = 0x00
	} else {
		res[1] = 0xff
	}
	// since i think 0x02 takes higher priority, 
	// if there is both 0x02 AND 0x00 we will stay with 0x02
	
	_, err := conn.Write(res)
	if err != nil {
   	return err 
	}
	if(res[1]==0xff){
		conn.Close()
		return errors.New("no acceptable methods")
	}
	if(res[1]==0x02){
		authenticateUserPass(conn)
	}

	return nil
}

func authenticateUserPass(conn net.Conn){
	
	header := make([]byte, 2) // taking in version and userLen
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	ver, usernameLen := header[0], header[1]
	if ver != 0x01 {
		return errors.New("not username password method")
	}
	username := make([]byte, usernameLen)
	if _, err := io.ReadFull(conn, username); err != nil {
		return err
	}
	
	passwordLen := make([]byte, 1)
	if _, err := io.ReadFull(conn, passwordLen); err != nil {
		return err
	}
	password := make([]byte, passwordLen[0])
	if _, err := io.ReadFull(conn, password); err != nil {
		return err
	}



}


func handleConnection(conn net.Conn) {
	defer conn.Close()

	negotiateAuth(conn); // immedietly go check SOCKS5 and the username password thingy


	// TODO: Implement SOCKS5 protocol
	// 1. Read client greeting and negotiate authentication method
	// 2. Perform authentication if required (when PROXY_USER env var is set)
	// 3. Read CONNECT request
	// 4. Connect to target server
	// 5. Send success/error reply
	// 6. Relay data between client and target
}
