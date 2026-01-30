# System Design - OPD Token Allocation Engine

## Architecture
The system follows a modular Go architecture:
- **`cmd/service`**: Main entry point. Initializes the engine and HTTP server.
- **`internal/core`**: Pure domain entities (`Doctor`, `Slot`, `Token`) and interfaces. No external dependencies.
- **`internal/algo`**: The "Business Logic" layer implementing `AllocationEngine`. 
- **`internal/api`**: HTTP Handlers that translate JSON requests to engine calls.

## Data Structures & Optimization
To ensure high performance and thread safety, the `InMemoryEngine` uses:
1.  **`doctors map[string]*Doctor`**: Direct access to doctor schedules.
2.  **`tokenMap map[string]*Slot` (Optimization)**: 
    - Maps `TokenID` -> `Slot`.
    - Allows **O(1)** lookup for cancellations (instead of iterating all doctors/slots).
3.  **`waitlist map[string][]*Token`**:
    - Maps `SlotID` -> List of waiting tokens.
    - Used for fallback bookings and preemption storage.
4.  **`sync.RWMutex`**: Protects all maps from concurrent access race conditions.

## Allocation Logic

### 1. Booking Strategy
When a booking request arrives:
1.  **Capacity Check**: If `CurrentTokens < Capacity`, allocate immediately.
2.  **Slot Full?**:
    - Identify the **Weakest Link** (lowest priority token) in the slot.
    - Tie-Breaker: If priorities are equal, we target the **Latest Arrival** (LIFO preemption).
3.  **Preemption Decision**:
    - If `NewToken.Priority > WeakestLink.Priority`:
        - **Bump** the weakest link (Status = `WAITING`).
        - Move bumped token to **Waitlist**.
        - Allocate slot to new token.
    - If `NewToken.Priority <= WeakestLink.Priority`:
        - New token is added directly to **Waitlist** (Status = `WAITING`).
        - HTTP 201 Created is returned (client should check status).

### 2. Waitlist Promotion (Reallocation)
When a confirmed token is **Cancelled**:
1.  Remove token from slot and `tokenMap`.
2.  **Scan Waitlist** for that slot.
3.  **Find Best Candidate**:
    - Highest Priority.
    - Tie-Breaker: Earliest Timestamp (FIFO).
4.  **Promote**: Move winner from Waitlist to Slot (Status `BOOKED`).

## Edge Cases
- **Concurrent Bookings**: Handled via `Expected Locking` (Mutex).
- **Waitlist Cycle**: A user can be bumped to waitlist, then promoted back if a higher priority user cancels.
- **Fairness**:
    - **Preemption**: Unfair to latest arrival (LIFO) to protect long-waiting patients.
    - **Promotion**: Fair to earliest arrival (FIFO) to reward patience.

## Trade-offs
- **In-Memory State**: Fast but not persistent. Restarting server loses data. Production would use Redis/DB.
- **Eager Reallocation**: We fill gaps immediately. A "Lazy" approach might wait for batch processing.
