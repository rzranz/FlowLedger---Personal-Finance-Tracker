# Changelog

## 1.1.0 (core patch)

Resolves cross-document conflicts found in review.

- NEW `core/task-state-model.md` v1.0.0: single source for task statuses, valid transitions, review-decision mapping, and review_status alignment.
- DONE: now one definition everywhere (acceptance met + verification recorded + review approved + approvals recorded + limitations disclosed).
- FAILED: now one definition; non-terminal; recovery decision required. Terminal exhaustion is expressed by FAILED -> CANCELLED.
- `core/orchestration-rules.md`: added DELIVERY_READY -> IN_PROGRESS and DELIVERY_READY -> BLOCKED; COMPLETED is terminal, post-completion changes start a new project; precedence level 4 includes core specifications; Founder decision vs project configuration sync rule; default mandatory gates; scope-prefix note for overlapping status names.
- `core/delegation-protocol.md`: schema now defers to the task contract spec with a field-name mapping; reviewer vs Coordinator roles; independent review fallback; REJECT/REQUEST_REVISION mapping by reference; untrusted-content notification format; lightweight completion report; readiness-defect rule.
- `core/studio-coordinator.md`: lightweight intake; intake/planning boundary; explicit references to task contract and state model.
- `core/task-contract-specification.md` v1.1.0: lightweight profile (4.13); example flattened and `review_status` corrected; per-criterion `verification_method`; default semantics for `output_format`/`destination`; `source_of_truth` as list; new adapter operations (`cancel_task`, `reopen_task`, `list_tasks_by_owner`); `update_task` limit on material fields; precedence level stated.
- NEW `workflows/website-production.md` v1.0.0.

Migration: tasks created under 1.0.0 that use `FAILED` for terminal exhaustion should be moved to CANCELLED with `failure_reason` kept. Delegation-protocol field names map per its Section 3 table.

## 1.0.0
- Initial core drafts, 9 role files, 10 skill files.

## Still open
- Project record schema (scope changes, approvals, decision log).
- `communication-protocol`, `approval-policy`, `failure-handling`.
- Other workflows, templates, standards, `adapters/` (runtime adapter interface).
- Stack, breakpoint, and coding conventions for frontend skills (needs Founder decision).
