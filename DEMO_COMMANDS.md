# Quick Start & Demo Commands

Run these commands in your terminal to see the app in action.

## 1. Setup (Terminal 1)
Start the server:
```bash
go run cmd/service/main.go
```

## 2. API Demo (Terminal 2)

### A. Book a Normal Token
Book a slot for "John Doe" (Online Booking).
```bash
curl -X POST http://localhost:8080/book \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "doc1",
    "slot_id": "doc1_slot1",
    "patient_name": "John Doe",
    "patient_type": "ONLINE_BOOKING"
  }'
```

### B. View Schedule
See the booking you just made.
```bash
curl "http://localhost:8080/schedule?doctor_id=doc1"
```

### C. Fill the Slot (Force Capacity Limit)
Run this 3 times to fill the remaining spots in the 9-10 AM slot (Capacity is 3).
```bash
# Booking 2
curl -X POST http://localhost:8080/book -H "Content-Type: application/json" -d '{"doctor_id": "doc1", "slot_id": "doc1_slot1", "patient_name": "Grid User 2", "patient_type": "ONLINE_BOOKING"}'

# Booking 3
curl -X POST http://localhost:8080/book -H "Content-Type: application/json" -d '{"doctor_id": "doc1", "slot_id": "doc1_slot1", "patient_name": "Grid User 3", "patient_type": "ONLINE_BOOKING"}'
```

### D. Test Waitlist (Overflow)
Try to book a 4th person. Since the slot is full, they should be **Waitlisted**.
```bash
curl -X POST http://localhost:8080/book \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "doc1",
    "slot_id": "doc1_slot1",
    "patient_name": "Waitlist User",
    "patient_type": "ONLINE_BOOKING"
  }'
```
*Expected Output*: JSON response with `"status": "WAITING"`.

### E. Test Preemption (Emergency)
Book an **EMERGENCY** patient. They should bump one of the existing users.
```bash
curl -X POST http://localhost:8080/book \
  -H "Content-Type: application/json" \
  -d '{
    "doctor_id": "doc1",
    "slot_id": "doc1_slot1",
    "patient_name": "Emergency User",
    "patient_type": "EMERGENCY"
  }'
```
*Expected Output*: JSON response with `"status": "BOOKED"`. The server logs will show a preemption event.

### F. Run Stress Test Simulation
Instead of manual `curl`, run the automated script to fire many requests at once.
```bash
go run simulation/sim.go
```
