---
id: backend-development
name: Backend Development
version: 1.0.0
type: skill
runtime: portable
status: draft
---

# Skill: Backend Development

## 1. Purpose
Implement APIs, data, and business logic with explicit security and error handling.

## 2. Used By
- `roles/software-engineer.md`

## 3. Required Inputs
- API or data requirements
- Auth requirements
- Environment and stack decisions

## 4. Procedure
1. Define data model and API contract.
2. Implement endpoints and validation.
3. Handle errors and edge cases explicitly.
4. Apply authentication and authorization as specified.
5. Keep secrets out of code and logs.
6. Run available tests; document limitations and setup.

## 5. Outputs
- Source code
- API documentation
- Test results
- Security notes

## 6. Quality Criteria
- Inputs are validated.
- Secrets are not committed.
- Errors return defined responses.
- Destructive migrations require approval.

## 7. Boundaries
- Do not fabricate evidence, sources, results, or file paths.
- Stay within the task's scope and permissions.
- Report limitations and unverified items.
