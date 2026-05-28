package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
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

	for i := 0; i < int(nmethods); i++ {
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

	expectedUser := os.Getenv("PROXY_USER") // asking the OS if the admin provided a password
	if (expectedUser != "" && authNeeded==true){ // the client gave 0x02 (username and password) AND admin had set auth info
		res[1] = 0x02
	} else if (expectedUser == "" && authNotNeeded==true) { // supports no auth AND admin gave no auth info
		res[1] = 0x00
	} else { // else error
		res[1] = 0xff 
	}

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
	// if its 0x00: continue somewhere, donno yet
	return nil
}

func authenticateUserPass(conn net.Conn) error {
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

	expectedUser := os.Getenv("PROXY_USER")
	expectedPass := os.Getenv("PROXY_PASS")
	if (expectedUser == string(username) && expectedPass == string(password)){
		// correct auth info! 
	} else{
		// close connection and send error
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
