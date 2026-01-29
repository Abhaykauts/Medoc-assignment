# System Design - OPD Token Allocation Engine

## Architecture
The system is built as a modular Go application:
- `cmd/service`: Entry point for the HTTP service.
- `internal/core`: Domain entities (Doctor, Slot, Token).
- `internal/algo`: Core allocation logic and data structures.
- `internal/api`: HTTP handlers.

## Prioritization Logic
The system uses a priority-based allocation strategy.

### Priority Levels
1. **Emergency**: Highest priority. Can preempt existing bookings if necessary.
2. **Paid Priority (VIP)**: High priority.
3. **Follow-up**: Medium priority.
4. **Online Booking / Walk-in**: Standard priority.

### Allocation Strategy
- **Modelling**: Each slot (e.g., 9-10 AM) is a constrained resource.
- **Booking**:
    - If slot is not full: Allocate immediately.
    - If slot is full:
        - Check if request is higher priority than the lowest priority token in the slot.
        - If yes: **Preempt** (bump) the lower priority token to the next available slot or waiting list.
        - If no: Add to waiting list or suggest next slot.

## Edge Cases & Failure Handling
- **Cancellations**: Trigger an "Eager Reallocation" to pull the highest priority waiting patient into the freed slot.
- **Delays**: If a doctor is delayed, events can be effectively shifted.
- **Emergency Insertions**: Always accepted. If the schedule is physically full (time-wise), the system simulates "overbooking" or bumps the last standard patient.

## Trade-offs
- **Eager vs Lazy Reallocation**: We chose **Eager Reallocation** to maximize slot utilization. As soon as a cancellation occurs, we fill the gap.
- **In-Memory Storage**: For this assignment, state is kept in-memory. In a production system, this would require a persistent database (PostgreSQL/Redis) with locking to prevent race conditions.
