---
id: responsive-audit
name: Responsive Audit
version: 1.0.0
type: skill
runtime: portable
status: draft
---

# Skill: Responsive Audit

## 1. Purpose
Verify layout and behavior across defined viewport sizes.

## 2. Used By
- `roles/qa-engineer.md`

## 3. Required Inputs
- Deliverable location
- Breakpoint list
- Browser targets

## 4. Procedure
1. Confirm the deliverable can be opened and tested.
2. Check each viewport for overflow, clipping, and readability.
3. Check touch targets, navigation, and media scaling.
4. Check keyboard and focus behavior.
5. Record evidence (notes or screenshots where available).
6. List defects by severity and untested areas.

## 5. Outputs
- Audit results per viewport
- Defect list
- Readiness statement

## 6. Quality Criteria
- Each tested viewport has a recorded result.
- Untested viewports are stated.
- Findings are reproducible.

## 7. Boundaries
- Do not fabricate evidence, sources, results, or file paths.
- Stay within the task's scope and permissions.
- Report limitations and unverified items.
