---
id: website-production
name: Website Production Workflow
version: 1.0.0
type: workflow
runtime: portable
status: draft
---

# Workflow: Website Production

## 1. Purpose
Take a website request from intake to delivery-ready, with explicit tasks, dependencies, gates, and approval points. Publication is never part of this workflow's autonomous scope.

## 2. Applicability
- Use for new websites or major site sections (portfolio, business site, landing page, interactive site).
- Not for web applications with accounts, databases, or business logic; use a web app workflow (not yet defined).
- Use the Lightweight Variant (Section 9) when the request qualifies for lightweight intake.

## 3. Precedence and Gates
This workflow ranks at precedence level 6 (`core/orchestration-rules.md`). It cannot override core rules or approvals.

Mandatory gates for this workflow:

| Gate | Applies to |
|---|---|
| A: Requirements | Every deliverable |
| B: Specialist Review | Creative direction, UI/UX spec, implementation |
| C: Verification | Implementation (responsive audit, accessibility check, console check) |
| D: Integration | Final assembly of design, content, and code |
| E: Approval | Brief, creative direction, final delivery, and any publication |
| F: Delivery | Final handoff |

## 4. Stages and Tasks

Task IDs are templates; the Coordinator assigns real IDs per project. All tasks use `core/task-contract-specification.md`.

| # | Task | Owner | Skills | Depends on | Output |
|---|---|---|---|---|---|
| T1 | Project brief and scope | product-strategist | prd-generation | Founder request | Brief, scope, acceptance criteria |
| T2 | Research and references | research-analyst | competitor-analysis | T1 | Research brief, source log |
| T3 | Creative direction | creative-director | visual-identity | T1 (T2 if run) | Creative brief, art direction |
| T4 | Visual identity (if needed) | brand-designer | visual-identity | T3 | Brand specification, assets |
| T5 | UI/UX specification | ui-ux-designer | ui-ux-specification | T1, T3 (T4 if run) | Sitemap, flows, screen specs |
| T6 | Content and copy | content-strategist | content-planning | T1, T3 | Copy deck |
| T7 | Implementation | software-engineer | frontend-development, interactive-website | T5, T6 (T4 if run) | Site source, setup notes |
| T8 | QA verification | qa-engineer | responsive-audit | T7 | Test results, defects, readiness |
| T9 | SEO basics (conditional) | marketing-growth | seo-audit | T6, T7 | SEO findings |
| T10 | Integration and delivery report | studio-coordinator | none | T7, T8 (T9 if run) | Delivery report |

Conditional tasks: T2 only if the brief needs evidence beyond what the Founder supplied; T4 only if brand assets are missing; T9 only if search visibility is a stated goal. Do not activate a role merely because it exists.

## 5. Sequencing and Parallelism
- T2 may run in parallel with early T3 only if T3 does not depend on research conclusions.
- T5 and T6 may run in parallel once T3 is accepted, if they do not edit the same artifact.
- T7 does not begin until T5 and T6 are DONE (not merely submitted).
- T8 is performed by a role other than the T7 owner.

## 6. Approval Points (Founder)
1. After T1: approve brief, scope, and acceptance criteria.
2. After T3 (and T4 if run): approve creative direction before UI specification and build.
3. After T8 and T10 preparation: approve delivery.
4. Separate approval before any publication, domain change, paid service, or external contact.

Silence is not approval.

## 7. Revision Loop
1. QA or reviewer returns a decision per `core/task-state-model.md`.
2. REQUEST_REVISION findings name the unmet criteria and evidence.
3. Fix, then recheck the affected criteria and relevant regressions.
4. After two failed revision cycles on the same criterion, the Coordinator reassesses the approach and escalates when needed.

## 8. Delivery Requirements
Delivery report includes: deliverables with real references, per-criterion results, verification actually performed and not performed, known limitations, open issues, and approvals recorded. Never describe unverified work as verified.

## 9. Lightweight Variant
Use when the site is a single page or a small, simple set of pages with supplied content and no new costs.
- Tasks: T1 (reduced), T5 + T6 (merged into one owner-approved page spec if the Coordinator judges it safe), T7, T8, T10.
- Skip T2, T4, T9 unless explicitly needed.
- Gates A, C, E, F remain mandatory. Approval before publication remains mandatory.
- Use the lightweight task profile and lightweight completion report.

## 10. Exit Conditions
The workflow ends when the project reaches DELIVERY_READY and then COMPLETED per `core/orchestration-rules.md`, or is CANCELLED by an authorized party.
