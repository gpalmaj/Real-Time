package faulttolerance

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	mutexAddr         = "127.0.0.1:35000"
	heartbeatAddr     = "127.0.0.1:35001"
	heartbeatInterval = 100 * time.Millisecond
	heartbeatTimeout  = 2 * time.Second
)

// Starts a process pair. Uses a port as a mutex to determine which process is main and which is copy
// If the process can bind to the mutex port, it becomes main, starts the heartbeat and creates a copy process
// If the process cannot bind to the mutex port, it becomes a copy and waits for the heartbeat to stop before trying to become main again
func RunProcessPair(id int) {
	for {
		mutexListener, err := net.Listen("tcp", mutexAddr) // Try to take mutex port
		if err == nil {
			// This is main
			go runHeartbeat()
			create_copy(id)

			// Keep mutex alive while we are main
			_ = mutexListener
			return
		}

		// Port is busy, meaning there is already a main running, this is copy
		waitForHeartbeatLoss()
		fmt.Printf("Copy takes over \n")
	}
}

// Listens for heartbeat on port 35001. If there is no heartbeat for two seconds it returns, allowing the copy to try to become main
func waitForHeartbeatLoss() {
	listen, err := net.Listen("tcp", heartbeatAddr)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return
	}
	defer listen.Close()

	conn, err := listen.Accept()
	check(err)
	defer conn.Close()

	buf := make([]byte, 1)

	for {
		_ = conn.SetReadDeadline(time.Now().Add(heartbeatTimeout))

		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Heartbeat lost, restarting main")
			return
		}
	}
}

// Starts heartbeat by connecting to port 35001 and sending a byte every 100 milliseconds
func runHeartbeat() {
	for {
		conn, err := net.Dial("tcp", heartbeatAddr)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		ticker := time.NewTicker(heartbeatInterval)

		for range ticker.C {
			_, err := conn.Write([]byte{1})
			if err != nil {
				ticker.Stop()
				_ = conn.Close()
				break
			}
		}
	}
}

// Error checking helper function, panics if there is an error
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Creates a copy with the same ID
func create_copy(id int) {
	exe, err := os.Executable()
	check(err)

	cmd := exec.Command(exe, "--id", strconv.Itoa(id))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil

	err = cmd.Start()
	check(err)
}
