-- =========================
-- PROJECT
-- =========================

-- name: CreateProject :one
INSERT INTO project (
    id, name, description, workflow_id, workflow_version, status, created_at, updated_at, deleted_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW(), NULL
)
RETURNING *;

-- name: GetProjectById :one
SELECT * FROM project
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProjects :many
SELECT * FROM project
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE project
SET
    name = $2,
    description = $3,
    workflow_id = $4,
    workflow_version = $5,
    status = $6,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProject :exec
UPDATE project
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- WORKFLOW
-- =========================

-- name: CreateWorkflow :one
INSERT INTO workflow (
    id, name, version, start_state, active, created_at, deleted_at
)
VALUES (
    $1, $2, $3, $4, $5, NOW(), NULL
)
RETURNING *;

-- name: GetWorkflowById :one
SELECT * FROM workflow
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflows :many
SELECT * FROM workflow
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

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
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- WORKFLOW STATE
-- =========================

-- name: CreateWorkflowState :one
INSERT INTO workflow_state (
    id, workflow_id, state_key, state_name, agent_name,
    timeout_seconds, retry_limit, is_terminal, deleted_at
)
VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8,NULL
)
RETURNING *;

-- name: GetWorkflowStateById :one
SELECT * FROM workflow_state
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowStates :many
SELECT * FROM workflow_state
WHERE workflow_id = $1 AND deleted_at IS NULL;

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
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- WORKFLOW TRANSITION
-- =========================

-- name: CreateWorkflowTransition :one
INSERT INTO workflow_transition (
    id, workflow_id, from_state, event_name, condition, to_state, deleted_at
)
VALUES ($1,$2,$3,$4,$5,$6,NULL)
RETURNING *;

-- name: GetWorkflowTransitionById :one
SELECT * FROM workflow_transition
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowTransitions :many
SELECT * FROM workflow_transition
WHERE workflow_id = $1 AND deleted_at IS NULL;

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
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- WORKFLOW EXECUTION
-- =========================

-- name: CreateWorkflowExecution :one
INSERT INTO workflow_execution (
    id, project_id, workflow_id, current_state, status,
    started_at, finished_at, deleted_at
)
VALUES ($1,$2,$3,$4,$5,NOW(),NULL,NULL)
RETURNING *;

-- name: GetWorkflowExecutionById :one
SELECT * FROM workflow_execution
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowExecutions :many
SELECT * FROM workflow_execution
WHERE project_id = $1 AND deleted_at IS NULL;

-- name: UpdateWorkflowExecution :one
UPDATE workflow_execution
SET
    current_state = $2,
    status = $3,
    finished_at = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflowExecution :exec
UPDATE workflow_execution
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- WORKFLOW EXECUTION STATE
-- =========================

-- name: CreateWorkflowExecutionState :one
INSERT INTO workflow_execution_state (
    id, execution_id, state_key, status, entered_at, exited_at, deleted_at
)
VALUES ($1,$2,$3,$4,NOW(),NULL,NULL)
RETURNING *;

-- name: GetWorkflowExecutionStateById :one
SELECT * FROM workflow_execution_state
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowExecutionStates :many
SELECT * FROM workflow_execution_state
WHERE execution_id = $1 AND deleted_at IS NULL;

-- name: UpdateWorkflowExecutionState :one
UPDATE workflow_execution_state
SET
    status = $2,
    exited_at = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteWorkflowExecutionState :exec
UPDATE workflow_execution_state
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- WORKFLOW EVENT
-- =========================

-- name: CreateWorkflowEvent :one
INSERT INTO workflow_event (
    id, execution_id, event_name, payload, created_at, deleted_at
)
VALUES ($1,$2,$3,$4,NOW(),NULL)
RETURNING *;

-- name: ListWorkflowEvents :many
SELECT * FROM workflow_event
WHERE execution_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;


-- =========================
-- AGENT
-- =========================

-- name: CreateAgent :one
INSERT INTO agent (
    id, name, endpoint, type, active, deleted_at
)
VALUES ($1,$2,$3,$4,$5,NULL)
RETURNING *;

-- name: GetAgentById :one
SELECT * FROM agent
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAgents :many
SELECT * FROM agent
WHERE deleted_at IS NULL;

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
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- AGENT RUN
-- =========================

-- name: CreateAgentRun :one
INSERT INTO agent_run (
    id, execution_state_id, agent_id, status, input, output,
    started_at, finished_at, deleted_at
)
VALUES ($1,$2,$3,$4,$5,$6,NOW(),NULL,NULL)
RETURNING *;

-- name: GetAgentRunById :one
SELECT * FROM agent_run
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAgentRuns :many
SELECT * FROM agent_run
WHERE execution_state_id = $1 AND deleted_at IS NULL;

-- name: UpdateAgentRun :one
UPDATE agent_run
SET
    status = $2,
    output = $3,
    finished_at = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAgentRun :exec
UPDATE agent_run
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- ARTIFACT
-- =========================

-- name: CreateArtifact :one
INSERT INTO artifact (
    id, project_id, type, name, latest_version, created_at, deleted_at
)
VALUES ($1,$2,$3,$4,$5,NOW(),NULL)
RETURNING *;

-- name: GetArtifactById :one
SELECT * FROM artifact
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListArtifacts :many
SELECT * FROM artifact
WHERE project_id = $1 AND deleted_at IS NULL;

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
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- ARTIFACT VERSION
-- =========================

-- name: CreateArtifactVersion :one
INSERT INTO artifact_version (
    id, artifact_id, version, storage_uri, checksum, metadata, created_at, deleted_at
)
VALUES ($1,$2,$3,$4,$5,$6,NOW(),NULL)
RETURNING *;

-- name: ListArtifactVersions :many
SELECT * FROM artifact_version
WHERE artifact_id = $1 AND deleted_at IS NULL;


-- =========================
-- MEMORY CHUNK
-- =========================

-- name: CreateMemoryChunk :one
INSERT INTO memory_chunk (
    id, artifact_version_id, chunk_index, content, embedding, metadata, deleted_at
)
VALUES ($1,$2,$3,$4,$5,$6,NULL)
RETURNING *;

-- name: ListMemoryChunks :many
SELECT * FROM memory_chunk
WHERE artifact_version_id = $1 AND deleted_at IS NULL;


-- =========================
-- APPROVAL
-- =========================

-- name: CreateApproval :one
INSERT INTO approval (
    id, execution_state_id, status, reviewer, comment, approved_at, deleted_at
)
VALUES ($1,$2,$3,$4,$5,$6,NULL)
RETURNING *;

-- name: GetApprovalById :one
SELECT * FROM approval
WHERE id = $1 AND deleted_at IS NULL;

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
SET deleted_at = NOW()
WHERE id = $1;


-- =========================
-- AUDIT LOG
-- =========================

-- name: CreateAuditLog :one
INSERT INTO audit_log (
    id, project_id, actor, action, payload, created_at, deleted_at
)
VALUES ($1,$2,$3,$4,$5,NOW(),NULL)
RETURNING *;

-- name: ListAuditLogs :many
SELECT * FROM audit_log
WHERE project_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;


-- =========================
-- WORKFLOW STATE INPUT
-- =========================

-- name: CreateWorkflowStateInput :one
INSERT INTO workflow_state_input (
    id, workflow_state_id, artifact_type, required, deleted_at
)
VALUES ($1,$2,$3,$4,NULL)
RETURNING *;

-- name: ListWorkflowStateInputs :many
SELECT * FROM workflow_state_input
WHERE workflow_state_id = $1 AND deleted_at IS NULL;


-- =========================
-- WORKFLOW STATE OUTPUT
-- =========================

-- name: CreateWorkflowStateOutput :one
INSERT INTO workflow_state_output (
    id, workflow_state_id, artifact_type, deleted_at
)
VALUES ($1,$2,$3,NULL)
RETURNING *;

-- name: ListWorkflowStateOutputs :many
SELECT * FROM workflow_state_output
WHERE workflow_state_id = $1 AND deleted_at IS NULL;
