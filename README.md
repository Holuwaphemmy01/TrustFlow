# 🛡️ TrustFlow: Transparency & Control for AI Agents

TrustFlow is a **security orchestration layer** designed to sit between AI Agents and the Blockchain. It transforms "Black Box" AI decisions into "Glass Box" verifiable workflows, ensuring safety, compliance, and explainability.

---

## 🏗️ Architecture

TrustFlow acts as a middleware that **Simulates**, **Validates**, and **Logs** every transaction before execution.

```mermaid
graph TD
    User[AI Agent / User] -->|1. Submit Intent| API[TrustFlow API :8081]
    
    subgraph Core System
        API --> Orchestrator[Orchestrator]
        Orchestrator -->|2. Dry Run| Simulator[Simulator]
        Simulator -->|Check State| Chain[Cronos zkEVM]
        
        Orchestrator -->|3. Audit Log| DB[(SQLite Database)]
        Orchestrator -->|4. Execute| Executor[Executor]
    end
    
    Executor -->|5. Final Tx| Chain
    
    Dashboard[Streamlit Dashboard :8501] -->|6. Monitor| API
```

---

## 🚀 Features

### 1. **Proactive Trust (The Simulator)**
Before any funds move, TrustFlow runs a "Dry Run" simulation.
- **Balance Checks**: Prevents "Insufficient Funds" errors before they hit the chain.
- **Contract Scanning**: Detects potential reverts or malicious patterns.
- **Budget Enforcer**: (Mock) Ensures transactions stay within daily limits (e.g., $100).

### 2. **Fail-Safe Orchestration**
- **Multi-Step Workflows**: Handles complex sequences (e.g., `Approve` -> `Transfer`).
- **Atomic Halting**: If Step 1 fails, the workflow **stops immediately**. No partial states or stuck funds.

### 3. **The "Glass Box" Dashboard**
A React-style Streamlit UI that provides deep observability:
- **🚦 Traffic Light Status**: Green (Safe), Red (Blocked).
- **🛑 Human-Readable Errors**: Translates `execution reverted` into *"PREVENTED: Contract Rejection"*.
- **📜 Audit Trace**: Side-by-side view of the **Raw Intent (JSON)** vs. **Execution Result**.

### 4. **Persistent Audit Log**
Every action is recorded in a local SQLite database (`trustflow.db`), ensuring a permanent, queryable history of all AI actions.

---

## ⚡ Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.25+ (optional, for local dev)

### One-Click Deploy
Run the entire stack (Server + Dashboard + Database) with a single command:

```bash
docker compose up --build
```

- **API Server**: `http://localhost:8081`
- **Dashboard**: `http://localhost:8501`

---

## 🔌 API Reference

### 1. Submit Intent
**POST** `/intents`

```json
{
  "steps": [
    {
      "action": "payment",
      "params": {
        "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
        "amount": "0.1",
        "token": "TCRO"
      }
    }
  ]
}
```

### 2. Check Status (Polling)
**GET** `/status/:id`

Returns the real-time state of the intent, including simulation results and execution steps.

---

## 📂 Project Structure

```
TrustFlow/
├── dashboard/          # Streamlit UI
│   ├── app.py          # Frontend Logic (Visualizer)
│   └── Dockerfile      # Python Environment
├── src/
│   ├── cmd/server/     # Go Entrypoint
│   ├── internal/
│   │   ├── orchestrator/ # Core Logic (Fail-Safe)
│   │   ├── simulator/    # Safety Checks
│   │   └── storage/      # SQLite Implementation
│   └── pkg/types/      # Shared Data Models
├── docker-compose.yml  # Stack Orchestration
└── Dockerfile          # Go Server Build (Multi-Stage)
```

---

## 🛠️ Troubleshooting

- **Database Locked?**: The SQLite DB is mounted as a volume. If you can't open it locally, it might be in use by the container.
- **Build Errors?**: Ensure you are using the latest Docker version. The build uses `golang:1.25-alpine` to ensure compatibility.

---

*Built with ❤️ for the AI Agent Economy.*
