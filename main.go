package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"encoding/binary"
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

	// TODO: Implement SOCKS5 protocol
	// 1. V- Read client greeting and negotiate authentication method
	// 2. V- Perform authentication if required (when PROXY_USER env var is set)
	// 3. Read CONNECT request
	// 4. Connect to target server
	// 5. Send success/error reply
	// 6. V- Relay data between client and target

// returns error or nil if negotiation went well
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
		result := authenticateUserPass(conn)
		if(result!=nil) { // function returns nil or error
			return result
		}
	}
	return nil
}

// returns error or nil if authentication went well
func authenticateUserPass(conn net.Conn) error {
	header := make([]byte, 2) // taking in version and userLen
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	ver, usernameLen := header[0], header[1]
	if ver != 0x01 {
		conn.Close()
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

	res := make([]byte, 2)
	res[0] = 0x01


	if (expectedUser == string(username) && expectedPass == string(password)){
		// correct auth info! 
		res[1] = 0x00

	} else{
		res[1] = 0x01
		// error
	}
	
	_, err := conn.Write(res)
	if err != nil {
   	return err 
	}

	if(res[1]== 0x00){
		return nil
	} else {
		conn.Close()
		return errors.New("wrong authentication info")
	}


}

// returns error or nil if connectionwent wel
func handleConnection(conn net.Conn) error {
	defer conn.Close()
	err := negotiateAuth(conn); // immedietly go check SOCKS5 and the username password thingy
	if(err!=nil){ // returned with issue
		// not closing connection since it's already bene closed earlier
		return err
	}
	
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if(header[0]!=0x05 || header[1] != 0x01 || header[2] != 0x00){
		// at least ONE of the first bytes isnt good
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		conn.Close()
		return errors.New("issue in the connection attempt")
	}
	adress := ""
	if(header[3] == 0x01){
		ipArr := make([]byte, 4)
		if _, err := io.ReadFull(conn, ipArr); err != nil {
			return err
		}
		adress = net.IP(ipArr).String()
	} else if(header[3] == 0x03) {
		domainLength := make([]byte, 1)
		if _, err := io.ReadFull(conn, domainLength); err != nil {
			return err
		}
		domainArr := make([]byte, int(domainLength[0]))
		if _, err := io.ReadFull(conn, domainArr); err != nil {
			return err
		}
		adress = string(domainArr)
	} else {
		// not IPv4 and not domain: error
		conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		conn.Close()
		return errors.New("issue in ATYP (not domain and not IPv4)")
	}
	port := make([]byte, 2)
	if _, err := io.ReadFull(conn, port); err != nil {
		return err
	}
	portNum := binary.BigEndian.Uint16(port)
	finalAdress := fmt.Sprintf("%s:%d", adress, portNum)
	// if everything so far is okay: dail and then reply about the connection
	target, err := net.Dial("tcp", finalAdress)
   if err != nil { // if failed: say it, close, the entire thing, and return error
   	conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		conn.Close()
      return err
   }
	defer targetConn.Close() // i hate defer! i get it but it just doesnt click for me just yet

	// send "green light" and nowe we can relay
	err:= conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if(err!=nil){
		conn.Close()
		return err
	}
	
	relay(conn, targetConn)

    return nil
}


func relay(client net.Conn, target net.Conn) {
	//adapting the concurrent relay function to the full signature to prevent TCP Deadlock 
	var wg sync.WaitGroup
	wg.Add(2)
	//client -> target
	go func() {
		defer wg.Done()
		_, _ = io.Copy(target, client)
		if tcpConn, ok := target.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite() 
		}
	}()
	//target -> client
	go func() {
		defer wg.Done()
		_, _ = io.Copy(client, target)
		if tcpConn, ok := client.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite() 
		}
	}()
	wg.Wait()
} 


