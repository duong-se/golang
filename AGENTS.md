# Agent Instructions

## Developer Commands
Run these from the `agent-orchestrator/` directory:
- Generate SQL: `make sqlc`
- Manage Migrations: `make dbmate`
- Install Dependencies: `make install`

## Database
- Uses `pgvector/pgvector:pg16`.
- Default connection: `postgres://postgres:postgres@db:5432/workflow?sslmode=disable`
- Managed via `dbmate`.

## Project Structure
- `agent-orchestrator/`: Main application logic.
- `agent-orchestrator/internal/agent/`: Core agent implementations.
