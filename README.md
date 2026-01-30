# 🏥 OPD Token Allocation Engine

A high-performance, concurrent backend system for managing hospital OPD bookings. It supports **Hard Capacity Limits**, **Elastic Priority Management**, and **Dynamic Waitlist Reallocation**.

Built with **Go (Golang)**.

---

## ✨ Key Features

### 1. Priority & Preemption
The system assigns prioritization based on `PatientType`. If a slot is full, a higher-priority patient can **"bump"** (preempt) a lower-priority patient to the waitlist.
*   🚑 **Emergency** (Priority 100): Can bump anyone.
*   💎 **Paid Priority** (Priority 80): Can bump Follow-up/Standard.
*   🔄 **Follow-up** (Priority 60): Can bump Standard.
*   👤 **Online Booking** (Priority 40): Standard ticket.

### 2. Smart Waitlist Management
*   **Automatic Fallback**: If a slot is full and preemption isn't possible, the user is added to a valid **Waitlist** (Status: `WAITING`).
*   **Auto-Promotion**: When a slot opens (Cancellation/No-Show), the **highest-priority** person on the waitlist is automatically promoted to `BOOKED`.
*   **Fairness**:
    *   **Bumping**: LIFO (Last-In-First-Out) - The person who booked *most recently* is bumped to protect early bookers.
    *   **Promotion**: FIFO (First-In-First-Out) - The person who has been waiting longest gets the seat.

### 3. High Performance
*   **O(1) Lookups**: Uses an internal `TokenMap` to find and cancel tokens instantly without scanning slots.
*   **Thread Safety**: Fully thread-safe using `sync.RWMutex` to handle thousands of concurrent requests.

---

## 📂 Project Structure

```text
/medoc-assignment
├── cmd/service/           # Application Entry Point
│   └── main.go            # Server initialization
├── internal/
│   ├── algo/              # Core Logic (The "Brain")
│   ├── api/               # HTTP Handlers (The "Face")
│   └── core/              # Domain Models (Types & Interfaces)
├── simulation/            # Stress Testing Script
│   └── sim.go             # Simulates 3 doctors + concurrent load
├── API.md                 # Detailed API Documentation
└── DESIGN.md              # Architectural decisions
```

---

## 🚀 Getting Started

### Prerequisites
*   Go 1.18 or higher.

### 1. Start the Server
```bash
go run cmd/service/main.go
```
*Server starts on port `8080`.*

### 2. Run the Stress Test (Simulation)
Open a new terminal and run the automated simulation. It creates 3 doctors and fires 20+ concurrent events to test locking and preemption.
```bash
go run simulation/sim.go
```

### 3. Manual Testing
See [DEMO_COMMANDS.md](DEMO_COMMANDS.md) for copy-pasteable `curl` commands.

---

## 🔌 API Summary
*See [API.md](API.md) for full schema details.*

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/book` | Book a token. Handles Waitlisting/Preemption automatically. |
| `POST` | `/cancel` | Cancel a token. Triggers auto-promotion from waitlist. |
| `GET` | `/schedule` | View doctor's slots and current capacity. |

---

## 🧠 Design Decisions
Please refer to **[DESIGN.md](DESIGN.md)** for a deep dive into:
*   Why we chose **Eager Reallocation**.
*   In-Memory vs Database trade-offs.
*   Edge cases handled (No-shows, Delays).
