# Go Distributed RAG Pipeline

A scalable, event-driven system for **Retrieval-Augmented Generation (RAG)**.

This project demonstrates a microservices architecture where a high-performance **Go** API handles ingestion, offloading heavy compute tasks (PDF parsing, embeddings) to a **Python** worker via **Apache Kafka** and **gRPC**.

## Architecture

* **Infrastructure as Code (Terraform):** Fully automated provisioning of the local development environment (Kafka, Redis, Qdrant) using Docker providers.
* **Ingestion API (Go + Echo):** Accepts PDF uploads, validates input, and immediately returns a Job ID.
* **Service Communication (gRPC):** Facilitates low-latency, strictly typed communication between the Go Gateway and Python Worker.
* **Message Queue (Apache Kafka):** Decouples the API from heavy processing logic using the KRaft protocol.
* **State Management (Redis):** Tracks real-time job states (Pending, Processing, Completed, Failed).
* **Worker Service (Python):** Consumes jobs, generates embeddings, and stores them in a Vector DB.

## Tech Stack

* **Infrastructure:** Terraform, Docker
* **Language:** Go (1.23+), Python (3.11+)
* **Communication:** gRPC & Protocol Buffers
* **Web Framework:** Echo (Go)
* **Messaging:** Apache Kafka (KRaft Mode)
* **Cache/State:** Redis

## Quick Start (Local Dev)

Follow these steps to spin up the ingestion layer.

### Step 1: Provision Infrastructure
Ensure Docker is running, then use Terraform to spin up the required services.

    cd terraform
    terraform init
    terraform apply -auto-approve

### Step 2: Initialize Kafka Topic
Kafka requires a moment to boot. Wait ~10 seconds after the Terraform apply finishes, then run this command to create the required topic:

    docker exec -it kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
      --create \
      --topic pdf-processing \
      --bootstrap-server localhost:9092 \
      --partitions 1 \
      --replication-factor 1

### Step 3: Start the Go Gateway
From the project root, start the API server. 

    go run src/gateway/cmd/api/main.go

### Step 4: Test the Ingestion Pipeline
Upload the `test.pdf` from the project root to verify the flow.

    curl -X POST http://localhost:8080/upload -F "file=@test.pdf"

**Expected Output:**
```json
{
  "job_id": "job-173981...",
  "status": "Pending",
  "message": "File uploaded successfully"
}