# fixThis.md

Issues found across the project, organized by category.

---

## Spec-Related

### 1. `clearAtFloor` in cost.go clears all directions instead of matching direction only

**File:** `coordinator/cost.go:150-163`

The cost simulation clears all three order types (up, down, cab) at every stop, while the real FSM (`fsm.go:85-114`) only clears the direction-matching call. This makes the cost function optimistic — it thinks an elevator clears more work per stop than it actually does, which can lead to suboptimal order assignment.

Affects the secondary spec requirement: *"Calls should be distributed across the elevators in such a way that they are serviced as soon as possible."*

**Fix:** Mirror the FSM's logic:

```go
func clearAtFloor(orders *[config.N][3]bool, floor int, dir int) {
    orders[floor][2] = false // always clear cab
    switch dir {
    case 1:
        if orders[floor][0] {
            orders[floor][0] = false
        } else if !ordersAbove(floor, *orders) {
            orders[floor][1] = false
        }
    case -1:
        if orders[floor][1] {
            orders[floor][1] = false
        } else if !ordersBelow(floor, *orders) {
            orders[floor][0] = false
        }
    case 0:
        orders[floor][0] = false
        orders[floor][1] = false
    }
}
```

---

## Code Quality

### 2. Busy-wait loop in `ElevInit`

**File:** `controller/elev_init.go:22-24`

```go
for elevio.GetFloor() == -1 {

}
```

This spins the CPU at full speed waiting for the elevator to reach a floor. Each iteration also does a TCP round-trip via `elevio.GetFloor()`, hammering the elevator server.

**Fix:** Add a small sleep:

```go
for elevio.GetFloor() == -1 {
    time.Sleep(elevio._pollRate) // or just 20ms
}
```

### 3. Empty if-branch in `ElevatorController`

**File:** `controller/elev_controller.go:74-78`

```go
if !fsm.Orders[no.Floor][elevio.BT_Cab] {

} else {
    orderCh <- no
}
```

The intent is "only send if the order is still active", but the empty branch is confusing.

**Fix:** Invert the condition:

```go
if fsm.Orders[no.Floor][elevio.BT_Cab] {
    orderCh <- no
}
```

### 4. Dead commented-out code in `clearAtFloor`

**File:** `coordinator/cost.go:154-162`

A commented-out switch block sits below the active code. Either implement it (see issue #1) or remove it.

**Fix:** Remove the commented-out block, or replace the function body with the direction-aware logic.

### 5. Terse variable names throughout

Several variables are too short to read without context:

| Variable | File | Clearer name |
|----------|------|-------------|
| `no` | `elev_controller.go:59` | `order` |
| `ao` | `elev_controller.go:95` | `assigned` |
| `ro` | `system_coordinator.go:72` | (already in case, but `removeOrd` if extracted) |
| `wv` | `system_coordinator.go:12` | `worldview` |
| `sm` | `system_coordinator.go:83` | `status` |
| `hb` | `system_coordinator.go:23` | `heartbeat` |
| `no` | `debug/print.go:59` | `order` |

**Fix:** Rename to the clearer alternatives. Single-letter names like `f` or `i` in tight loops are fine; multi-letter abbreviations for domain objects are not.

### 6. Stale fields bug in `OrdersFromKB`

**File:** `debug/print.go:58-90`

The `no` struct is declared once outside the loop. When a user types `2 c`, `no.Cab` is set to `true`. On the next input `3 u`, `no.Dir` is set but `no.Cab` is still `true` from before. The coordinator checks `Cab` first, so this sends a cab call instead of a hall-up call.

**Fix:** Reset `no` at the start of each iteration:

```go
for {
    var no models.Order
    // ...
}
```

### 7. Inconsistent direction representation

The controller uses `elevio.MotorDirection` (typed int: 1, -1, 0), but the coordinator uses raw `int`. The conversion happens implicitly via `int(fsm.Direction)` in `elev_controller.go:54` and is never documented.

The cost function uses `1` for up, `-1` for down, `0` for stop — matching `MotorDirection` by coincidence, but `targetDir` uses `0` for up and `1` for down (array indices matching `BT_HallUp`/`BT_HallDown`). These two direction schemes coexist in the same file.

**Fix:** Define a shared direction type or constants in `models/` that both packages use, or at minimum add a comment at the top of `cost.go` documenting the two direction conventions used.

### 8. `maxInt` magic number in cost.go

**File:** `coordinator/cost.go:29`

```go
bestCost := int(^uint(0) >> 1) //max int
```

Works but is cryptic.

**Fix:** Use `math.MaxInt`:

```go
import "math"
bestCost := math.MaxInt
```

### 9. Typo in comment

**File:** `coordinator/cost.go:84`

`"sould"` should be `"should"`.

### 10. Typo in main.go

**File:** `main.go:26`

`"standartd"` should be `"standard"`.

### 11. Error handling in `main.go`

**File:** `main.go:18-21`

`LocalIP()` error is printed but execution continues. If `ipStr` is empty, `net.ParseIP` returns nil, and `ip.To4()[3]` panics.

**Fix:** Fatal on error:

```go
ipStr, err := network.LocalIP()
if err != nil {
    log.Fatal("Error finding IP: ", err)
}
```

### 12. Ignored error on `strconv.Atoi`

**File:** `main.go:29`

```go
id, _ = strconv.Atoi(os.Args[1])
```

If the argument is not a valid integer, `id` silently becomes 0.

**Fix:** Check the error:

```go
var err error
id, err = strconv.Atoi(os.Args[1])
if err != nil {
    log.Fatal("Invalid ID: ", os.Args[1])
}
```

### 13. Ignored error on `ResolveUDPAddr`

**File:** `network/broadcast.go:17`

```go
addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", config.Port))
```

Error is silently discarded. If resolution fails, `addr` is nil and `WriteTo` will panic.

**Fix:** Check the error or use `log.Fatal`.

### 14. `DialBroadcastUDP` doesn't return errors

**Files:** `network/conn_darwin.go`, `network/conn_linux.go`

Every socket operation prints the error but continues. If the socket fails to create, `conn` is nil and the caller panics on first use.

**Fix:** Return `(net.PacketConn, error)` and let the caller handle it, or `log.Fatal` on failure since the system can't function without the socket.

### 15. Unexplained buffer size on `assignCh`

**File:** `main.go:44`

```go
assignCh := make(chan models.Order, 8)
```

This is the only buffered channel. The buffer size of 8 has no documented rationale. If the controller can't keep up, the coordinator blocks on the 9th unprocessed assignment.

**Fix:** Add a comment explaining why 8 was chosen, or derive it from `config.N` (e.g., `2 * config.N` since there are at most N up-calls and N down-calls).

### 16. No doc comments on exported functions

None of the exported functions (`SystemCoordinator`, `ElevatorController`, `Assign`, `Cost`, `MergeWorldview`, `ComputeHallLights`, `DetectDisconnections`, `ElevInit`) have doc comments. Go convention is a `// FunctionName ...` comment on every export.

**Fix:** Add a one-line doc comment to each exported function describing its purpose.
