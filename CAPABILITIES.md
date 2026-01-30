# OPD Allocation Service - Capabilities Guide

This service automatically manages doctor appointments using a smart priority system. Here is what it can do for you:

## 🏥 Core Services

### 1. Smart Booking (`/book`)
**Goal**: Get a confirmed token for a specific doctor and time slot.
*   **How it works**: You tell us *Who* (Patient Name), *Priority* (Emergency/Normal), and *Where* (Doctor/Slot).
*   **Smart Outcome**:
    *   **Success**: You get a booked token immediately.
    *   **Waitlist**: If full, you are added to a waitlist.
    *   **Preemption (Emergency)**: If you are an Emergency case and the slot is full, the system might bump a lower-priority patient to the waitlist to make room for you.

### 2. View Live Schedule (`/schedule`)
**Goal**: See real-time availability.
*   **What you see**:
    *   Doctor's name.
    *   All time slots (e.g., 9-10 AM).
    *   Current remaining capacity (e.g., "Full" or "2 seats left").

### 3. Automated Cancellation & Reallocation (`/cancel`)
**Goal**: Cancel an appointment and automatically fill the gap.
*   **How it works**: When a booking is cancelled, the service **instantly** looks at the waitlist.
*   **Auto-Promotion**: The highest priority patient waiting for that slot is automatically moved from "Waiting" to "Booked".

## 🌟 Key Features

*   **Elastic Priority**: Not all patients are equal. Emergencies > VIPs > Normal bookings. The system enforces this hierarchy automatically.
*   **Waitlist Management**: No "try again later". If a slot is full, we hold your place in line.
*   **Fairness**:
    *   **Bumping Logic**: Last-In-First-Out (LIFO). We bump the person who arrived most recently, preserving early bookings.
    *   **Waitlist Logic**: First-In-First-Out (FIFO). When a seat opens, the person who has been waiting longest (for that priority level) gets it.

## 🛠 User Scenarios

| Use Case | Input | Outcome |
| :--- | :--- | :--- |
| **Normal Booking** | John (Online Booking) requests 9 AM. | **Booked** (if space exists). |
| **Full Slot** | Bob (Online Booking) requests 9 AM (Full). | **Waitlisted** automatically. |
| **Emergency** | Jane (EMERGENCY) requests 9 AM (Full). | **Booked**. System bumps Bob to Waitlist to make room. |
| **Cancellation** | John cancels his 9 AM booking. | Bob (from Waitlist) is **Promoted** to Booked instantly. |
