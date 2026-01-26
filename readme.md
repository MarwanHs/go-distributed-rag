# Go Distributed RAG Pipeline

A scalable, event-driven system for **Retrieval-Augmented Generation (RAG)**.

This project demonstrates a microservices architecture where a high-performance **Go** API handles ingestion, offloading heavy compute tasks (PDF parsing, embeddings) to a **Python** worker via **Apache Kafka**.

## Architecture

* **Ingestion API (Go + Echo):** Accepts PDF uploads, validates input, and immediately returns a Job ID.
* **Message Queue (Apache Kafka):** Decouples the API from the heavy processing logic (running in KRaft mode, no Zookeeper).
* **State Management (Redis):** Tracks the real-time status of jobs (Pending, Processing, Completed, Failed).
* **Worker Service (Python):** Consumes jobs, generates embeddings, and stores them in a Vector DB.
* **Vector DB (Qdrant):** Stores semantic embeddings for fast retrieval.
* **Inference (Ollama):** Local LLM for generating answers.

## Tech Stack

* **Language:** Go (1.23+), Python (3.11+)
* **Web Framework:** Echo (Go)
* **Messaging:** Apache Kafka (KRaft Mode)
* **Cache/State:** Redis
* **Vector Database:** Qdrant
* **LLM Runtime:** Ollama (Llama 3)
* **Infrastructure:** Docker & Docker Compose

## Quick Start (Local Dev)

Follow these steps to spin up the ingestion layer (Phase 1).

### Step 1: Start the Infrastructure
Start the Kafka broker and the Go API server using Docker Compose.

    docker-compose up -d --build

### Step 2: Initialize the Kafka Topic - ***skip for now, resolved 
Wait about 10 seconds for Kafka to fully boot, then run this **one-time command** to create the required topic.
*(Note: We run this manually to avoid complex initialization scripts during development.)*

    docker-compose exec kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
      --bootstrap-server localhost:9092 \
      --create \
      --topic pdf-processing \
      --partitions 1 \
      --replication-factor 1

### Step 3: Test the Ingestion Pipeline
Upload a sample PDF to verify the API, Redis, and Kafka producer are all talking to each other.

    curl -X POST http://localhost:8080/upload -F "file=@test.pdf"

**Expected Output:**

    {
      "job_id": "job-173981...",
      "status": "Pending",
      "message": "File uploaded successfully"
    }

---

## Troubleshooting

**Kafka crashing or "Cluster ID" errors?**
If you change network configurations or restart frequently, Kafka's data volume might get out of sync. Perform a **Hard Reset** to wipe the state and start fresh:

    # Stop containers and delete data volumes
    docker-compose down -v

    # Restart services
    docker-compose up -d

*(Remember to re-run the "Initialize Kafka Topic" command in Step 2 after a hard reset!)*

---

## API Design (Endpoints)

### Upload a Document
`POST /upload`
* **Body (Multipart Form):** `file` (PDF), `user_id` (Text)
* **Behavior:** Saves file to disk, creates Job ID, pushes event to Kafka topic `pdf-processing`.

### Check Job Status
`GET /status/:job_id`
* **Behavior:** Checks Redis for the current state of the processing job.

## Roadmap

* [x] **Phase 1:** Core Ingestion API (Go) with Kafka Producer & Redis State.
* [ ] **Phase 2:** Python Worker (Kafka Consumer + PDF Text Extraction).
* [ ] **Phase 3:** Embedding Generation (Sentence-Transformers) & Qdrant Storage.
* [ ] **Phase 4:** Retrieval & Chat Endpoint (Ollama Integration).

## License
MIT