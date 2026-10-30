---
id: task-state-model
name: Task State Model
version: 1.0.0
type: core-specification
runtime: portable
status: draft
---

# Task State Model

## 1. Purpose

Single source of truth for task statuses, their definitions, valid
transitions, and the effect of review decisions. Other documents refer to
this one and must not redefine these items.

## 2. Statuses

| Status | Definition |
|---|---|
| PENDING | Identified or created, but not yet eligible to start. |
| READY | Readiness conditions are met and execution may begin. |
| IN_PROGRESS | The owner is actively working on the task. |
| BLOCKED | Progress is prevented by an unresolved dependency, decision, permission, resource, or capability. |
| REVIEW | Required output is available and awaits evaluation. |
| DONE | Acceptance criteria are met (or approved exceptions recorded), required verification is recorded, required review is approved, required approvals are recorded, and limitations are disclosed. |
| FAILED | The current attempt failed, or its result was rejected. A recovery decision is required. Not terminal. |
| CANCELLED | Explicitly cancelled by an authorized decision-maker. Terminal. |

DONE does not mean the parent project is approved or complete.

## 3. Valid Transitions

Any transition not listed is invalid.

| From | To | Condition |
|---|---|---|
| PENDING | READY | Readiness rules in the task contract specification are met. |
| PENDING | BLOCKED | A prerequisite cannot be satisfied without a decision or resource. |
| PENDING | CANCELLED | Authorized cancellation. |
| READY | IN_PROGRESS | Owner acknowledged; required capabilities are available. |
| READY | BLOCKED | Owner finds a missing input, permission, or ambiguity (record as readiness defect). |
| READY | PENDING | A mandatory dependency regressed after READY. |
| READY | CANCELLED | Authorized cancellation. |
| IN_PROGRESS | REVIEW | Required output is available for evaluation. |
| IN_PROGRESS | BLOCKED | Blocker encountered. |
| IN_PROGRESS | FAILED | Attempt failed. |
| IN_PROGRESS | CANCELLED | Authorized cancellation. |
| BLOCKED | IN_PROGRESS | Blocker resolved and work had already started (`started_at` set). |
| BLOCKED | READY | Blocker resolved and work had not started. |
| BLOCKED | PENDING | Blocker requires replanning or a new dependency. |
| BLOCKED | FAILED | Blocker cannot be resolved and the approach must change. |
| BLOCKED | CANCELLED | Authorized cancellation. |
| REVIEW | DONE | Decision ACCEPT and all conditions of DONE are satisfied. |
| REVIEW | IN_PROGRESS | Decision REQUEST_REVISION. |
| REVIEW | BLOCKED | Decision MARK_BLOCKED, or ESCALATE while awaiting a decision. |
| REVIEW | FAILED | Decision REJECT. |
| REVIEW | CANCELLED | Authorized cancellation. |
| DONE | REVIEW | An upstream change or discovered defect invalidates acceptance. Record in `change_log`. |
| FAILED | READY | Coordinator authorizes a new attempt with a corrected contract or approach. |
| FAILED | PENDING | New prerequisites are needed before another attempt. |
| FAILED | CANCELLED | Recovery is not authorized or attempts are exhausted. |

DONE never moves directly to IN_PROGRESS; reopening always goes through REVIEW.
CANCELLED has no outgoing transitions. To resume cancelled work, create a new task.

If `attempt_count` has reached the project's retry limit, FAILED -> READY
requires Coordinator escalation to the Founder.

## 4. Review Decision Mapping

Review decisions are not statuses.

| Decision | Resulting status | Notes |
|---|---|---|
| ACCEPT | DONE | Only if every condition of DONE is satisfied; otherwise the task stays in REVIEW and the missing condition is recorded. |
| REQUEST_REVISION | IN_PROGRESS | Findings must name the unmet criteria. |
| MARK_BLOCKED | BLOCKED | Record the missing input or capability. |
| ESCALATE | BLOCKED | While awaiting the decision; resume per the decision. |
| REJECT | FAILED | The Coordinator then chooses a new attempt or cancellation. |

## 5. Review Status vs Task Status

`review_status` (not_required, pending, in_review, changes_requested,
approved, rejected) tracks the review process. `status` tracks the task.
Required alignments:

- `status: REVIEW` requires `review_status` of `pending`, `in_review`, or `not_required`.
- `status: DONE` requires `review_status` of `approved` or `not_required`.
- `review_status: rejected` implies `status: FAILED`.
- `review_status: changes_requested` implies `status: IN_PROGRESS`.

## 6. Project-Level Notes

Project statuses are defined in `core/orchestration-rules.md`. Several names
overlap with task statuses (READY, IN_PROGRESS, REVIEW, BLOCKED). Always
state the scope when reporting ("project READY", "task READY").
