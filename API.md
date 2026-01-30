# API Documentation

Base URL: `http://localhost:8080`

## 1. Book a Token
**Endpoint**: `POST /book`  
**Description**: Requests a token for a specific slot. Handles booking, waitlisting, and preemption automatically.

### Request Body
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `doctor_id` | string | Yes | ID of the doctor (e.g., `doc1`). |
| `slot_id` | string | Yes | ID of the time slot (e.g., `doc1_slot1`). |
| `patient_name` | string | Yes | Name of the patient. |
| `patient_type` | string | Yes | Priority level. Options: `EMERGENCY`, `PAID_PRIORITY`, `FOLLOW_UP`, `ONLINE_BOOKING`. |

**Example Request**:
```json
{
  "doctor_id": "doc1",
  "slot_id": "doc1_slot1",
  "patient_name": "John Doe",
  "patient_type": "ONLINE_BOOKING"
}
```

### Response
**Success (201 Created)**  
Returns the allocated token. Check status to see if Booked or Waitlisted.
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "patient_name": "John Doe",
  "type": "ONLINE_BOOKING",
  "timestamp": "2026-01-30T10:00:00Z",
  "slot_id": "doc1_slot1",
  "doctor_id": "doc1",
  "status": "BOOKED" 
}
```
*Note: Status can be `BOOKED` or `WAITING`.*

**Error Responses**:
*   `400 Bad Request`: Invalid JSON.
*   `404 Not Found`: Doctor or Slot ID does not exist.
*   `409 Conflict`: Slot is Full (and waitlist/preemption failed - rare with current logic).

---

## 2. Get Doctor Schedule
**Endpoint**: `GET /schedule`  
**Description**: Retrieves the current schedule, slots, and booked tokens for a doctor.

### Query Parameters
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `doctor_id` | string | Yes | ID of the doctor to look up. |

**Example Request**:
`GET /schedule?doctor_id=doc1`

### Response
**Success (200 OK)**
```json
{
  "id": "doc1",
  "name": "Dr. A (Cardiology)",
  "slots": [
    {
      "id": "doc1_slot1",
      "doctor_id": "doc1",
      "start_time": "...",
      "end_time": "...",
      "capacity": 3,
      "tokens": [ ... ] 
    }
  ]
}
```

**Error Responses**:
*   `400 Bad Request`: Missing `doctor_id`.
*   `404 Not Found`: Doctor ID does not exist.
