You are implementing the next stage of the FastFood System project. Follow these steps in order.

## Arguments
$ARGUMENTS — optional. Can be `--stage N` to target a specific stage, or `--step N` to target a sub-step within a stage. If omitted, determine the next unimplemented stage automatically.

## Step 1: Determine what to implement

Run these in parallel:
- Read `documentation/Execution-order.md` — the canonical list of all stages and their steps
- Run `git log --oneline` to see which stages have already been committed
- Read `documentation/BRD.md` for business requirements and constraints
- Read `documentation/developer_guide.md` for architecture rules you must follow

Cross-reference the git log commit messages against the stages in Execution-order.md to find the next unimplemented stage (or use the `--stage`/`--step` argument if provided).

## Step 2: Read the architecture before touching code

Before writing any code, read:
- `CLAUDE.md` — key conventions (monetary values as centavos int64, UUID v7 for orders, pgx positional params, etc.)
- `cmd/api/main.go` — the wiring file; any new service/repo/handler must be registered here
- The existing implementation of the most recently completed stage for pattern reference

## Step 3: Plan in plan mode

Enter plan mode and produce a plan that covers:
- Which files to create and which to modify
- The exact interface signatures for any new repository and service interfaces
- SQL migration needed (if any) and the migration name
- Route registrations to add in `cmd/api/main.go`
- How the new code fits the Handler → Service (interface) → Repository (interface) → pgxpool.Pool chain

Get the plan approved before writing any code.

## Step 4: Implement

Follow the layered architecture strictly:
- `internal/models/` — domain structs and DTOs (`*_dto.go` alongside the domain file)
- `internal/repository/` — raw SQL via pgx, define the interface here
- `internal/service/` — business logic, depend only on repository interfaces
- `internal/handlers/` — Chi HTTP handlers, depend only on service interfaces
- `cmd/api/main.go` — wire everything together

## Step 5: Verify

After implementation, run in sequence:
```bash
make fmt
make vet
make lint
make test
make build
```

Fix any failures before declaring done. Report which stage/step was implemented and what was created.
