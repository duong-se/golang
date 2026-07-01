-- migrate:up
SET TIME ZONE 'UTC';
CREATE EXTENSION vector;
-- PROJECT
CREATE TABLE project (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  workflow_id UUID,
  workflow_version INTEGER DEFAULT 1,
  status VARCHAR(30),
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_project_workflow_id ON project(workflow_id);
CREATE INDEX idx_project_deleted_at ON project(deleted_at);

-- WORKFLOW
CREATE TABLE workflow (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  version INTEGER,
  start_state VARCHAR(100),
  active BOOLEAN,
  created_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_deleted_at ON workflow(deleted_at);

-- WORKFLOW STATE
CREATE TABLE workflow_state (
  id UUID PRIMARY KEY,
  workflow_id UUID,
  state_key VARCHAR(100),
  state_name TEXT,
  agent_name TEXT,
  timeout_seconds INTEGER,
  retry_limit INTEGER,
  is_terminal BOOLEAN,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_state_workflow_id ON workflow_state(workflow_id);
CREATE INDEX idx_workflow_state_deleted_at ON workflow_state(deleted_at);

-- WORKFLOW TRANSITION
CREATE TABLE workflow_transition (
  id UUID PRIMARY KEY,
  workflow_id UUID,
  from_state VARCHAR(100),
  event_name VARCHAR(100),
  condition TEXT,
  to_state VARCHAR(100),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_transition_workflow_id ON workflow_transition(workflow_id);
CREATE INDEX idx_workflow_transition_from_state ON workflow_transition(from_state);
CREATE INDEX idx_workflow_transition_to_state ON workflow_transition(to_state);
CREATE INDEX idx_workflow_transition_deleted_at ON workflow_transition(deleted_at);

-- WORKFLOW EXECUTION
CREATE TABLE workflow_execution (
  id UUID PRIMARY KEY,
  project_id UUID,
  workflow_id UUID,
  current_state VARCHAR(100),
  status VARCHAR(30),
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_execution_project_id ON workflow_execution(project_id);
CREATE INDEX idx_workflow_execution_workflow_id ON workflow_execution(workflow_id);
CREATE INDEX idx_workflow_execution_deleted_at ON workflow_execution(deleted_at);

-- WORKFLOW EXECUTION STATE
CREATE TABLE workflow_execution_state (
  id UUID PRIMARY KEY,
  execution_id UUID,
  state_key VARCHAR(100),
  status VARCHAR(30),
  entered_at TIMESTAMPTZ,
  exited_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_wes_execution_id ON workflow_execution_state(execution_id);
CREATE INDEX idx_wes_state_key ON workflow_execution_state(state_key);
CREATE INDEX idx_wes_deleted_at ON workflow_execution_state(deleted_at);

-- WORKFLOW EVENT
CREATE TABLE workflow_event (
  id UUID PRIMARY KEY,
  execution_id UUID,
  event_name VARCHAR(100),
  payload JSONB,
  created_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_event_execution_id ON workflow_event(execution_id);
CREATE INDEX idx_workflow_event_deleted_at ON workflow_event(deleted_at);

-- AGENT
CREATE TABLE agent (
  id UUID PRIMARY KEY,
  name TEXT,
  endpoint TEXT,
  type TEXT,
  active BOOLEAN,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_agent_deleted_at ON agent(deleted_at);

-- AGENT RUN
CREATE TABLE agent_run (
  id UUID PRIMARY KEY,
  execution_state_id UUID,
  agent_id UUID,
  status VARCHAR(30),
  input JSONB,
  output JSONB,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_agent_run_execution_state_id ON agent_run(execution_state_id);
CREATE INDEX idx_agent_run_agent_id ON agent_run(agent_id);
CREATE INDEX idx_agent_run_deleted_at ON agent_run(deleted_at);

-- ARTIFACT
CREATE TABLE artifact (
  id UUID PRIMARY KEY,
  project_id UUID,
  type VARCHAR(50),
  name TEXT,
  latest_version INTEGER,
  created_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_artifact_project_id ON artifact(project_id);
CREATE INDEX idx_artifact_deleted_at ON artifact(deleted_at);

-- ARTIFACT VERSION
CREATE TABLE artifact_version (
  id UUID PRIMARY KEY,
  artifact_id UUID,
  version INTEGER,
  storage_uri TEXT,
  checksum TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_artifact_version_artifact_id ON artifact_version(artifact_id);
CREATE INDEX idx_artifact_version_deleted_at ON artifact_version(deleted_at);

-- MEMORY CHUNK
CREATE TABLE memory_chunk (
  id UUID PRIMARY KEY,
  artifact_version_id UUID,
  chunk_index INTEGER,
  content TEXT,
  embedding VECTOR(1536),
  metadata JSONB,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_memory_chunk_artifact_version_id ON memory_chunk(artifact_version_id);
CREATE INDEX idx_memory_chunk_deleted_at ON memory_chunk(deleted_at);

-- APPROVAL
CREATE TABLE approval (
  id UUID PRIMARY KEY,
  execution_state_id UUID,
  status VARCHAR(30),
  reviewer TEXT,
  comment TEXT,
  approved_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_approval_execution_state_id ON approval(execution_state_id);
CREATE INDEX idx_approval_deleted_at ON approval(deleted_at);

-- AUDIT LOG
CREATE TABLE audit_log (
  id UUID PRIMARY KEY,
  project_id UUID,
  actor TEXT,
  action TEXT,
  payload JSONB,
  created_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_audit_log_project_id ON audit_log(project_id);
CREATE INDEX idx_audit_log_deleted_at ON audit_log(deleted_at);

-- WORKFLOW STATE INPUT
CREATE TABLE workflow_state_input (
  id UUID PRIMARY KEY,
  workflow_state_id UUID NOT NULL,
  artifact_type VARCHAR(100) NOT NULL,
  required BOOLEAN DEFAULT TRUE,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_wsi_workflow_state_id ON workflow_state_input(workflow_state_id);
CREATE INDEX idx_wsi_deleted_at ON workflow_state_input(deleted_at);

-- WORKFLOW STATE OUTPUT
CREATE TABLE workflow_state_output (
  id UUID PRIMARY KEY,
  workflow_state_id UUID NOT NULL,
  artifact_type VARCHAR(100) NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_wso_workflow_state_id ON workflow_state_output(workflow_state_id);
CREATE INDEX idx_wso_deleted_at ON workflow_state_output(deleted_at);


-- migrate:down

DROP TABLE workflow_state_output;
DROP TABLE workflow_state_input;
DROP TABLE audit_log;
DROP TABLE approval;
DROP TABLE memory_chunk;
DROP TABLE artifact_version;
DROP TABLE artifact;
DROP TABLE agent_run;
DROP TABLE agent;
DROP TABLE workflow_event;
DROP TABLE workflow_execution_state;
DROP TABLE workflow_execution;
DROP TABLE workflow_transition;
DROP TABLE workflow_state;
DROP TABLE workflow;
DROP TABLE project;
