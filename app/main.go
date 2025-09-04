package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		resp, err := readRESP(reader)
		if err != nil {
			fmt.Println("Error reading RESP:", err.Error())
			return
		}
		handleRESP(conn, resp)
	}
}

// *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
func readRESP(reader *bufio.Reader) ([]string, error) {
	header, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(header, "*") {
		fmt.Println("Error reading header:", err.Error())
		return nil, err
	}
	var count int
	fmt.Sscanf(header, "*%d", &count)
	args := make([]string, count)
	for i := 0; i < count; i++ {
		commandLength, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command length:", err.Error())
			return nil, err
		}
		var length int
		fmt.Sscanf(commandLength, "$%d", &length)
		data := make([]byte, length+2)
		_, err = reader.Read(data)
		fmt.Println("Read data:", string(data), "length:", length)
		if err != nil {
			fmt.Println("Error reading command argument:", err.Error())
			return nil, err
		}
		args[i] = string(data[:length])
	}
	return args, nil
}

func handleRESP(conn net.Conn, resp []string) {
	if len(resp) == 0 {
		return
	}
	command := strings.ToUpper(resp[0])
	switch command {
	case "ECHO":
		if len(resp) != 2 {
			conn.Write([]byte("-ERR wrong number of arguments for 'echo' command\r\n"))
			return
		}
		response := fmt.Sprintf("$%d\r\n%s\r\n", len(resp[1]), resp[1])
		conn.Write([]byte(response))
	case "SET":
		if len(resp) <= 2 {
			conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
			return
		}
		key := resp[1]
		value := resp[2]
		storeMutex.Lock()
		store[key] = value
		storeMutex.Unlock()
		conn.Write([]byte("+OK\r\n"))
	case "GET":
		if len(resp) != 2 {
			conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
			return
		}
		key := resp[1]
		storeMutex.RLock()
		value, exists := store[key]
		storeMutex.RUnlock()
		if !exists {
			conn.Write([]byte("$-1\r\n"))
		} else {
			response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
			conn.Write([]byte(response))
		}
	case "PING":
		if len(resp) == 1 {
			conn.Write([]byte("+PONG\r\n"))
		}
	default:
		conn.Write([]byte("-ERR unknown command\r\n"))
	}
}
