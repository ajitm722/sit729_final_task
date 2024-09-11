# SIT729 Final Task

## Project Overview

**Distinction Task: Automated Cooling System**

### Simulation

**Simulate Sensors in Two Rooms (Golang Client Code):**

- Each room has three sensors: temperature, humidity, and people detection.
- These sensors send simulated data via MQTT at regular intervals to the edge server.

**Edge Server:**

- Receives data from the sensors.
- Applies a PID controller to adjust the AC cooling based on the desired temperature.
- Sends batches of data to a central server at specified intervals.

**Central Server (EC2 Instances):**

- Receives data from the edge server.
- Displays data using Golang and HTML.
- Persists data in MongoDB Atlas.

**Concurrency (Goroutines):**

- Utilizes goroutines for efficient data collection, PID control, and server communication.

### High-Level Block Diagram

Here's a simple diagram showing the flow:

**Rooms:**

- Sensors (Temperature, Humidity, People detection)
- Data sent via MQTT to the edge server.

**Edge Server:**

- Receives sensor data.
- Applies PID controller for AC adjustment.
- Batches data and sends it to a public load balancer.

**Central Server (EC2 Instances):**

- Receives data and displays it.
- Allows users to set the desired temperature.
- Persists data to MongoDB Atlas.

![sit729_distinction](https://github.com/user-attachments/assets/399af3ea-e291-49ea-abdd-22733a8eb0b1)

**display.go** is set up on EC2 machines on AWS behind a load balancer.

## Flaws in the Existing Architecture & Proposed Improvements

**1. Single Edge Server Handling Multiple Responsibilities:**

- **Flaw:** The current architecture has a single edge server managing both HVAC actuation and sending sensor data to the cloud, introducing complexity and increasing the chance of failure.
- **Improvement:** Split the edge server into two:
  - One for HVAC actuation.
  - Another for sending data to the cloud. This separation makes the system more modular and fault-tolerant.

**2. No Circuit Breaker on Cloud Data Transmission:**

- **Flaw:** The current system lacks a mechanism to handle external service failures (e.g., central server downtime or slow response).
- **Improvement:** Implement a circuit breaker pattern on the edge server responsible for sending data to the cloud. If the central server is unresponsive, the circuit breaker will stop sending requests and retry after a specified timeout. This reduces resource wastage and avoids cascading failures.

**3. Lack of API Gateway:**

- **Flaw:** The load balancer currently routes all requests to the same server for both data display and persistence.
- **Improvement:** Add an API Gateway to route requests separately:
  - One path for displaying data.
  - Another path for persisting data into MongoDB Atlas. This separation ensures each service can be optimized independently, improving performance and scalability.

**4. Single Load Balancer for Both Data Display & Persistence:**

- **Flaw:** The same EC2 instances handle both data display and persistence, which can lead to resource contention and reduced performance.
- **Improvement:** Implement two distinct sets of EC2 instances:
  - One set for displaying data.
  - Another set for data persistence. Each set of instances will have its own load balancer and autoscaling group, improving scalability, fault tolerance, and resource allocation.

**The Circuit Breaker Pattern:**

The Circuit Breaker pattern is designed to prevent cascading failures in distributed systems by halting services from repeatedly attempting operations likely to fail. It detects faults and "opens the circuit," temporarily stopping further requests to a failing service and providing an immediate error response (Titmus, 2021).

**The Upgraded Architecture:**

![diagram-export-05-09-2024-23_23_57](https://github.com/user-attachments/assets/1408a4d9-b94f-4eed-b96d-a90999bf77ab)

---
