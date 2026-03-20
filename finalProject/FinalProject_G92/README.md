# FinalProject_G92 — Distributed Elevator Control System

A distributed real-time elevator control system in Go. Multiple elevator nodes communicate over UDP broadcast to coordinate hall calls, synchronize worldviews, and assign orders via a cost function. Each node controls one physical elevator and collaborates with peers to provide consistent, fault-tolerant service across 4 floors.

## Requirements Fulfilled

This system satisfies the specification from `Project description.pdf`:

**Main requirements:**
- **Button lights as service guarantee** — once a hall call light turns on, an elevator will arrive at that floor
- **No calls are lost** — cab calls persist through network disconnection, software crash, and power loss via backup log recovery
- **Correct light and button behavior** — hall buttons summon elevators; cab buttons are local; lights reflect the true state of each call
- **Door behavior** — doors stay open for 3 seconds; doors do not close while obstructed
- **Efficient individual behavior** — elevators do not take unnecessary stops and clear calls based on travel direction

**Secondary requirements:**
- Hall calls are distributed across elevators using a cost function that minimizes total service time

## Project Structure

```
FinalProject_G92/
├── main.go                          # Entry point: channel wiring and goroutine launch
├── config/
│   └── config.go                    # Shared constants (floors, ports, timeouts)
├── models/
│   └── models.go                    # Shared domain types (HallCall, Order, Worldview, etc.)
├── controller/                      # Elevator hardware control and FSM
│   ├── elev_controller.go           # Hardware event loop (bridge between HW and coordinator)
│   ├── fsm.go                       # Elevator finite state machine (Idle/Moving/DoorOpen/Stopped)
│   ├── elev_init.go                 # Initialization: known-good state on startup
│   ├── lights.go                    # Hall light updates from network consensus
│   └── elevio/
│       └── elevio.go                # Low-level TCP driver to elevator server
├── coordinator/                     # Network synchronization and order assignment
│   ├── system_coordinator.go        # Central coordination loop (main select)
│   ├── sync.go                      # Worldview merging, consensus, disconnection detection
│   └── cost.go                      # Cost function for intelligent order assignment
├── network/                         # UDP broadcast communication
│   ├── broadcast.go                 # Heartbeat sender and listener
│   ├── localip.go                   # Local IP address discovery
│   ├── conn_darwin.go               # macOS UDP socket (SO_REUSEPORT)
│   └── conn_linux.go                # Linux UDP socket
└── debug/
    └── print.go                     # Debug printing and keyboard input for testing
```

## Dependency Graph

```
main ──> controller, coordinator, network, debug, config, models
controller ──> config, models, elevio
coordinator ──> config, models
network ──> config, models
debug ──> config, models
```

No circular dependencies. All packages depend downward into `config` and `models`.

---

## Module Reference

### `main.go` — Entry Point

Wires everything together:

1. Discovers the local IP address
2. Derives node ID from the last octet of the IP (or from a command-line argument)
3. Initializes the elevator driver connection (`elevio.Init`)
4. Creates 7 channels connecting the goroutines
5. Launches goroutines and blocks on `SystemCoordinator`

**Goroutines launched:**
- `HeartbeatListener` — receives UDP heartbeats from peers
- `HeartbeatSender` — broadcasts this node's worldview
- `OrdersFromKB` — debug: manual order input from keyboard
- `HallLights` — updates hall button lamps from network consensus
- `ElevatorController` — polls elevator hardware, runs FSM, emits orders

---

### `config/config.go` — Shared Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `N` | `4` | Number of floors |
| `Port` | `3000` | UDP broadcast port |
| `DisconnectTimeout` | `700ms` | Time before a silent node is marked disconnected |
| `HeartbeatInterval` | `70ms` | How often each node broadcasts its state |
| `DoorOpenDuration` | `3s` | How long the door stays open |
| `BetweenFloorsDuration` | `3s` | Time to travel between adjacent floors |
| `ElevatorServerPort` | `15657` | TCP port for elevator server connection |

---

### `models/models.go` — Shared Domain Types

| Type | Fields | Purpose |
|------|--------|---------|
| `HallCall` | `Up`, `Down`, `UpSeq`, `DownSeq` | Hall call state for one floor. Sequence numbers enable conflict-free merging — highest sequence number wins. |
| `Order` | `Cab`, `Dir`, `Floor` | A request to add or remove an order. `Cab=true` means cab call; otherwise `Dir` indicates up/down. |
| `StatusMessage` | `Floor`, `Direction`, `Operational` | Snapshot of an elevator's current position, direction, and operational state. |
| `Worldview` | `HallCalls`, `CabCalls`, `CabCallLog`, `Status` | A node's complete view: shared hall calls, local cab calls, backup log of all nodes' cab calls, and elevator status. |
| `Heartbeat` | `ID`, `IP`, `Worldview` | The packet broadcast over UDP. Contains the sender's identity and full worldview. |
| `Node` | `Alive`, `Lastseen`, `Worldview` | Represents a peer. Tracks liveness via `Lastseen` timestamp. |

---

### `controller/` — Elevator Hardware Control

#### `elevio/elevio.go` — Low-Level Driver

Communicates with the elevator server over TCP using 4-byte commands/responses. Provides motor control, lamp control, sensor reads, and polling goroutines for floor, button, stop, and obstruction events.

#### `elev_init.go` — Initialization

Puts the elevator into a known-good state on startup: stops the motor, turns off all lamps, closes doors, and drives to the nearest floor if between floors.

#### `fsm.go` — Finite State Machine

Models elevator behavior as four states:

```
         ButtonPress              FloorArrival (should stop)
  Idle ──────────────> Moving ──────────────────────────> DoorOpen
   ^                                                        │
   │                    DoorOpenDuration (3s)                │
   └────────────────────────────────────────────────────────┘
                         (no more orders)

  Any state ──StopButton──> Stopped
```

Key logic:
- `shouldStop()` — stops for matching directional orders, cab calls, or when no more orders ahead
- `clearOrdersAtFloor()` — clears only the call matching travel direction (not both), enabling direction-aware service
- `chooseDirectionAndMove()` — continues in same direction if orders ahead, reverses if needed, or goes idle

#### `elev_controller.go` — Hardware Event Loop

Bridges physical elevator events and the coordinator. Polls sensors via channels and:
- Converts button presses into `Order` messages for the coordinator
- Receives assigned orders from the coordinator via `assignCh`
- Sends remove orders when the FSM clears served calls
- Reports elevator status (floor, direction, operational state)
- Handles door closing with obstruction awareness via `closeDoors()`
- Detects motor stalls with a watchdog timer

#### `lights.go` — Hall Light Updates

Receives consensus hall light state from the coordinator and sets the physical lamps. Event-driven — only updates when consensus changes.

---

### `coordinator/` — Network Synchronization & Order Assignment

#### `system_coordinator.go` — Central Coordination Loop

The brain of the system. Runs on the main goroutine with a select loop handling:

1. **Incoming heartbeat** — on first boot, recovers cab calls from the peer's `CabCallLog` (crash recovery). Merges the peer's hall calls via sequence numbers. Updates the cab call log. Computes hall light consensus. Runs the cost-based order assignment.

2. **New order** — adds a hall call or cab call to the local worldview with a sequence number bump for hall calls.

3. **Remove order** — clears a call from the local worldview with a sequence number bump.

4. **Status update** — stores the elevator's current floor, direction, and operational state.

5. **Disconnect ticker** — marks nodes as dead if no heartbeat received within `DisconnectTimeout`.

#### `sync.go` — Worldview Synchronization

Pure functions for distributed state management:

- **`MergeWorldview`** — for each floor, adopts the remote hall call state if its sequence number is higher. This is how nodes converge.
- **`UpdateCabCallLog`** — snapshots every node's cab calls for crash recovery backup.
- **`ComputeHallLights`** — a light turns on only if **all** alive nodes agree the call exists (consensus). Prevents flickering during propagation.
- **`DetectDisconnections`** — marks nodes as dead after `DisconnectTimeout` of silence.

#### `cost.go` — Order Assignment

Distributes hall calls across elevators for efficient service:

- **`Assign`** — for each consensus hall call, evaluates the cost for every alive, operational elevator and assigns the call to the cheapest one. Tie-breaker: lower ID wins. Deterministic across all nodes.
- **`Cost`** — simulates the elevator's path from its current position through its pending orders plus the target call. Accounts for door open time (3s per stop) and travel time (3s per floor). Returns total estimated seconds.

---

### `network/` — UDP Broadcast Communication

#### `broadcast.go` — Heartbeat I/O

- **`HeartbeatSender`** — broadcasts the node's worldview every 70ms via UDP on port 3000. Uses `gob` encoding.
- **`HeartbeatListener`** — listens for incoming UDP broadcasts, decodes, and forwards to the coordinator.

#### `localip.go` — IP Discovery

Discovers the machine's LAN IP by connecting to `8.8.8.8:53` and reading the local address.

#### `conn_linux.go` / `conn_darwin.go` — Platform-Specific UDP

Creates UDP broadcast sockets with platform-specific socket options. Build tags ensure the correct implementation compiles per OS.

---

### `debug/print.go` — Debug Utilities

- `PrintHallCalls` — floor-by-floor view of hall call state
- `PrintLobby` — multi-column view of all alive nodes' hall calls and cab calls
- `OrdersFromKB` — manual order input for testing (`<floor> <dir>`, e.g. `2 u` for floor 2 up, `2 U` to remove)

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
                    ┌─────┴──────┐
                    │ Elevator    │
                    │ Controller  │  polls HW, runs FSM, emits orders
                    └──┬──┬──┬───┘
                       │  │  │
            orderCh ───┘  │  └─── statusCh
            rmOrderCh ────┘
                       │  │  │
                    ┌──┴──┴──┴──┐
                    │  System    │
                    │ Coordinator│  merges worldviews, assigns orders, manages lobby
                    └──┬──┬──┬──┘
                       │  │  │
         worldviewCh ──┘  │  └── assignCh (back to controller)
                   lightsCh
                       │  │
           ┌───────────┴┐ ┌┴──────────┐
           │ Heartbeat   │ │ HallLights │
           │ Sender      │ │            │
           └─────┬───────┘ └────────────┘
                 │               sets physical lamps
           UDP broadcast
                 │
          ┌──────┴──────┐
          │  Heartbeat   │
          │  Listener    │
          └──────┬──────┘
                 │
           heartbeatCh ──> SystemCoordinator
```

## Synchronization Protocol

1. Each node maintains its own `Worldview` with sequence-numbered hall calls, local cab calls, and a backup log of all nodes' cab calls.
2. Every 70ms, the node broadcasts its worldview as a `Heartbeat` via UDP.
3. On receiving a heartbeat, the node merges hall calls: **higher sequence number wins**. Both additions and removals bump the sequence, ensuring deterministic propagation.
4. Cab calls are **not merged** — they are local to each elevator. But every node backs up every other node's cab calls in `CabCallLog`, enabling crash recovery.
5. Hall lights use **consensus**: a light turns on only when all alive nodes agree the call exists.
6. On reboot, a node recovers its cab calls from the first peer's `CabCallLog[myId]`.

## Order Assignment

When consensus hall calls are computed, every node independently runs the same cost function to determine which elevator should serve each call:

1. For each active consensus call, simulate every alive elevator's path to that call
2. The cost is the total estimated time (travel + door stops) to reach the target floor
3. The elevator with the lowest cost is assigned the call
4. Ties are broken by lower node ID

Because all nodes use the same inputs (consensus calls + lobby state), they all arrive at the same assignment — no centralized dispatcher needed.

## Running

```bash
# Start the elevator server first (e.g., SimElevatorServer on port 15657)

# Run a single node (ID derived from IP last octet)
cd finalProject/FinalProject_G92
go run main.go

# Run multiple nodes on the same machine with different IDs
go run main.go 0   # connects to elevator server on port 15657
go run main.go 1   # connects to elevator server on port 15658
go run main.go 2   # connects to elevator server on port 15659
```

Each node connects to the elevator server at `<localIP>:<15657+ID>` and broadcasts/listens on UDP port 3000.
