# Peer Reviews

---

## Project 0910 (FinalProject_G92)

**Score: 8 — Reasonable with caveats.**

### Feedback

- **Strong module decomposition.** The split into `controller/`, `coordinator/`, `network/`, `models/`, and `config/` is clean and each package has a well-defined responsibility. The coordinator owns distributed state, the controller owns the FSM and hardware, and the network owns UDP broadcast. This is a textbook example of good cohesion — each module presents one consistent abstraction.

- **Sequence-based Last-Write-Wins for hall calls is elegant.** Using monotonically increasing sequence numbers for both additions and removals of hall calls is a sound distributed systems approach. It gives you conflict-free merging without a central authority. The consensus-based lights (AND across all alive nodes) are a nice touch that prevents user-facing flickering.

- **The constant `N = 4` is an unacceptable name.** A single-letter constant for the number of floors violates every naming principle. Anyone reading `for i := 0; i < config.N; i++` has to go look up what `N` means. It should be `NumFloors` or `FloorCount`. This is the kind of name that makes code silently hostile to new readers.

- **Pervasive abbreviations erode readability.** `wv` for worldview, `hb` for heartbeat, `rm` for remove, `btn` for button, `obstr` for obstruction — these are all shortcuts that save a few keystrokes and cost minutes of comprehension for anyone who didn't write them. The naming guidelines are clear: variable names must describe what the variable represents without abbreviation. `worldview`, `heartbeat`, `removeOrder`, `button`, `obstruction` — none of these are long enough to justify abbreviating.

- **The cost function duplicates FSM logic.** `cost.go` re-implements `shouldStop()`, `chooseDirection()`, and `clearAtFloor()` — the exact same decision logic that lives in `fsm.go`. This is a coupling and maintainability hazard: if you change the stopping logic in the FSM, you must remember to change it in the cost function too, or the cost estimates become wrong. This should either share code or be explicitly documented as a deliberate simulation copy.

- **Missing `debug` import causes compilation failure.** `system_coordinator.go` calls `debug.PrintLobby()` but the `debug` package is not included in the repository. This means the code as submitted does not compile. Regardless of whether this was a debug stub, shipping code that doesn't compile is a defect.

- **`closeDoors()` is spawned as a new goroutine on every door-open event.** If events arrive quickly (e.g., rapid button presses at the current floor), multiple `closeDoors()` goroutines could be running simultaneously, each trying to close the door and choose a direction. There is no guard preventing this accumulation. A single persistent goroutine or a cancellation mechanism would be safer.

---

## Project 6d9d

**Score: 8 — Reasonable with caveats.**

### Feedback

- **Excellent type system and state modeling.** The explicit `HallCallState` enum (`Absent`, `UnconfirmedAdd`, `ConfirmedAdd`, `UnconfirmedRemove`) makes the distributed order lifecycle visible at the type level. You can read the code and understand the state machine of an order without any comments. This is self-documenting code done right.

- **Clean separation of the `orders/` module into focused files.** `action.go`, `clear.go`, `request_behaviour.go`, `state_sync.go`, `request_distribution.go`, `reassignment.go`, `lamp_state.go`, `cab_persistence.go` — each file has a single responsibility. This is well above average for a student project and makes the code navigable.

- **Process pair for fault tolerance is well-implemented.** The main/copy pattern with heartbeat monitoring is a sound approach. The atomic rename for cab call persistence (`temp file -> sync -> rename -> sync directory`) is the correct way to do crash-safe file writes. This shows the group understands fault tolerance at the systems level, not just the application level.

- **Revision-based CRDT merge with peer ID tie-breaking is solid.** The `recomputeHallCalls()` function uses Last-Writer-Wins with highest revision, and breaks ties by lower peer ID. This is deterministic and convergent. The separation between `selfHallCalls` (what I contribute) and `hallCalls` (merged view) prevents feedback loops in the gossip protocol. Good distributed systems thinking.

- **Direction encoding is inconsistent across modules.** `StatusMessage` uses `1=up, -1=down, 0=stopped` while the cost function's `targetDir` uses `0=up, 1=down` (array index form). This silent encoding difference is a bug waiting to happen. A shared direction type with explicit conversion functions would eliminate the risk.

- **`om` as a receiver name for `OrderManager` is too terse.** Go convention allows short receiver names, but `om` is ambiguous enough that it hurts readability when scanning unfamiliar code. At minimum, something like `mgr` or `orderMgr` would be clearer. The same applies to `netw` for Network.

- **The merge logic in `state_sync.go` is complex enough to warrant property-based tests.** The interaction between revisions, peer contributions, tie-breaking, and the UnconfirmedAdd/ConfirmedAdd transition has many edge cases. If any of these go wrong, orders silently disappear or persist forever. This is the kind of logic that is very hard to verify by reading alone and very easy to verify with randomized testing.

---

## Project 86ab (b-gals)

**Score: 7 — Underlying design visible but rough.**

### Feedback

- **The code does not compile as submitted.** `worldview/worldview.go` uses `config.FloorUnknown` on line 26 but does not import the `config` package — only `heis/elevtype` and `heis/order` are imported. This is a compilation error. Submitting code that does not compile is a serious defect that undermines confidence in the rest of the codebase. It suggests the code was not built from a clean state before submission.

- **Dependency on an external D-language binary for the cost function is a significant coupling risk.** The Go code shells out to `costfunction/cost_fns/hall_request_assigner/` which is a separately compiled Dihedral/D binary. This means the system cannot be built, tested, or deployed with `go build` alone. The build script (`build.sh`) has to compile D code first. If the D binary's JSON interface changes, the Go code breaks silently at runtime — there is no compile-time contract between them.

- **Barrier-based order synchronization is a good design idea.** The concept — orders advance status only when all alive peers have acknowledged — is a valid approach to distributed consensus. The status progression (Unknown -> None -> Unconfirmed -> Confirmed -> Finished -> None) gives a clear lifecycle for each order.

- **However, the order merge logic is dangerously complex.** `order_merge.go` has `MergeOrder()`, `resolveUnknown()`, `guardIllegalRollback()`, `adoptHigherStatus()`, `advanceBarrier()` — five interacting functions with subtle state transitions. The `guardIllegalRollback()` function specifically exists to prevent a bug (None <- Finished regression), which suggests this bug was encountered and patched rather than prevented by design. Code that needs guards against its own state machine is a sign the abstraction is fighting itself.

- **Duplicate IP-finding code.** `network/locallip/locallip.go` and `network/ip/ip.go` appear to implement the same functionality (connect to 8.8.8.8:53, read local address). Duplicate code is a maintenance hazard — fix a bug in one, forget the other.

- **Naming uses non-idiomatic Go conventions.** `N_FLOORS`, `N_BUTTONS` (C-style screaming snake case), `B_Idle`, `B_Moving`, `D_Down`, `D_Up` (prefix-based disambiguation instead of type-based). Go already has type safety — `Direction.Down` is unambiguous without the `D_` prefix. The module name `heis` is Norwegian for "elevator" and is opaque to anyone outside the course.

- **The SCAN algorithm implementation in `requests.go` is clean and correct.** `ChooseDirection()` and `ShouldStop()` implement the elevator SCAN algorithm clearly with well-structured if-else chains. The `collectClearedEvents()` function correctly generates finished-order events. This is one of the better parts of the codebase.

---

## Project fdd2 (final-code-49)

**Score: 7 — Underlying design visible but rough.**

### Feedback

- **`controller.go` is 955 lines and does far too many things.** It contains: the Elevator struct, the ElevatorController singleton, the state machine (8 states), the observer/pub-sub system (subscribe/unsubscribe/notify for 6 event types), hardware polling, motor control, door control, light handling, motor watchdog, order clearing, arrival handling, direction logic, and string formatting. This is a textbook cohesion violation. The observer system alone could be its own module. The state machine should be separate from the hardware I/O. The light handling is unrelated to motor control. This file needs to be broken into at least 3-4 focused modules.

- **The codebase is riddled with typos.** `obstruciton` (used as both a variable name and in `notifyObstruciton()`), `subscirbe` (in a comment on line 219), `foloowing`, `elevaotr`, `vairables`, `funcitons` — all in comments and identifiers that ship with the code. This signals a lack of care in review. Typos in identifiers are especially dangerous because they become part of the API — every caller must now misspell "obstruction" to match.

- **`time.Sleep(doorOpenTime)` in `openAndCloseDoorAtcurrentFloor()` is a blocking call.** This function is called from the `completeWaitingOrders()` goroutine, which is the main state machine loop. While it sleeps for 3 seconds, the goroutine cannot process new floor arrivals, cab orders, or hall orders. If the elevator arrives at a floor during the sleep, that event is queued but not processed until the door closes. This creates a window where the elevator is deaf to state changes.

- **Multiple sequential `GetElevatorValues()` calls create TOCTOU races.** For example, in `handleArrivalAtFloorGoingUp()`, the function calls `GetElevatorValues()` to get the state, then calls `moreOrdersAbove()` which calls `GetElevatorValues()` again internally. Between these calls, another goroutine (e.g., `pollElevatorState()`) could modify the elevator state. Each snapshot is individually consistent but the sequence of decisions is not atomic. This could lead to contradictory behavior: deciding there are no orders above based on one snapshot, then deciding there are orders below based on a different snapshot where orders have changed.

- **8 elevator states is excessive.** `Idle`, `MovingUp`, `MovingDown`, `DoorOpenHeadingUp`, `DoorOpenHeadingDown`, `DoorOpenIdle`, `Obstructed`, `MotorFailure` — compare this to other projects that use 3-4 states (Idle, Moving, DoorOpen, optionally Stopped). The `DoorOpenHeading*` states encode direction into the door-open state, which means the state machine has more transitions to get wrong. A cleaner design would separate "door status" from "travel direction" into independent variables.

- **Commented-out code left in production (lines 708-731 in `closeDoor()`).** There is a 24-line block of commented-out code labeled "potential new solution to obstructions" that was never cleaned up. This is dead weight that confuses readers: is this the intended behavior? Was it abandoned? Does the current implementation have the bug this was supposed to fix? Commented-out code should be deleted — version control exists for a reason.

- **Debug artifact: `fmt.Println("")` on line 653.** Inside `handleNewHallOrder()`, there is a bare `fmt.Println("")` that prints an empty line. This is clearly a leftover debug statement that was never removed. It is a minor symptom of the same lack of review discipline that produced the typos.

- **The observer pattern with subscribe/unsubscribe is well-intentioned but adds significant complexity.** The ElevatorController has 6 separate subscriber lists, each with subscribe, unsubscribe, and notify methods. The 1-slot buffered channels with non-blocking sends are a reasonable implementation. But this is a lot of machinery for what could be simpler goroutine-to-goroutine communication. The pattern would be more justified if there were many dynamic subscribers, but in practice the subscriber count is small and static.
