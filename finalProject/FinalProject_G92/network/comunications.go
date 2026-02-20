package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sort"
	"time"
)

const (
	// number of floors
	N = 4
	//UPD Broadcast port
	Port = 3000
)

// move types to types.go file, possibly?
type HallCall struct {
	Up      bool
	Down    bool
	UpSeq   int
	DownSeq int
}

type Order struct {
	//true Up | false Down
	Cab   bool
	Dir   bool
	Floor int
}

type Worldview struct {
	HallCalls  [N]HallCall
	CabCalls   [N]bool
	CabCallLog map[int][N]bool
}

type Heartbeat struct {
	ID        int
	IP        net.IP
	Worldview Worldview
}

type Node struct {
	Alive     bool
	Lastseen  time.Time
	Worldview Worldview
}

func PrintHallCalls(hc [N]HallCall) {

	for i := len(hc) - 1; i >= 0; i-- {

		up, down := "-", "-"
		if hc[i].Up {
			up = "↑"
		}
		if hc[i].Down {
			down = "↓"
		}

		fmt.Printf("%d| %s | %s \n", i, up, down)
	}
	fmt.Println()
}

func PrintLobby(lobby map[int]Node) {
	keys := make([]int, 0, len(lobby))
	for k := range lobby {
		if lobby[k].Alive {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)

	// Print header
	for _, k := range keys {
		fmt.Printf("    Node %-6d", k)
	}
	fmt.Println()

	// Print rows from top to bottom
	for i := N - 1; i >= 0; i-- {
		for _, k := range keys {
			up, down, cab := "-", "-", "◯"
			if lobby[k].Worldview.HallCalls[i].Up {
				up = "↑"
			}
			if lobby[k].Worldview.HallCalls[i].Down {
				down = "↓"
			}
			if lobby[k].Worldview.CabCalls[i] {
				cab = "⏺"
			}

			fmt.Printf("%d| %s | %s | %s   ", i, up, down, cab)
		}
		fmt.Println()
	}
	fmt.Println()
}

func Heart(worldviewCh chan Worldview, ip net.IP, id int) {

	conn := DialBroadcastUDP(Port)
	defer conn.Close()

	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", Port))

	var wv Worldview

	//once a secodn to facilitate testing - Normaly, would be 100ms
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case wv = <-worldviewCh:

		case <-ticker.C:
			hb := Heartbeat{ID: id, IP: ip, Worldview: wv}

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

func Listener(heartbeatCh chan Heartbeat) {
	conn := DialBroadcastUDP(Port)
	defer conn.Close()

	buf := make([]byte, 2048)

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

// TO BE REPLACED BY HW
func OrdersFromKB(newOrder, removeOrder chan Order) {
	//taking keyboard input for tests
	var no Order

	var floor int
	var dir string

	for {
		fmt.Print("Floor and direction (e.g. '2 u'): \n")
		fmt.Scan(&floor, &dir)
		if floor >= 0 && floor < N {
			no.Floor = floor
			switch dir {
			case "u":
				no.Dir = true
				newOrder <- no
			case "d":
				no.Dir = false
				newOrder <- no
			case "c":
				no.Cab = true
				newOrder <- no

			case "U":
				no.Dir = true
				removeOrder <- no
			case "D":
				no.Dir = false
				removeOrder <- no
			case "C":
				no.Cab = true
				removeOrder <- no

			}

		}
	}
}

// NEEDS TO SEND INFORMATION TO HW
func UpdateLights(lobby map[int]Node) [N]HallCall {
	var lights [N]HallCall
	for i := range N { //for every floor

		allUp := true
		for _, elevator := range lobby {
			if !elevator.Worldview.HallCalls[i].Up {
				allUp = false
				break
			}
		}
		lights[i].Up = allUp

		allDown := true
		for _, elevator := range lobby {
			if !elevator.Worldview.HallCalls[i].Down {
				allDown = false
				break
			}
		}
		lights[i].Down = allDown
	}

	fmt.Println("lights:")
	PrintHallCalls(lights)

	return lights

}

func NetworkManager(myId int, worldviewCh chan Worldview, heartbeatCh chan Heartbeat, newOrder, removeOrder chan Order) {

	var wv Worldview
	var hb Heartbeat

	lobby := make(map[int]Node)
	wv.CabCallLog = make(map[int][N]bool)

	booted := false

	//managing disconnections
	const DisconnectTimeout = 3 * time.Second
	disconnectTicker := time.NewTicker(1 * time.Second)
	defer disconnectTicker.Stop()

	//syncing from incomming heartbeats
	for {
		select {
		case hb = <-heartbeatCh:

			node := lobby[hb.ID]
			node.Worldview.HallCalls = hb.Worldview.HallCalls

			if !booted {
				if myCabCalls, ok := hb.Worldview.CabCallLog[myId]; ok {
					wv.CabCalls = myCabCalls
					booted = true
				}
			}

			node.Worldview.CabCalls = hb.Worldview.CabCalls

			node.Lastseen = time.Now()
			node.Alive = true
			lobby[hb.ID] = node

			for i := range N {
				if wv.HallCalls[i].UpSeq < hb.Worldview.HallCalls[i].UpSeq {
					wv.HallCalls[i].Up = hb.Worldview.HallCalls[i].Up
					wv.HallCalls[i].UpSeq = hb.Worldview.HallCalls[i].UpSeq
				}
				if wv.HallCalls[i].DownSeq < hb.Worldview.HallCalls[i].DownSeq {
					wv.HallCalls[i].Down = hb.Worldview.HallCalls[i].Down
					wv.HallCalls[i].DownSeq = hb.Worldview.HallCalls[i].DownSeq
				}
			}

			for key := range lobby {
				wv.CabCallLog[key] = lobby[key].Worldview.CabCalls
			}
			worldviewCh <- wv

			PrintLobby(lobby)
			UpdateLights(lobby)

		case no := <-newOrder:
			if no.Cab {
				wv.CabCalls[no.Floor] = true
			} else if no.Dir {
				wv.HallCalls[no.Floor].Up = true
				wv.HallCalls[no.Floor].UpSeq++
			} else {
				wv.HallCalls[no.Floor].Down = true
				wv.HallCalls[no.Floor].DownSeq++
			}

		case ro := <-removeOrder:
			if ro.Cab {
				wv.CabCalls[ro.Floor] = false
			} else if ro.Dir {
				wv.HallCalls[ro.Floor].Up = false
				wv.HallCalls[ro.Floor].UpSeq++
			} else {
				wv.HallCalls[ro.Floor].Down = false
				wv.HallCalls[ro.Floor].DownSeq++
			}

		case <-disconnectTicker.C:
			for id, node := range lobby {
				if node.Alive && time.Since(node.Lastseen) > DisconnectTimeout {
					node.Alive = false
					lobby[id] = node
					fmt.Printf("Node %d disconnected", id)
				}
			}
		}

	}

}
