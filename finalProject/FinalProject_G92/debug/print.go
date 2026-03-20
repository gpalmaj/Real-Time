package debug

import (
	"FinalProject_G92/config"
	"FinalProject_G92/models"
	"fmt"
	"sort"
)

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
