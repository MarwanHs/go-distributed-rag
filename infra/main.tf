terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

provider "docker" {
  host = "unix:///var/run/docker.sock" # Use "npipe:////./pipe/docker_engine" for Windows
}

# This allows Go, Python, Kafka, and Redis to talk to each other by name.
resource "docker_network" "rag_net" {
  name = "rag_network"
}

resource "docker_image" "redis" {
  name = "redis:7-alpine"
}

resource "docker_container" "redis" {
  name  = "rag-redis"
  image = docker_image.redis.image_id
  networks_advanced {
    name = docker_network.rag_net.name
  }
  ports {
    internal = 6379
    external = 6379
  }
}

resource "docker_image" "qdrant" {
  name = "qdrant/qdrant:latest"
}

resource "docker_container" "qdrant" {
  name  = "rag-qdrant"
  image = docker_image.qdrant.image_id
  networks_advanced {
    name = docker_network.rag_net.name
  }
  ports {
    internal = 6333
    external = 6333
  }
  volumes {
    host_path      = abspath("${path.cwd}/qdrant_data")
    container_path = "/qdrant/storage"
  }
}

resource "docker_image" "kafka" {
  name = "bitnamilegacy/kafka:3.7"
}

resource "docker_container" "kafka" {
  name  = "rag-kafka"
  image = docker_image.kafka.image_id
  networks_advanced {
    name = docker_network.rag_net.name
  }
  ports {
    internal = 9092
    external = 9092
  }

  env = [
    "KAFKA_CFG_NODE_ID=0",
    "KAFKA_CFG_PROCESS_ROLES=controller,broker",
    "KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@rag-kafka:9093",
    "KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
    "KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092",
    "KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
    "KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER",
    "ALLOW_PLAINTEXT_LISTENER=yes"
  ]
}