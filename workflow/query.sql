-- =========================
-- PROJECT
-- =========================

-- name: ListProjects :many
SELECT *
FROM project
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetProjectById :one
SELECT *
FROM project
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateProject :one
INSERT INTO project (
  id, name, description, workflow_id, workflow_version, status, created_at, updated_at
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateProject :one
UPDATE project
SET
  name = $2,
  description = $3,
  workflow_id = $4,
  workflow_version = $5,
  status = $6,
  updated_at = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProject :exec
UPDATE project
SET deleted_at = now()
WHERE id = $1;

-- =========================
-- WORKFLOW
-- =========================

-- name: ListWorkflows :many
SELECT * FROM workflow
WHERE deleted_at IS NULL;

-- name: GetWorkflowById :one
SELECT * FROM workflow
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflow :one
INSERT INTO workflow (
  id, name, version, start_state, active, created_at
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: UpdateWorkflow :one
UPDATE workflow
SET
  name = $2,
  version = $3,
  start_state = $4,
  active = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflow :exec
UPDATE workflow
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- WORKFLOW STATE
-- =========================

-- name: ListWorkflowStates :many
SELECT * FROM workflow_state
WHERE deleted_at IS NULL
AND workflow_id = $1;

-- name: GetWorkflowStateById :one
SELECT * FROM workflow_state
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflowState :one
INSERT INTO workflow_state (
  id, workflow_id, state_key, state_name, agent_name,
  timeout_seconds, retry_limit, is_terminal
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;

-- name: UpdateWorkflowState :one
UPDATE workflow_state
SET
  state_key = $2,
  state_name = $3,
  agent_name = $4,
  timeout_seconds = $5,
  retry_limit = $6,
  is_terminal = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflowState :exec
UPDATE workflow_state
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- WORKFLOW TRANSITION
-- =========================

-- name: ListWorkflowTransitions :many
SELECT * FROM workflow_transition
WHERE deleted_at IS NULL
AND workflow_id = $1;

-- name: GetWorkflowTransitionById :one
SELECT * FROM workflow_transition
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflowTransition :one
INSERT INTO workflow_transition (
  id, workflow_id, from_state, event_name, condition, to_state
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: UpdateWorkflowTransition :one
UPDATE workflow_transition
SET
  from_state = $2,
  event_name = $3,
  condition = $4,
  to_state = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflowTransition :exec
UPDATE workflow_transition
SET deleted_at = now()
WHERE id = $1;

-- =========================
-- WORKFLOW EXECUTION
-- =========================

-- name: ListWorkflowExecutions :many
SELECT * FROM workflow_execution
WHERE deleted_at IS NULL
AND project_id = $1;

-- name: GetWorkflowExecutionById :one
SELECT * FROM workflow_execution
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflowExecution :one
INSERT INTO workflow_execution (
  id, project_id, workflow_id, current_state, status, started_at, finished_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING *;

-- name: UpdateWorkflowExecution :one
UPDATE workflow_execution
SET
  current_state = $2,
  status = $3,
  started_at = $4,
  finished_at = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflowExecution :exec
UPDATE workflow_execution
SET deleted_at = now()
WHERE id = $1;

-- =========================
-- WORKFLOW EXECUTION STATE
-- =========================

-- name: ListExecutionStates :many
SELECT * FROM workflow_execution_state
WHERE deleted_at IS NULL
AND execution_id = $1;

-- name: GetExecutionStateById :one
SELECT * FROM workflow_execution_state
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateExecutionState :one
INSERT INTO workflow_execution_state (
  id, execution_id, state_key, status, entered_at, exited_at
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: UpdateExecutionState :one
UPDATE workflow_execution_state
SET
  state_key = $2,
  status = $3,
  entered_at = $4,
  exited_at = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteExecutionState :exec
UPDATE workflow_execution_state
SET deleted_at = now()
WHERE id = $1;

-- =========================
-- WORKFLOW EVENT
-- =========================

-- name: ListWorkflowEvents :many
SELECT * FROM workflow_event
WHERE deleted_at IS NULL
AND execution_id = $1;

-- name: GetWorkflowEventById :one
SELECT * FROM workflow_event
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflowEvent :one
INSERT INTO workflow_event (
  id, execution_id, event_name, payload, created_at
)
VALUES ($1,$2,$3,$4,$5)
RETURNING *;

-- name: DeleteWorkflowEvent :exec
UPDATE workflow_event
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- AGENT
-- =========================

-- name: ListAgents :many
SELECT * FROM agent
WHERE deleted_at IS NULL;

-- name: GetAgentById :one
SELECT * FROM agent
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateAgent :one
INSERT INTO agent (id, name, endpoint, type, active)
VALUES ($1,$2,$3,$4,$5)
RETURNING *;

-- name: UpdateAgent :one
UPDATE agent
SET
  name = $2,
  endpoint = $3,
  type = $4,
  active = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAgent :exec
UPDATE agent
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- AGENT RUN
-- =========================

-- name: ListAgentRuns :many
SELECT * FROM agent_run
WHERE deleted_at IS NULL
AND execution_state_id = $1;

-- name: GetAgentRunById :one
SELECT * FROM agent_run
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateAgentRun :one
INSERT INTO agent_run (
  id, execution_state_id, agent_id, status, input, output, started_at, finished_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;

-- name: UpdateAgentRun :one
UPDATE agent_run
SET
  status = $2,
  input = $3,
  output = $4,
  started_at = $5,
  finished_at = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAgentRun :exec
UPDATE agent_run
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- ARTIFACT
-- =========================

-- name: ListArtifacts :many
SELECT * FROM artifact
WHERE deleted_at IS NULL
AND project_id = $1;

-- name: GetArtifactById :one
SELECT * FROM artifact
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateArtifact :one
INSERT INTO artifact (
  id, project_id, type, name, latest_version, created_at
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: UpdateArtifact :one
UPDATE artifact
SET
  type = $2,
  name = $3,
  latest_version = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteArtifact :exec
UPDATE artifact
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- ARTIFACT VERSION
-- =========================

-- name: ListArtifactVersions :many
SELECT * FROM artifact_version
WHERE deleted_at IS NULL
AND artifact_id = $1;

-- name: GetArtifactVersionById :one
SELECT * FROM artifact_version
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateArtifactVersion :one
INSERT INTO artifact_version (
  id, artifact_id, version, storage_uri, checksum, metadata, created_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING *;

-- name: DeleteArtifactVersion :exec
UPDATE artifact_version
SET deleted_at = now()
WHERE id = $1;

-- =========================
-- MEMORY CHUNK
-- =========================

-- name: ListMemoryChunks :many
SELECT * FROM memory_chunk
WHERE deleted_at IS NULL
AND artifact_version_id = $1;

-- name: GetMemoryChunkById :one
SELECT * FROM memory_chunk
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateMemoryChunk :one
INSERT INTO memory_chunk (
  id, artifact_version_id, chunk_index, content, embedding, metadata
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: DeleteMemoryChunk :exec
UPDATE memory_chunk
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- APPROVAL
-- =========================

-- name: ListApprovals :many
SELECT * FROM approval
WHERE deleted_at IS NULL
AND execution_state_id = $1;

-- name: GetApprovalById :one
SELECT * FROM approval
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateApproval :one
INSERT INTO approval (
  id, execution_state_id, status, reviewer, comment, approved_at
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: UpdateApproval :one
UPDATE approval
SET
  status = $2,
  reviewer = $3,
  comment = $4,
  approved_at = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteApproval :exec
UPDATE approval
SET deleted_at = now()
WHERE id = $1;


-- =========================
-- AUDIT LOG
-- =========================

-- name: ListAuditLogs :many
SELECT * FROM audit_log
WHERE deleted_at IS NULL
AND project_id = $1
ORDER BY created_at DESC;

-- name: GetAuditLogById :one
SELECT * FROM audit_log
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateAuditLog :one
INSERT INTO audit_log (
  id, project_id, actor, action, payload, created_at
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: DeleteAuditLog :exec
UPDATE audit_log
SET deleted_at = now()
WHERE id = $1;

-- =========================
-- WORKFLOW STATE INPUT
-- =========================

-- name: ListWorkflowStateInputs :many
SELECT * FROM workflow_state_input
WHERE deleted_at IS NULL
AND workflow_state_id = $1;

-- name: GetWorkflowStateInputById :one
SELECT * FROM workflow_state_input
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflowStateInput :one
INSERT INTO workflow_state_input (
  id, workflow_state_id, artifact_type, required
)
VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: UpdateWorkflowStateInput :one
UPDATE workflow_state_input
SET
  artifact_type = $2,
  required = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflowStateInput :exec
UPDATE workflow_state_input
SET deleted_at = now()
WHERE id = $1;

-- =========================
-- WORKFLOW STATE OUTPUT
-- =========================

-- name: ListWorkflowStateOutputs :many
SELECT * FROM workflow_state_output
WHERE deleted_at IS NULL
AND workflow_state_id = $1;

-- name: GetWorkflowStateOutputById :one
SELECT * FROM workflow_state_output
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflowStateOutput :one
INSERT INTO workflow_state_output (
  id, workflow_state_id, artifact_type
)
VALUES ($1,$2,$3)
RETURNING *;

-- name: UpdateWorkflowStateOutput :one
UPDATE workflow_state_output
SET
  artifact_type = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflowStateOutput :exec
UPDATE workflow_state_output
SET deleted_at = now()
WHERE id = $1;
