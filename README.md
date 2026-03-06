# FinalProject_G92 — Distributed Elevator Control System

A distributed real-time elevator control system in Go. Multiple elevator nodes communicate over UDP broadcast to coordinate hall calls, synchronize worldviews, and serve orders. Each node controls one physical elevator and collaborates with peers to provide consistent service.

## Project Structure

```
FinalProject_G92/
├── main.go                          # Entry point: wiring and goroutine launch
├── config/
│   └── config.go                    # Shared constants
├── types/
│   └── types.go                     # Shared domain types
├── hardware/
│   ├── elevio/
│   │   └── elevio.go                # Low-level elevator I/O driver
│   ├── elevator_states.go           # ElevatorState struct and initialization
│   ├── fsm.go                       # Elevator finite state machine
│   ├── manager.go                   # Hardware event loop (bridge between HW and network)
│   └── lights.go                    # Hall light updates from network consensus
├── network/
│   ├── broadcast.go                 # UDP heartbeat sender and listener
│   ├── conn_linux.go                # Linux UDP broadcast socket
│   ├── conn_darwin.go               # macOS UDP broadcast socket
│   ├── localip.go                   # Local IP address discovery
│   ├── manager.go                   # NetworkManager select loop
│   └── sync.go                      # Worldview merging, light consensus, disconnect detection
└── debug/
    └── print.go                     # Debug printing and keyboard input for testing
```

## Dependency Graph

```
main ──> hardware, network, debug, config, types
hardware ──> config, types, elevio
network ──> config, types
debug ──> config, types
```

No circular dependencies. All packages depend downward into `config` and `types`.

---

## Module Reference

### `main.go` — Entry Point

The sole job of `main.go` is to wire everything together:

1. Discovers the local IP address
2. Initializes the elevator driver connection (`elevio.Init`)
3. Parses the optional node ID from command-line args (for multi-instance testing on one machine)
4. Creates all channels that connect the goroutines
5. Launches goroutines and blocks on `NetworkManager`

**Goroutines launched:**
- `HeartbeatListener` — receives UDP heartbeats from peers
- `HeartbeatSender` — broadcasts this node's worldview
- `OrdersFromKB` — debug: manual order input from keyboard
- `HallLights` — updates hall button lamps when network consensus changes
- `HardwareManager` — polls elevator hardware and feeds orders into the network

**Why:** Keeps all dependency wiring visible in one place. No business logic lives here.

---

### `config/config.go` — Shared Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `N` | `4` | Number of floors |
| `Port` | `3000` | UDP broadcast port |
| `DisconnectTimeout` | `3s` | Time before a silent node is marked disconnected |
| `HeartbeatInterval` | `1s` | How often each node broadcasts its state |

**Why:** Centralizes magic numbers that were previously scattered across `network` and `hardware`. Every package imports `config` instead of reaching into each other for constants.

---

### `types/types.go` — Shared Domain Types

| Type | Fields | Purpose |
|------|--------|---------|
| `HallCall` | `Up`, `Down`, `UpSeq`, `DownSeq` | Represents hall call state for one floor. The sequence numbers (`UpSeq`, `DownSeq`) enable conflict-free merging — the highest sequence number wins. |
| `Order` | `Cab`, `Dir`, `Floor` | A request to add or remove an order. `Cab=true` means cab call, otherwise `Dir=true` means up, `Dir=false` means down. |
| `Worldview` | `HallCalls`, `CabCalls`, `CabCallLog` | A node's view of all orders. `HallCalls` are shared across nodes. `CabCalls` are this node's cab calls. `CabCallLog` backs up every node's cab calls for crash recovery. |
| `Heartbeat` | `ID`, `IP`, `Worldview` | The packet broadcast over UDP. Contains the sender's identity and their full worldview. |
| `Node` | `Alive`, `Lastseen`, `Worldview` | Represents a peer in the lobby. Tracks liveness via `Lastseen` timestamp. |

**Why:** These types are used by both `hardware` and `network`. Placing them in a neutral package prevents the dependency cycle that previously existed (`hardware` importing `network`).

---

### `hardware/elevio/elevio.go` — Low-Level Elevator Driver

A provided library that communicates with the elevator server over TCP. It sends 4-byte commands and reads 4-byte responses.

**Key functions:**
- `Init(addr, numFloors)` — connects to the elevator server via TCP
- `SetMotorDirection`, `SetButtonLamp`, `SetFloorIndicator`, `SetDoorOpenLamp` — write commands
- `GetFloor`, `GetButton`, `GetStop`, `GetObstruction` — read sensor state
- `PollButtons`, `PollFloorSensor`, `PollStopButton`, `PollObstructionSwitch` — polling goroutines that send events on channels when state changes

**Why:** This is the hardware abstraction layer. All physical I/O goes through here. The rest of the system never touches the TCP connection directly.

---

### `hardware/elevator_states.go` — Elevator State & Initialization

Defines `ElevatorState` — a snapshot of the elevator's physical condition (floor, direction, stopped, doors, etc.).

`ElevInit(eState)` puts the elevator into a known-good state on startup:
- Stops the motor
- Reads the current floor
- Turns off all lamps
- Closes doors

**Why:** Ensures the elevator starts from a clean state regardless of how it was left by a previous run.

---

### `hardware/fsm.go` — Elevator Finite State Machine

Models the elevator's behavior as four states with deterministic transitions:

```
         ButtonPress              FloorArrival (should stop)
  Idle ──────────────> Moving ──────────────────────────> DoorOpen
   ^                                                        │
   │                    doorOpenDuration (3s)                │
   └────────────────────────────────────────────────────────┘
                         (no more orders)

  Any state ──StopButton──> Stopped
```

**States:**
- `Idle` — stationary, no pending orders
- `Moving` — traveling toward an order
- `DoorOpen` — arrived at floor, doors open for 3 seconds
- `Stopped` — emergency stop pressed

**Core logic:**
- `OnButtonPress(floor, btn)` — stores the order. If idle, picks a direction and starts moving.
- `OnFloorArrival(floor)` — decides whether to stop (matching order, cab call, or no orders ahead). If stopping, opens doors for 3s then picks next direction.
- `shouldStop()` — stops if there's a matching directional order, a cab call, or no more orders in the current direction.
- `chooseDirectionAndMove()` — continues in the same direction if there are orders ahead, reverses if needed, or goes idle if no orders remain.

**Why:** Replaces the ad-hoc order handling that had critical bugs (using empty variables instead of actual event data, empty hall call cases). The FSM provides a structured, correct way to handle all button types and manage the elevator lifecycle.

---

### `hardware/manager.go` — Hardware Manager

The bridge between physical elevator events and the network layer. Runs as a goroutine with a select loop that:

1. **Floor arrival** (`floorCh`) — Updates state, snapshots current orders, runs `fsm.OnFloorArrival()`. If the FSM cleared any orders at this floor (the elevator served them), sends remove orders on `rmOrderCh` to notify the network.

2. **Button press** (`btnCh`) — Converts the `elevio.ButtonEvent` into a `types.Order` and sends it on `orderCh` to the network. Also feeds the FSM via `OnButtonPress` so the elevator starts moving.

3. **Stop button** (`stopCh`) — Delegates to `fsm.OnStopButton()`.

4. **Obstruction** (`obstrCh`) — Logs the event (handling TBD).

**Why:** This is the function that was on a different branch (`hwManager.go`). It completes the system by connecting hardware inputs to network outputs. Without it, button presses on the physical elevator would never reach the network, and completed orders would never be removed.

---

### `hardware/lights.go` — Hall Light Updates

```go
func HallLights(lightsCh <-chan [config.N]types.HallCall)
```

Receives hall light state from the network (via `lightsCh`) and sets the physical button lamps accordingly. Runs as a goroutine.

**Why:** Hall lights must reflect network consensus (all nodes agree a call exists), not just local state. This function is event-driven — it only updates lamps when the network sends a new consensus, avoiding the CPU-spinning tight loop that existed before.

---

### `network/broadcast.go` — UDP Heartbeat I/O

**`HeartbeatSender(worldviewCh, ip, id)`**

Broadcasts this node's worldview to all peers every `HeartbeatInterval` (1s) via UDP broadcast on port 3000. Uses `gob` encoding to serialize the `Heartbeat` struct. Picks up the latest worldview from `worldviewCh` whenever the NetworkManager sends one.

**`HeartbeatListener(heartbeatCh)`**

Listens for incoming UDP broadcasts. Decodes each packet into a `Heartbeat` and sends it to `heartbeatCh` for the NetworkManager to process.

**Why:** Separated from the manager because these are pure I/O functions — they serialize/deserialize and send/receive. They have no knowledge of worldview merging or order logic.

---

### `network/conn_linux.go` / `conn_darwin.go` — Platform-Specific UDP

Creates a UDP broadcast socket using platform-specific syscalls.

- Linux: `SO_REUSEADDR` + `SO_BROADCAST`
- macOS: same + `SO_REUSEPORT` (required on Darwin for multiple processes to bind the same port)

**Why:** UDP broadcast requires raw socket options that differ by OS. Build tags (`//go:build linux` / `//go:build darwin`) ensure the correct implementation compiles on each platform.

---

### `network/localip.go` — Local IP Discovery

Discovers this machine's LAN IP by briefly connecting to `8.8.8.8:53` (Google DNS) and reading the local address from the connection. The result is cached.

**Why:** The node needs its own IP to include in heartbeats and to connect to the local elevator server. Moved out of `main.go` because it's a network utility.

---

### `network/manager.go` — Network Manager

The central coordination loop. Runs on the main goroutine (blocking).

**Select cases:**

1. **Incoming heartbeat** — Updates the lobby entry for that peer. On first boot, recovers this node's cab calls from the peer's `CabCallLog` (crash recovery). Merges the peer's hall calls into the local worldview using sequence numbers. Updates the cab call log. Sends the updated worldview to the heartbeat sender. Sends the computed hall light consensus to `lightsCh`.

2. **New order** — Adds a hall call or cab call to the local worldview (with sequence number bump for hall calls). The next heartbeat will propagate it.

3. **Remove order** — Clears a hall call or cab call from the local worldview (with sequence number bump). The next heartbeat will propagate the removal.

4. **Disconnect ticker** — Every second, checks if any node hasn't been heard from in `DisconnectTimeout` and marks them dead.

**Why:** This is the brain of the network layer. It maintains the lobby (all known peers) and ensures worldviews converge across nodes. Kept slim by delegating logic to `sync.go` functions.

---

### `network/sync.go` — Worldview Synchronization

Pure (or near-pure) functions extracted from NetworkManager for testability:

- **`MergeWorldview(local, remote)`** — For each floor, adopts the remote's hall call state if its sequence number is higher. This is how nodes converge: the highest sequence number always wins, ensuring that a new call or a removal propagates to all nodes.

- **`UpdateCabCallLog(wv, lobby)`** — Snapshots every node's cab calls into the local worldview's `CabCallLog`. This log is broadcast in heartbeats so that if a node crashes and reboots, it can recover its cab calls from a peer.

- **`ComputeHallLights(lobby)`** — Determines which hall lights should be on. A light is on only if **all** alive nodes agree the call exists (consensus). This prevents flickering during propagation.

- **`DetectDisconnections(lobby, timeout)`** — Marks nodes as dead if they haven't sent a heartbeat within the timeout.

**Why:** Each function is independently testable and has a clear contract. Separating them from the manager's select loop makes the system easier to reason about and debug.

---

### `debug/print.go` — Debug & Test Utilities

- **`PrintHallCalls(hc)`** — Prints a floor-by-floor view of hall calls (arrows for active calls).
- **`PrintLobby(lobby)`** — Prints a multi-column view of all alive nodes' hall calls and cab calls.
- **`OrdersFromKB(newOrder, removeOrder)`** — Reads keyboard input to manually add/remove orders. Input format: `<floor> <direction>` where direction is `u`/`d`/`c` (add) or `U`/`D`/`C` (remove).

**Why:** These are test scaffolding, not production logic. Separating them into `debug/` makes it clear they're not part of the runtime system and prevents them from cluttering production packages.

---

## Data Flow

```
                    Physical Elevator
                          │
                    ┌─────┴─────┐
                    │   elevio   │  (TCP to elevator server)
                    └─────┬─────┘
                          │
              floor/button/stop/obstruction events
                          │
                    ┌─────┴─────┐
                    │ Hardware   │
                    │ Manager    │  polls HW, feeds FSM, emits Orders
                    └──┬─────┬──┘
                       │     │
              orderCh  │     │  rmOrderCh
                       │     │
                    ┌──┴─────┴──┐
                    │  Network   │
                    │  Manager   │  merges worldviews, manages lobby
                    └──┬─────┬──┘
                       │     │
          worldviewCh  │     │  lightsCh
                       │     │
            ┌──────────┴┐   ┌┴──────────┐
            │ Heartbeat  │   │ HallLights │
            │ Sender     │   │           │
            └─────┬──────┘   └───────────┘
                  │                sets physical lamps
            UDP broadcast
                  │
           ┌──────┴──────┐
           │  Heartbeat   │
           │  Listener    │
           └──────┬──────┘
                  │
            heartbeatCh ──> NetworkManager
```

## Worldview Synchronization Protocol

1. Each node maintains its own `Worldview` (hall calls with sequence numbers, cab calls, cab call backup log).
2. Every second, the node broadcasts its worldview as a `Heartbeat` via UDP.
3. When a node receives a heartbeat, it merges the remote worldview: for each hall call, the **higher sequence number wins**. This ensures that both additions (`Up=true, Seq++`) and removals (`Up=false, Seq++`) propagate deterministically.
4. Cab calls are **not** merged — they are local to each elevator. But every node backs up every other node's cab calls in `CabCallLog`, so a rebooting node can recover its cab calls from a peer.
5. Hall lights use **consensus**: a light turns on only when all alive nodes agree the call exists. This prevents lights from flickering during propagation.

## Running

```bash
# Start the elevator server first (e.g., SimElevatorServer on port 15657)

# Run a single node (ID defaults to 0)
go run main.go

# Run multiple nodes on the same machine with different IDs
go run main.go 0
go run main.go 1
go run main.go 2
```

The node connects to the elevator server at `<localIP>:15657` and broadcasts/listens on UDP port 3000.
