---
id: orchestration-rules
name: Orchestration Rules
version: 1.1.0
type: core-protocol
runtime: portable
status: draft
---

# Orchestration Rules

## 1. Purpose

Define the runtime-neutral rules for project lifecycle management,
task routing, dependency handling, quality gates, and escalation.

These rules govern how the Studio Coordinator orchestrates work across
roles and skills.

## 2. Authority and Precedence

Apply the following precedence model:

1. System and platform safety requirements
2. Runtime permissions and security boundaries
3. Explicit Founder decisions and project approvals
4. Studio-wide operating rules and core specifications
   (everything under `core/`, including the task contract and task state model)
5. Approved project configuration and scope
6. Workflow-specific rules
7. Role instructions
8. Skill-specific procedures
9. Individual task instructions

This ordering must not be used to override a higher-priority safety
requirement or an explicit restriction.

If two instructions at the same authority level conflict, identify the
conflict and seek clarification when it materially affects execution.

A task-specific instruction cannot silently authorize a scope change
or override a mandatory approval requirement.

If an explicit Founder decision (level 3) differs from the recorded project
configuration (level 5), the decision governs. The Coordinator must update
the project configuration and affected tasks before issuing the next
dependent task.

## 3. Canonical Project Lifecycle

Use the following lifecycle states:

- INTAKE
- CLARIFICATION_REQUIRED
- PLANNING
- READY
- IN_PROGRESS
- REVIEW
- BLOCKED
- DELIVERY_READY
- COMPLETED
- CANCELLED

### State Definitions

INTAKE:
The request has been received but not fully assessed.

CLARIFICATION_REQUIRED:
A material ambiguity prevents safe or reliable planning.

PLANNING:
Scope, tasks, dependencies, roles, and acceptance criteria are being defined.

READY:
The plan is sufficiently complete and authorized for execution.

IN_PROGRESS:
One or more tasks are being executed.

REVIEW:
Required outputs are undergoing validation or specialist review.

BLOCKED:
Progress cannot continue because a dependency, decision, permission,
resource, or capability is unavailable.

DELIVERY_READY:
Required deliverables have passed applicable checks and are awaiting
any remaining mandatory approval or handoff.

COMPLETED:
The required completion conditions have been met and required approvals
are recorded. COMPLETED is terminal.

CANCELLED:
The project has been explicitly cancelled by an authorized party.
CANCELLED is terminal.

Note: READY, IN_PROGRESS, REVIEW, and BLOCKED also exist as task statuses
with different meanings. In logs and reports, always state the scope
("project READY" or "task READY").

## 4. State Transition Rules

INTAKE -> CLARIFICATION_REQUIRED:
A material ambiguity must be resolved before reliable planning.

INTAKE -> PLANNING:
The available brief is sufficient to construct a plan.

CLARIFICATION_REQUIRED -> PLANNING:
Required clarification has been obtained or an authorized assumption
has been documented.

PLANNING -> READY:
Tasks, ownership, dependencies, acceptance criteria, and required approvals
have been established.

READY -> IN_PROGRESS:
Execution is authorized and the required capabilities are available.

IN_PROGRESS -> REVIEW:
A deliverable or milestone is ready for evaluation.

IN_PROGRESS -> BLOCKED:
A material dependency, permission, decision, or execution capability
prevents further progress.

REVIEW -> IN_PROGRESS:
A deliverable requires correction or additional work.

REVIEW -> DELIVERY_READY:
Applicable quality gates have passed and outstanding exceptions are
documented and handled according to project policy.

BLOCKED -> IN_PROGRESS:
The blocker is resolved and affected tasks are ready to resume.

BLOCKED -> PLANNING:
The blocker requires material replanning.

DELIVERY_READY -> COMPLETED:
All mandatory approval, handoff, and completion requirements are satisfied.

DELIVERY_READY -> IN_PROGRESS:
The Founder or an authorized approver rejects the delivery or requests
changes. The Coordinator records the findings, creates revision tasks,
and reassesses gates that the changes invalidate.

DELIVERY_READY -> BLOCKED:
A required approval or handoff cannot be obtained.

Any active state -> CANCELLED:
An authorized cancellation is recorded and remaining work is handled safely.
Active states are all states except COMPLETED and CANCELLED.

Do not skip mandatory approval or quality gates.

### 4.1 Changes After Completion

COMPLETED projects are not reopened. Post-completion changes start a new
project or change-request project that references the completed project
and its artifacts. Material defects found after completion are reported
to the Founder, who decides whether to open a new project.

## 5. Task Status Model

The canonical task statuses, their definitions, valid transitions, and the
mapping from review decisions to statuses are defined in
`core/task-state-model.md`. That document is the single source of truth.
Do not redefine task statuses elsewhere.

Task statuses:

- PENDING
- READY
- IN_PROGRESS
- BLOCKED
- REVIEW
- DONE
- FAILED
- CANCELLED

A task may enter REVIEW only when its required output is available for
evaluation.

DONE does not automatically mean the entire project is approved.

## 6. Task Routing

For each task, the Coordinator must:

1. Identify the intended outcome.
2. Identify required expertise.
3. Identify the relevant role.
4. Select the minimum required skills.
5. Confirm that the required tools and permissions are available.
6. Assign one accountable owner.
7. Provide the task contract and relevant context.
8. Define acceptance criteria and dependencies.
9. Specify review requirements.
10. Record the assignment and status.

Route tasks according to capability, not merely by matching keywords.

If no suitable role exists, the Coordinator may perform the task within
its competence or request the creation of a new specialist capability.

Do not claim that an unconfigured role has been activated.

## 7. Dependency Management

Represent material task dependencies explicitly.

A dependent task must not begin until its required inputs are available
and approved to the extent required by the workflow.

Independent tasks may be executed in parallel when:

- The runtime supports parallel execution.
- Inputs are sufficiently stable.
- Tasks do not conflict over shared artifacts.
- Each task has a clear owner.
- The expected benefit exceeds coordination overhead.

Otherwise, execute tasks sequentially.

If an upstream deliverable changes, identify and reassess all affected
downstream tasks. A DONE task whose inputs changed returns to REVIEW
through the transitions in `core/task-state-model.md`.

## 8. Context and Artifact Management

Pass only relevant project context to each task.

Prefer explicit artifact references over copying entire conversation
histories.

For every material artifact, track where available:

- Artifact identifier or path
- Artifact type
- Responsible owner
- Version or revision
- Source task
- Current status
- Relevant acceptance criteria
- Review status

Do not assume artifact storage, version control, or persistence exists
unless provided by the runtime.

Never fabricate a file path or claim that an artifact was saved when
the operation has not succeeded.

## 9. Quality Gates

Use applicable gates based on the deliverable type.

GATE A — Requirements:
The work maps to the approved brief and acceptance criteria.

GATE B — Specialist Review:
Required specialist review has been completed.

GATE C — Verification:
Applicable tests, inspections, or evidence checks have been performed.

GATE D — Integration:
Outputs are compatible with dependent deliverables and project constraints.

GATE E — Approval:
Required Founder or stakeholder approval has been recorded.

GATE F — Delivery:
Required files, documentation, known limitations, and handoff information
are accounted for.

Not every task requires every gate. The workflow must define which gates
are mandatory.

Default when a workflow does not define mandatory gates: Gate A and
Gate F are mandatory for every deliverable, and Gate E is mandatory
whenever a task or project has `approval_required: true`. Gate C is
mandatory for any deliverable whose behavior can be tested with
available tools.

A failed mandatory gate prevents progression unless an authorized exception
is explicitly recorded.

## 10. Rework and Iteration

When a deliverable fails review:

1. Record the specific failed criteria.
2. Identify the responsible task owner.
3. Provide actionable findings and relevant evidence.
4. Create a revision task or return the task to execution.
5. Reassess affected dependencies.
6. Repeat the applicable checks after correction.

Do not request vague revisions such as "make it better" when concrete
findings can be supplied.

If repeated attempts fail, reassess the underlying approach and escalate
when necessary.

## 11. Blockers and Escalation

A task must be marked BLOCKED when it cannot proceed reliably within
its current scope, permissions, dependencies, or capabilities.

Record:

- Blocker description
- Affected tasks
- Impact
- Information or permission required
- Available alternatives
- Recommended next action
- Whether Founder approval is required

Escalate material scope conflicts, security concerns, financial commitments,
irreversible operations, and external publication requirements.

Do not bypass a blocker by inventing missing inputs or assuming permission.

## 12. Scope Change Management

A proposed material change must include:

- Requested change
- Reason
- Impact on deliverables
- Impact on dependencies
- Impact on cost or effort, if known
- Impact on timing, if known
- Required approvals

The Coordinator may propose a revised plan but must not treat a material
scope change as approved until the authorized decision is recorded.

After approval, update the project configuration and affected tasks.

## 13. Completion Rules

A project can become COMPLETED only when:

- Required deliverables are accounted for.
- Mandatory quality gates have passed.
- Required approvals are recorded.
- Known limitations are disclosed.
- Outstanding issues are handled according to project policy.
- The final report is accurate.

A blocked, failed, or unverified deliverable must not be silently treated
as complete.

## 14. Runtime Independence

These rules define logical behavior, not a specific execution mechanism.

A runtime adapter is responsible for translating:

- Role selection into agent or capability invocation
- Task contracts into executable assignments
- Status changes into available persistence mechanisms
- Artifact references into supported storage operations
- Tool requirements into supported tool calls
- Approval gates into available authorization mechanisms

If a runtime lacks a required capability, preserve the logical rule and
report the limitation. Do not simulate successful execution without evidence.

## 15. Definition of Done

Orchestration is functioning correctly when:

- Every active task has an owner and status.
- Dependencies are explicit.
- Routing decisions have a clear rationale.
- Mandatory gates cannot be silently skipped.
- Blockers and scope changes are traceable.
- Project state reflects actual progress.
- Final completion is supported by evidence.
