package costfunction

import (
	"encoding/json"
	"heis/config"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func Compute(
	myID string,
	hallRequests [][2]bool,
	states map[string]HRAElevState,
) ([][2]bool, bool) {

	input := HRAInput{
		HallRequests: hallRequests,
		States:       states,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, false
	}

	rawOutput, err := exec.Command(executableName(), "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		return nil, false
	}

	var result map[string][][2]bool
	if err := json.Unmarshal(rawOutput, &result); err != nil {
		return nil, false
	}

	assigned, ok := result[myID]
	if !ok {
		return nil, false
	}

	return assigned, true

}
func executableName() string {
	switch runtime.GOOS {
	case "linux":
		return config.HRAPath + "hall_request_assigner"
	case "windows":
		return config.HRAPath + "hall_request_assigner.exe"
	default:
		panic("unsupported OS: " + runtime.GOOS)
	}
}

