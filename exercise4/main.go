package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const heartbeatPort = ":9999"
const ip = "localhost"

var count = 0

func spawnBackup() error {
	//fetching path directory
	wd, _ := os.Getwd()

	//mounting script
	script := fmt.Sprintf(`tell app "Terminal" to do script "cd %s && go run main.go backup"`, wd)

	//execution
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Start()
}

func runAsBackup() {

	//listening udp connection
	conn, err := net.ListenPacket("udp", heartbeatPort)
	if err != nil {
		fmt.Println("Failed to connect to ", heartbeatPort)
	}
	defer conn.Close()

	fmt.Println("Backup listening on port", heartbeatPort)

	//receiving
	buffer := make([]byte, 1024)
	for {
		conn.SetReadDeadline(time.Now().Add(4 * time.Second))
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			//if timeout detected, breaks out of loop ot become primary
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout!")
				return
			}
		}

		message := string(buffer[:n])
		count, err = strconv.Atoi(message)
		if err != nil {
			fmt.Println("Failed to get last count (string to int conversion unsuccessful)")
		}

		fmt.Println(count)

	}
}

func runAsPrimary() {

	fmt.Printf("Starting up as prmary\n\ncount is %d\n", count)
	err := spawnBackup()
	if err != nil {
		fmt.Println("Couldn't spawn backup")
	}
	fmt.Println("Backup spawned!")
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("udp", ip+heartbeatPort)
	if err != nil {
		fmt.Println("")
	}
	defer conn.Close()

	fmt.Println("connected!")

	for {

		message := fmt.Sprintf("%d", count)
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("error sending count")
		}

		fmt.Println(count)

		count++
		time.Sleep(1 * time.Second)
	}

}

func main() {

	//Check if I'm primary or backup via command line arguments
	imBackup := false
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		imBackup = true
	}

	//backup routine
	if imBackup {
		runAsBackup()
	}

	runAsPrimary()
	return
}
