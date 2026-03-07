package debug

import (
	"FinalProject_G92/config"
	"FinalProject_G92/models"
	"fmt"
	"sort"
)

//testing on my rp

func PrintHallCalls(hc [config.N]models.HallCall) {
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

func PrintLobby(lobby map[int]models.Node) {
	keys := make([]int, 0, len(lobby))
	for k := range lobby {
		if lobby[k].Alive {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)

	for _, k := range keys {
		fmt.Printf("    Node %-6d", k)
	}
	fmt.Println()

	for i := config.N - 1; i >= 0; i-- {
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

func OrdersFromKB(newOrder, removeOrder chan models.Order) {
	var no models.Order
	var floor int
	var dir string

	for {
		fmt.Print("Floor and direction (e.g. '2 u'): \n")
		fmt.Scan(&floor, &dir)
		if floor >= 0 && floor < config.N {
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
