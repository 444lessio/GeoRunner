# GeoRunner: A High-Concurrency Spatial Simulator 🗺️

[![Go Version](https://img.shields.io/badge/Go-1.18%2B-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-18%2B-blue.svg)](https://reactjs.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

GeoRunner is a full-stack application that simulates **10,000 concurrent drivers** moving in real-time on a world map. It features a high-performance Go backend and a React (TypeScript) frontend.

The core of this project is a **thread-safe Quadtree** built from scratch to efficiently index and query thousands of rapidly moving spatial data points.

---

## 🎯 The Engineering Challenge

The goal of this project was not just to build a map app. It was to solve a classic, large-scale system design problem:

> **"How do you design a system that can track 10,000+ moving entities and, at the same time, instantly query which of them are 'nearby'?"**

A naive approach, like looping through a 10,000-item array, would be an $O(n)$ operation for every single API request. This approach does not scale. This project implements the correct, high-performance solution.

---

## 🛠️ Architecture & Tech Stack

### Backend (Go / Golang)
* **Go:** Chosen for its high performance and first-class concurrency model.
* **Goroutines:** The heart of the simulation. Each of the 10,000 drivers runs in its own lightweight Goroutine, simulating independent, concurrent movement.
* **Custom Quadtree:** The core data structure. Instead of an $O(n)$ scan, this provides an average-case spatial query complexity of **$O(\log n)$**. It was built from scratch.
* **`sync.RWMutex`:** This was critical for making the Quadtree **thread-safe**. The data structure is constantly being written to by 10,000 Goroutines while simultaneously being read from by the API. `RWMutex` allows for concurrent reads, maximizing performance while ensuring write safety (`Insert`/`Remove`).
* **Gin Framework:** A high-performance, lightweight HTTP router for exposing the `/find-nearby` API.
* **`gin-contrib/cors`:** Middleware to handle Cross-Origin Resource Sharing (CORS) for the React frontend.

### Frontend (React)
* **React (with TypeScript):** For a modern, responsive, and type-safe UI.
* **Leaflet.js:** A powerful and lightweight interactive mapping library.
* **`react-leaflet`:** React hooks and components for integrating Leaflet.
* **Polling (Real-time):** The frontend polls the Go API every 2 seconds to fetch updated driver locations, creating a live, real-time visualization.

---

## 🚀 How to Run Locally

You will need two terminals.

### 1. Run the Backend (Go)

```bash
# From the root directory (GeoRunner/)

# Install dependencies (Gin and CORS)
go mod tidy

# Run the server
go run main.go
