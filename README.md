# OPD Token Allocation Engine

A backend system for managing hospital OPD token allocation with elastic capacity, priority preemption, and waitlist management.

## Features
- **Hard Capacity Limits**: Enforces slot capacity checking.
- **Priority-Based Allocation**: Emergency > Paid Priority > Follow-up > Online Booking.
- **Preemption**: High-priority patients can bump lower-priority patients from full slots.
- **Waitlist Managment**:
    - **Waitlist Fallback**: Failed bookings (Slot Full) are automatically waitlisted.
    - **Waitlist Promotion**: Cancellations automatically promote the highest priority waitlisted patient.
    - **Waitlist Re-Queueing**: Bumped patients are moved to the top of the waitlist.
- **Simulation**: Includes a script to stress-test the logic with concurrent requests.

## Setup & Run

### 1. Requirements
- Go 1.18+

### 2. Start the Server
```bash
go run cmd/service/main.go
```
Server starts on `http://localhost:8080`.

### 3. Run the Simulation
In a separate terminal:
```bash
go run simulation/sim.go
```
This script simulates:
1.  **Concurrency**: Fires 10 requests at once.
2.  **Preemption**: Sends Emergency/VIP patients to a full slot to verify bumping.
3.  **Waitlist**: Verifies that overflow requests are queued (HTTP 201 Created).

## API Endpoints

### 1. Book a Token
**POST** `/book`

```json
{
  "doctor_id": "doc1",
  "slot_id": "doc1_slot1",
  "patient_name": "John Doe",
  "patient_type": "ONLINE_BOOKING" 
}
```
*   `patient_type` options: `EMERGENCY`, `PAID_PRIORITY`, `FOLLOW_UP`, `ONLINE_BOOKING`.

### 2. Get Schedule
**GET** `/schedule?doctor_id=doc1`

Returns the doctor's current state, including booked capacity.

### 3. Cancel Token
**POST** `/cancel` (Internal Logic Implemented)
*   Promotes the next best waitlisted candidate automatically.

## Design Decisions
See [DESIGN.md](DESIGN.md) for details on the architecture, trade-offs (Eager vs Lazy reallocation), and algorithms.
