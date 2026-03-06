package network

import (
	"FinalProject_G92/config"
	"FinalProject_G92/types"
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

func HeartbeatSender(worldviewCh chan types.Worldview, ip net.IP, id int) {
	conn := DialBroadcastUDP(config.Port)
	defer conn.Close()

	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", config.Port))

	var wv types.Worldview

	ticker := time.NewTicker(config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case wv = <-worldviewCh:

		case <-ticker.C:
			hb := types.Heartbeat{ID: id, IP: ip, Worldview: wv}

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

func HeartbeatListener(heartbeatCh chan types.Heartbeat) {
	conn := DialBroadcastUDP(config.Port)
	defer conn.Close()

	buf := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("error reading: ", err)
			continue
		}

		var hb types.Heartbeat
		dec := gob.NewDecoder(bytes.NewReader(buf[:n]))
		if err := dec.Decode(&hb); err != nil {
			fmt.Println("Error decoding Heartbeat: ", err)
			continue
		}

		heartbeatCh <- hb
	}
}
