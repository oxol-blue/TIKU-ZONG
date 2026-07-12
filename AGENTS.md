# Repository Guidelines

## Project Structure

This repository is currently a blank scaffold. The existing project plan is stored in `doc/`. As implementation begins, use this layout:

- `frontend/`: Vue 3, TypeScript, and `koi-ui` application code.
- `backend/`: Go/Gin API, services, database access, and migrations.
- `frontend/tests/` and `backend/**/_test.go`: automated tests.
- `assets/`: shared images, fixtures, and non-secret static resources.
- `doc/`: Markdown requirements, architecture, API, deployment, and development records.

Keep frontend and backend changes isolated. Put reusable business logic in backend services and API calls in dedicated frontend modules.

## Build, Test, and Development Commands

These commands become available after the corresponding applications are initialized:

- `cd frontend; pnpm install`: install frontend dependencies.
- `cd frontend; pnpm dev`: start the Vite development server.
- `cd frontend; pnpm build`: type-check and build the frontend.
- `cd backend; go run ./cmd/server`: run the Gin API locally.
- `cd backend; go build ./...`: compile all backend packages.
- `cd backend; go test ./...`: run backend unit and integration tests.

Run `gofmt -w` on changed Go files and the repository frontend formatter/linter before committing.

## Coding Style and Naming

Use tabs and idiomatic Go naming; document exported identifiers where useful. Use two-space indentation in TypeScript, Vue, JSON, and YAML. Use PascalCase for Vue components and Go types, camelCase for TypeScript variables and functions, and kebab-case for routes and filenames. Use descriptive Markdown headings and fenced command examples.

## Testing Guidelines

Name Go tests `Test<Behavior>` and keep them beside the package under test. Frontend tests should describe user-visible behavior and live under `frontend/tests/` or beside the component. Cover authentication, API-key authorization, package deduction, payment callbacks, question matching, and OCS/AI fallback. Run relevant tests before opening a pull request.

## Commits and Pull Requests

Use Conventional Commits, for example `feat(auth): add email login` or `fix(payment): make callback idempotent`. Pull requests should explain the change, link an issue, list verification commands, and include UI screenshots when applicable. Keep them focused and call out migrations or configuration changes.

## Security and Configuration

Never commit API keys, payment or AI credentials, passwords, tokens, or production connection strings. Use Git-ignored `.env` files and sanitized examples. Hash user API keys, encrypt provider secrets, validate payment callbacks server-side, and redact secrets from logs.
