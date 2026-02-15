package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

const (
	// number of floors
	N = 4
	//UPD Broadcast port
	Port = 3000
)

// move types to types.go file, possibly?
type Call struct {
	Up   bool
	Down bool
}

type Heartbeat struct {
	IP        net.IP
	Worldview [N]Call
}

func PrintWorldView(wv [N]Call) {

	for i := len(wv) - 1; i >= 0; i-- {

		up, down := "-", "-"
		if wv[i].Up {
			up = "↑"
		}
		if wv[i].Down {
			down = "↓"
		}
		fmt.Printf("%d| %s | %s \n", i, up, down)

	}
	fmt.Println()
}

func Heart(wordlviewCh chan [N]Call, ip net.IP) {

	conn := DialBroadcastUDP(Port)
	defer conn.Close()

	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", Port))

	var wv [N]Call

	//once a secodn to facilitate testing - Normaly, would be 100ms
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case wv = <-wordlviewCh:

		case <-ticker.C:
			hb := Heartbeat{IP: ip, Worldview: wv}

			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			if err := enc.Encode(hb); err != nil {
				fmt.Println("Error encoding heartbeat: ", err)
				continue
			}

			_, err := conn.WriteTo(buf.Bytes(), addr)
			if err != nil {
				fmt.Println("Error sending heartbeat: ", err)
			}

		}

	}

}

func Listener(heartbeatCh chan Heartbeat, ip net.IP) {
	conn := DialBroadcastUDP(Port)
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("error reading: ", err)
			continue
		}

		var hb Heartbeat
		dec := gob.NewDecoder(bytes.NewReader(buf[:n]))
		if err := dec.Decode(&hb); err != nil {
			fmt.Println("Error decoding Heartbeat: ", err)
			continue
		}

		heartbeatCh <- hb
	}
}

func WorldviewManager(worldviewCh chan [N]Call, heartbeatCh chan Heartbeat) {

	var wv [N]Call

	var floor int
	var dir string

	//taking keyboard input for tests
	for {
		fmt.Print("Floor and direction (e.g. '2 u'): \n")
		fmt.Scan(&floor, &dir)
		if floor >= 0 && floor < N {
			switch dir {
			case "u":
				wv[floor].Up = !wv[floor].Up
			case "d":
				wv[floor].Down = !wv[floor].Down
			default:
				fmt.Println("Use 'u' or 'd'")
				continue
			}
			worldviewCh <- wv
		}
	}
}
