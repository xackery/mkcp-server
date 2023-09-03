package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

func main() {
	isClient := false

	if len(os.Args) > 1 {
		if os.Args[1] == "client" {
			fmt.Println("Running as client")
			isClient = true
		}
	}
	if !isClient {
		fmt.Println("Running as server")
	}
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	if !isClient {
		listener, err := kcp.ListenWithOptions("127.0.0.1:12345", block, 10, 3)
		if err != nil {
			log.Fatal(err)
		}
		for {
			s, err := listener.AcceptKCP()
			if err != nil {
				log.Fatal(err)
			}
			go handleEcho(s)
		}
	}
	err := client()
	if err != nil {
		log.Fatal(err)
	}
}

// handleEcho send back everything it received
func handleEcho(conn *kcp.UDPSession) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("recv:", string(buf))

		buf = append(buf, []byte("echo: ")...)
		_, err = conn.Write(buf[:n+6])
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func client() error {
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	// wait for server to become ready
	time.Sleep(time.Second)

	// dial to the echo server
	sess, err := kcp.DialWithOptions("127.0.0.1:12345", block, 10, 3)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	go func() {
		for {
			buf := make([]byte, 4096)
			n, err := sess.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("recv:", string(buf[:n]))
			time.Sleep(time.Second)
		}
	}()
	for {
		data := time.Now().String()
		log.Println("sent:", data)
		_, err = sess.Write([]byte(data))
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		time.Sleep(time.Second)
	}
}
