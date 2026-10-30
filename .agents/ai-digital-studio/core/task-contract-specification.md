---
id: task-contract-specification
name: Task Contract Specification
version: 1.1.0
type: core-specification
runtime: portable
status: draft
language: en
---

# Task Contract Specification

## 1. Purpose

This specification defines the canonical structure used to create, assign, execute, review, and close tasks across the studio.

The task contract is the shared operational language between the Studio Coordinator, specialist roles, workflows, and any runtime adapter. It makes assignments explicit, supports traceable handoffs, and prevents completion from being based on unsupported claims.

This specification defines a **logical contract**, not a required storage format. A runtime may represent a task as Markdown, YAML, JSON, a database record, or another format, provided that the required semantics are preserved.

## 2. Design Principles

1. **One accountable owner:** Every active task has exactly one owner responsible for its outcome.
2. **Outcome over activity:** A task describes a result to produce, not merely an action to perform.
3. **Verifiable acceptance:** Every task has criteria that can be checked.
4. **Explicit dependencies:** Work must not begin while required prerequisites are unresolved.
5. **Least privilege:** A task receives only the access and authority needed to complete it.
6. **Evidence-based status:** Status updates describe observable progress and link to available evidence.
7. **Portable representation:** The contract does not depend on one agent framework or runtime.
8. **Controlled scope:** Material changes to deliverables, constraints, or acceptance criteria require coordinator review.
9. **Safe concurrency:** Parallel tasks must not create conflicting edits or unsafe side effects.
10. **Honest completion:** A task cannot be marked done solely because an agent says it is done.

## 3. Contract Structure

A task contract contains the following sections:

- Identity
- Assignment
- Objective and scope
- Inputs and context
- Deliverables
- Acceptance criteria
- Dependencies
- Constraints and permissions
- Review requirements
- Execution state
- Evidence and handoff
- Change history

A runtime may add fields, but it must not redefine the meaning of the required fields.

## 4. Canonical Fields

### 4.1 Identity

| Field | Required | Description |
|---|---|---|
| `task_id` | Yes | Unique, stable identifier within the studio's task system. |
| `project_id` | Yes | Identifier of the project that owns the task. |
| `title` | Yes | Short, specific task title. |
| `created_at` | Recommended | Timestamp when the task was created, in an agreed standard such as ISO 8601 UTC. |
| `updated_at` | Recommended | Timestamp of the latest recorded change. |
| `schema_version` | Yes | Version of this task-contract schema used by the record. |

### 4.2 Assignment

| Field | Required | Description |
|---|---|---|
| `owner` | Yes | One accountable role or agent identifier. |
| `contributors` | No | Additional roles or agents contributing to the task. |
| `required_capabilities` | Recommended | Skills, expertise, or tools required to complete the task. |
| `assigned_by` | Recommended | Coordinator or authorized entity that assigned the task. |
| `priority` | Yes | Relative urgency: `critical`, `high`, `normal`, or `low`. |
| `due_at` | No | Deadline in an agreed timestamp format. Must not be invented if not provided. |

The owner is accountable for the deliverable even when contributors perform parts of the work. The owner must not be left ambiguous between multiple agents.

### 4.3 Objective and Scope

| Field | Required | Description |
|---|---|---|
| `objective` | Yes | The outcome this task must achieve. |
| `in_scope` | Recommended | Work explicitly included in the task. |
| `out_of_scope` | Recommended | Work explicitly excluded. |
| `assumptions` | No | Assumptions currently being used. Each material assumption should be validated or surfaced. |
| `success_definition` | Yes | Concise statement of what success means, consistent with acceptance criteria. |

The objective must be specific enough that the owner can understand the intended outcome without reconstructing the entire conversation.

### 4.4 Inputs and Context

| Field | Required | Description |
|---|---|---|
| `inputs` | Yes | Source artifacts, briefs, assets, data, or references required for the task. Use an empty list when none are required. |
| `context_summary` | Yes | Relevant background and decisions needed to execute the task. |
| `source_of_truth` | Recommended | List of one or more canonical documents, files, or records to consult when sources conflict. |
| `upstream_decisions` | No | Decisions or approvals that constrain this task. |
| `open_questions` | No | Unresolved questions that may affect execution. |

Context should be sufficient but not needlessly duplicated. Prefer references to canonical artifacts over copying large documents into every task.

### 4.5 Deliverables

| Field | Required | Description |
|---|---|---|
| `deliverables` | Yes | One or more named outputs. Use an empty list only for a coordinator-approved non-artifact task. |
| `output_format` | Recommended | Default format, such as `.md`, `.fig`, source code, report, or structured data, for deliverables that do not state their own `format`. |
| `destination` | No | Default output location for deliverables that do not state their own `destination`. |
| `artifact_owner` | Recommended | Role responsible for maintaining the resulting artifact after handoff. |

Every deliverable should be identifiable and reviewable. If a deliverable is a file, record its actual path or stable reference only after it exists. Never fabricate paths, links, or artifact identifiers.

### 4.6 Acceptance Criteria

| Field | Required | Description |
|---|---|---|
| `acceptance_criteria` | Yes | List of items, each an object with `id`, `criterion`, and `verification_method`. Conditions must be observable. |
| `verification_method` | Recommended | How a criterion will be checked. Recorded per criterion inside each `acceptance_criteria` item. |
| `reviewer` | No | Role or person authorized to review the output. |
| `approval_required` | Yes | Boolean indicating whether explicit approval is required before closure or release. |

Acceptance criteria must be testable or assessable. Avoid vague criteria such as “make it good,” “ensure quality,” or “finish properly” unless they are expanded into observable checks.

For creative work, criteria may include strategic alignment, audience fit, visual consistency, message clarity, originality, accessibility, and required export formats. For software, criteria may include functional behavior, tests, responsiveness, error handling, accessibility, security, and documented limitations.

### 4.7 Dependencies

| Field | Required | Description |
|---|---|---|
| `dependencies` | Yes | List of task IDs that must reach the required state before execution or acceptance. Use an empty list when none exist. |
| `dependency_policy` | Recommended | Conditions under which dependent work may begin. |
| `blocked_by` | Yes | Current unresolved blockers; use an empty list when none exist. |

A dependency is not the same as a contextual reference. Dependencies represent prerequisites that can prevent safe or valid progress.

A task must not be marked `READY` while a mandatory dependency remains unresolved, unless the coordinator explicitly authorizes a limited, reversible preparatory activity.

### 4.8 Constraints and Permissions

| Field | Required | Description |
|---|---|---|
| `constraints` | Yes | Budget, scope, technical, brand, timing, policy, or other boundaries. Use an empty list when none are known. |
| `permissions` | Yes | Explicitly allowed actions and access boundaries. |
| `restricted_actions` | Yes | Actions prohibited without additional authorization. |
| `external_side_effects` | Recommended | Whether the task may affect external systems, publish content, contact people, incur costs, or change production data. |
| `sensitivity` | Recommended | Classification such as `public`, `internal`, `confidential`, or `restricted`, where relevant. |

Permissions must not be inferred from technical capability. The fact that an agent can perform an action does not mean the task authorizes it.

Actions that normally require explicit authorization include publishing or sending content, making purchases, changing production systems, deleting or overwriting important data, accessing sensitive information, committing to client terms, and expanding scope or budget.

### 4.9 Review Requirements

| Field | Required | Description |
|---|---|---|
| `review_requirements` | Yes | Required checks, reviewers, or quality gates. Use an empty list only when no additional review is needed. |
| `review_status` | Yes | One of `not_required`, `pending`, `in_review`, `changes_requested`, `approved`, or `rejected`. |
| `review_feedback` | No | Specific findings, defects, or requested changes. |
| `quality_gate` | Recommended | Gate that must pass before handoff or closure. |

A task may be technically complete but still remain in review. Completion of work and acceptance of work are distinct states.

### 4.10 Execution State

| Field | Required | Description |
|---|---|---|
| `status` | Yes | Current task state from the canonical lifecycle. |
| `progress_summary` | Yes | Short, factual description of current progress. |
| `started_at` | No | Timestamp when execution began. |
| `last_status_update_at` | Recommended | Timestamp of the most recent status update. |
| `attempt_count` | Recommended | Number of execution attempts, including retries. |
| `failure_reason` | No | Clear explanation when the task fails. |
| `recovery_plan` | No | Next steps for retry, reassignment, or escalation. |

Canonical task states:

- `PENDING` — Created but not yet eligible to start.
- `READY` — Required inputs and dependencies are available.
- `IN_PROGRESS` — Work is actively being performed.
- `BLOCKED` — Progress is prevented by an unresolved issue.
- `REVIEW` — Deliverables are submitted for review.
- `DONE` — Acceptance criteria are met and required reviews or approvals are complete.
- `FAILED` — The current attempt failed or its result was rejected; a recovery decision is required. FAILED is not terminal.
- `CANCELLED` — The task has been explicitly cancelled by an authorized decision-maker.

A runtime may define additional internal states, but it must map them consistently to these canonical states.

The authoritative definitions, valid transitions, review-decision mapping, and `review_status` alignment rules are in `core/task-state-model.md`. If this section differs from that document, that document prevails.

### 4.11 Evidence and Handoff

| Field | Required | Description |
|---|---|---|
| `completion_evidence` | Yes | Evidence supporting the reported outcome. Use an empty list before work begins. |
| `artifact_references` | Yes | Stable references to created or updated artifacts. Use an empty list when no artifacts exist. |
| `verification_results` | Recommended | Checks performed and their outcomes. |
| `limitations` | Yes | Known gaps, unverified claims, or limitations. Use an empty list when none are known. |
| `handoff_summary` | No | What the next owner or coordinator needs to know. |
| `next_actions` | No | Remaining actions, if any. |
| `completed_at` | No | Timestamp when the task entered `DONE`. |

Evidence may include file references, test results, screenshots, validation notes, review decisions, or other verifiable outputs. Evidence must be truthful and must not expose secrets or unauthorized personal data.

### 4.12 Change History

| Field | Required | Description |
|---|---|---|
| `change_log` | Recommended | Ordered record of meaningful changes to scope, owner, status, acceptance criteria, permissions, or dependencies. |
| `revision` | Recommended | Monotonically increasing revision number for the task contract. |
| `approval_records` | Recommended | References to material approvals and the decisions they authorize. |

Material contract changes should be traceable. A runtime that does not support a change log must preserve equivalent audit information elsewhere.

### 4.13 Lightweight Profile

For tasks that qualify for lightweight intake (single deliverable, no
external side effects, no new costs, low risk), a runtime may use the
reduced profile. Required fields: `task_id`, `project_id`, `schema_version`,
`title`, `owner`, `priority`, `objective`, `success_definition`,
`deliverables`, `acceptance_criteria`, `approval_required`, `status`,
`progress_summary`, `limitations`, `completion_evidence`,
`artifact_references`, `restricted_actions`.

Omitted fields take safe defaults: no dependencies, no blockers, no
external side effects, no extra permissions beyond the assigned workspace.
If the task grows in scope or risk, convert it to the full contract and
record the conversion in the change log.

## 5. Canonical YAML Example

The following is an illustrative record. IDs, paths, timestamps, and results must be replaced with real values when the task is created or executed.

```yaml
schema_version: "1.1.0"

task_id: "WEB-014"
project_id: "PORTFOLIO-001"
title: "Implement responsive project showcase section"

created_at: null
updated_at: null
revision: 1

owner: "software-engineer"
contributors:
  - "ui-ux-designer"
assigned_by: "studio-coordinator"
required_capabilities:
  - "responsive frontend implementation"
  - "semantic HTML and accessibility"
priority: "normal"
due_at: null

objective: >
  Implement the approved project showcase section so that users can
  browse selected projects on desktop and mobile devices.

in_scope:
  - "Implement the approved section layout"
  - "Support the agreed responsive breakpoints"
  - "Use the supplied project content and assets"

out_of_scope:
  - "Redesigning the overall site architecture"
  - "Publishing the site to production"

assumptions: []
success_definition: >
  The section matches the approved design direction, behaves correctly
  at supported viewport sizes, and passes the defined review checks.

inputs:
  - "Approved UI design reference"
  - "Project content and asset manifest"

context_summary: >
  The section is part of a portfolio website. Preserve the approved
  visual direction and avoid introducing unrelated design changes.

source_of_truth:
  - "project/design/project-showcase-reference.md"

upstream_decisions: []
open_questions: []

deliverables:
  - name: "Responsive project showcase section"
    format: "source code"
    destination: "project source tree"

output_format: "source code"
destination: null
artifact_owner: "software-engineer"

acceptance_criteria:
  - id: "AC-01"
    criterion: "All supplied project entries render with the required content."
    verification_method: "Manual content review"
  - id: "AC-02"
    criterion: "The layout works at all agreed viewport sizes without horizontal overflow."
    verification_method: "Responsive viewport checks"
  - id: "AC-03"
    criterion: "Interactive elements are keyboard accessible and have clear focus states."
    verification_method: "Keyboard accessibility review"
  - id: "AC-04"
    criterion: "The implementation introduces no known console errors in the tested flow."
    verification_method: "Browser console check"

reviewer: "qa-engineer"
approval_required: true

dependencies:
  - "WEB-009"
dependency_policy: "Required design and content tasks must be accepted before implementation."
blocked_by: []

constraints:
  - "Follow the approved visual reference."
  - "Do not add unapproved third-party dependencies."

permissions:
  - "Read approved project assets."
  - "Modify files in the assigned project workspace."
  - "Run local validation commands."

restricted_actions:
  - "Do not deploy or publish."
  - "Do not delete unrelated project files."
  - "Do not disclose credentials or private data."

external_side_effects: false
sensitivity: "internal"

review_requirements:
  - "QA review of responsive behavior"
  - "Accessibility review"
  - "Coordinator integration review"

review_status: "pending"
review_feedback: []
quality_gate: "QA and coordinator acceptance"

status: "PENDING"
progress_summary: "Task created; awaiting dependency completion."
started_at: null
last_status_update_at: null
attempt_count: 0
failure_reason: null
recovery_plan: null

completion_evidence: []
artifact_references: []
verification_results: []
limitations: []
handoff_summary: null
next_actions: []
completed_at: null

change_log:
  - revision: 1
    changed_by: "studio-coordinator"
    change: "Initial task contract created."
    timestamp: null

approval_records: []
```

## 6. Task Creation Rules

Before assigning a task, the Studio Coordinator must:

1. Give the task a unique ID and associate it with a project.
2. Define one accountable owner.
3. State the intended outcome and scope.
4. Provide the required inputs and relevant decisions.
5. Name each expected deliverable.
6. Define acceptance criteria and the verification method where practical.
7. Record dependencies and constraints.
8. Set permissions and restricted actions.
9. Specify required reviewers and approval gates.
10. Choose an initial status consistent with readiness.

If any missing information prevents safe or meaningful execution, the Coordinator must clarify it or record a bounded assumption and seek approval when required. The Coordinator must not silently invent client requirements, deadlines, budgets, assets, approvals, or technical facts.

## 7. Assignment and Readiness Rules

A task may transition to `READY` only when:

- The objective is sufficiently clear.
- The owner is assigned.
- Required inputs are available or their absence is explicitly accepted.
- Mandatory dependencies are satisfied.
- Acceptance criteria are defined.
- Constraints and permissions are clear enough for safe execution.

The owner must acknowledge the assignment according to the capabilities of the runtime. If acknowledgement is not supported, the runtime must record assignment as an event or status transition.

## 8. Status Update Contract

When reporting progress, the owner should provide:

- Current status.
- Work completed since the last update.
- Work remaining.
- Evidence or artifact references, where available.
- Blockers and their impact.
- Decisions or approvals needed.
- Expected next step.

Updates must distinguish facts from estimates. If an estimate cannot be supported, state the uncertainty instead of presenting a false precision.

A task in `BLOCKED` must identify the blocker, its impact, the party or decision needed to unblock it, and any safe work that can continue.

## 9. Completion and Acceptance Rules

An owner may submit a task for review only when the deliverables are available and the acceptance criteria have been checked to the extent possible.

A task may enter `DONE` only when:

1. All required deliverables exist and are referenced accurately.
2. Acceptance criteria are met, or approved exceptions are explicitly recorded.
3. Required verification has been performed and results are recorded.
4. Required review is approved.
5. Required approvals are recorded.
6. Known limitations are disclosed.
7. Handoff information is sufficient for the next stage.

If any criterion remains unmet, the task must remain in `REVIEW`, return to `IN_PROGRESS`, or become `BLOCKED`, as appropriate. A task must not be marked `DONE` merely because work has stopped.

## 10. Review Decision Contract

Reviewers must return one of these decisions:

- `ACCEPT` — Acceptance criteria are met.
- `REQUEST_REVISION` — Specific changes are required before acceptance.
- `MARK_BLOCKED` — Review cannot proceed because an input, decision, or dependency is missing.
- `ESCALATE` — The issue exceeds the reviewer's authority or risk tolerance.
- `REJECT` — The deliverable is unsuitable for acceptance under the current scope.

A review decision should include the reviewer, timestamp where supported, findings, affected acceptance criteria, and the required next action.

The effect of each decision on task status is defined in `core/task-state-model.md`.

## 11. Scope and Contract Changes

The following changes are material and must be reviewed by the Studio Coordinator:

- Objective or deliverable changes.
- Changes to acceptance criteria after work has started.
- Owner reassignment.
- Dependency changes that affect other tasks.
- Permission or restricted-action changes.
- Changes that introduce new costs, external effects, security exposure, or delivery commitments.
- Changes that invalidate previous approvals.

The coordinator must assess downstream impact and update dependent tasks when necessary. An owner may propose a change but may not unilaterally expand the approved scope.

## 12. Failure and Recovery

When a task fails, the owner must report:

- What failed.
- What was attempted.
- The observed error or evidence.
- Whether any artifacts or external systems were changed.
- Whether retrying is safe.
- The recommended recovery path.

Retries must be bounded and must not repeat irreversible or costly actions without authorization. Reassignment must preserve relevant findings and artifacts to avoid repeating failed approaches.

## 13. Concurrency and Artifact Safety

Tasks may run in parallel only when their dependencies and artifact ownership allow it.

Before parallel execution, the Coordinator must identify:

- Shared files or resources.
- Potentially conflicting edits.
- Ordering requirements.
- Integration owner.
- Merge or reconciliation procedure.

Two tasks must not independently overwrite the same canonical artifact unless a deliberate conflict-resolution mechanism exists. Where the runtime cannot enforce file locks or version control, the Coordinator should sequence conflicting work instead.

## 14. Security and External Actions

Task contracts do not override platform policies, runtime permissions, or higher-priority instructions.

Owners must:

- Use only authorized inputs and tools.
- Keep credentials and secrets out of task summaries and deliverables unless a secure, explicitly authorized mechanism requires them.
- Avoid sending confidential material to unapproved destinations.
- Request approval before performing restricted or consequential actions.
- Report suspected data exposure or security issues promptly.

A task's `permissions` field records the task-level authorization; it does not grant capabilities the runtime or platform has not granted.

## 15. Runtime Adapter Requirements

A conforming runtime adapter should support the following operations, directly or through an equivalent mechanism:

- `create_task(contract)`
- `read_task(task_id)`
- `update_task(task_id, changes)`
- `assign_task(task_id, owner)`
- `transition_task(task_id, new_status)`
- `record_evidence(task_id, evidence)`
- `submit_for_review(task_id)`
- `record_review_decision(task_id, decision)`
- `record_approval(task_id, approval)`
- `list_ready_tasks(project_id)`
- `list_blocked_tasks(project_id)`
- `list_tasks_by_status(project_id, status)`
- `list_tasks_by_owner(project_id, owner)`
- `cancel_task(task_id, authorization)`
- `reopen_task(task_id, reason)` (DONE to REVIEW only)

`update_task` must not change material fields (Section 11) except through the change process; adapters should reject or flag such updates.

These operation names are logical interface names, not mandatory executable function names. An adapter must enforce valid state transitions, required fields, permissions, and auditability to the extent supported by its runtime.

If the runtime lacks a capability, the adapter must disclose the limitation and provide a safe manual equivalent. It must not claim that delegation, persistence, parallel execution, file access, or approvals occurred unless the runtime confirms them.

## 16. Minimum Conformance Checklist

A runtime or workflow conforms to this specification when it can demonstrate that:

- [ ] Every task has a stable ID and project ID.
- [ ] Every active task has exactly one accountable owner.
- [ ] Objectives, deliverables, and acceptance criteria are explicit.
- [ ] Dependencies and blockers are represented.
- [ ] Permissions and restricted actions are recorded.
- [ ] Status transitions follow the canonical lifecycle.
- [ ] Completion requires evidence and required review.
- [ ] Material changes are traceable.
- [ ] Failures and limitations are reported honestly.
- [ ] Artifact references are real and verifiable.
- [ ] External or destructive actions require the appropriate authorization.
- [ ] Runtime limitations are disclosed rather than simulated as successful execution.

## 17. Relationship to Other Studio Documents

This specification is intended to work alongside:

- `core/studio-coordinator.md` — Defines the Coordinator's responsibilities and decision authority.
- `core/orchestration-rules.md` — Defines project/task lifecycle, routing, dependencies, and quality gates.
- `core/delegation-protocol.md` — Defines assignment, execution reporting, review, and handoff procedures.
- `core/task-state-model.md` — Defines task statuses, valid transitions, and review-decision mapping.
- `adapters/` — Maps this logical contract to a specific runtime's tools, storage, and agent capabilities.

This specification ranks as a studio-wide core document (precedence level 4). If conflicts appear, follow the precedence rules in `core/orchestration-rules.md`. Update related documents together when a schema change alters lifecycle, delegation, or approval behavior.

## 18. Versioning Policy

Use semantic versioning for this specification:

- **Major:** Breaking changes to required fields, field meaning, lifecycle semantics, or conformance requirements.
- **Minor:** Backward-compatible additions, optional fields, or expanded guidance.
- **Patch:** Clarifications, examples, and corrections that do not change the contract's meaning.

Any version change that affects active projects should include a migration note and an explicit compatibility decision.

## 19. Definition of Done for This Specification

This specification is ready to be considered stable when:

- The Coordinator, orchestration, and delegation documents use compatible terminology.
- The required fields and lifecycle rules are unambiguous.
- A sample task can be created, assigned, blocked, reviewed, revised, and completed using the contract.
- The intended runtime adapter can map the logical contract to its available capabilities.
- Acceptance and authorization requirements have been validated against representative workflows.
