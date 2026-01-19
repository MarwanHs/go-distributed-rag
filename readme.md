# 🚀 Go Distributed RAG Pipeline

A scalable, event-driven system for **Retrieval-Augmented Generation (RAG)**.

This project demonstrates a microservices architecture where a high-performance **Go** API handles ingestion, offloading heavy compute tasks (PDF parsing, embeddings) to a **Python** worker via **Apache Kafka**.

## 🏗 Architecture

1.  **Ingestion API (Go + Echo):** Accepts PDF uploads, validates input, and immediately returns a Job ID.
2.  **Message Queue (Apache Kafka):** Decouples the API from the heavy processing logic.
3.  **State Management (Redis):** Tracks the real-time status of jobs (Pending, Processing, Completed, Failed).
4.  **Worker Service (Python):** Consumes jobs, generates embeddings, and stores them in a Vector DB.
5.  **Vector DB (Qdrant):** Stores semantic embeddings for fast retrieval.
6.  **Inference (Ollama):** Local LLM for generating answers.

## 🛠 Tech Stack

* **Language:** Go (1.23+), Python (3.11+)
* **Web Framework:** Echo (Go)
* **Messaging:** Apache Kafka & Zookeeper
* **Cache/State:** Redis
* **Vector Database:** Qdrant
* **LLM Runtime:** Ollama (Llama 3)
* **Infrastructure:** Docker & Docker Compose

## 📡 API Design (Endpoints)

### 1. Upload a Document
`POST /upload`
* **Body (Multipart Form):** `file` (PDF), `user_id` (Text)
* **Behavior:** Saves file to disk, creates Job ID, pushes event to Kafka.

### 2. Check Job Status
`GET /status/:job_id`
* **Behavior:** Checks Redis for the current state of the processing job.

## 🗺 Roadmap

- [x] **Phase 1:** Core Ingestion API (Go) with Kafka Producer & Redis State.
- [ ] **Phase 2:** Python Worker (Kafka Consumer + PDF Text Extraction).
- [ ] **Phase 3:** Embedding Generation (Sentence-Transformers) & Qdrant Storage.
- [ ] **Phase 4:** Retrieval & Chat Endpoint (Ollama Integration).

## 📝 License
MIT