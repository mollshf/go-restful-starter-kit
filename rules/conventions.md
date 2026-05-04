# Conventions

Hard rules for this project. Rationale and tradeoffs live in [architecture.md](architecture.md).

## Commands

```bash
make run              # Run the application (uses PORT and DATABASE_URL from .env)
make migrate-up       # Apply migrations via goose
make migrate-down     # Rollback last migration
make migrate-status   # Show migration state
make migrate-create name=<migration_name>  # Create new migration file
make sqlgen           # Regenerate sqlc Go code from query/*.sql
```

No dedicated test or lint commands. Use `go test ./...` and `go vet ./...` directly.

Copy `.env.example` to `.env` and set `DATABASE_URL` (PostgreSQL connection string) before running.

## Project Structure

```
cmd/app/
  main.go         →  init DB, global middleware, start server
  config.go       →  load env, app configuration
  routes.go       →  register module routes under /api

internal/
  shared/
    utils/        →  pure utility, NO HTTP/I/O (hash, validator, cookie)
    web/          →  HTTP utility, Gin-specific (Wrap, response helpers)
      middleware/ →  cross-module middleware (session, rate_limit, max_bytes)
    queries/      →  sqlc generated — DO NOT EDIT MANUALLY (regenerate via make sqlgen)
    cache/        →  Redis wrapper
    mailer/       →  email wrapper
    storage/      →  file storage wrapper

  <module>/       →  e.g. auth, akademik, keuangan, aset
    handler.go    →  HTTP layer (use Wrap from shared/web)
    service.go    →  business logic
    repository.go →  infrastructure orchestrator
    dto.go        →  HTTP input/output types
    routes.go     →  DI wiring + route registration
    middleware.go →  module-specific middleware (optional)
```

`internal/<module>/` is for domain modules only. Infrastructure, HTTP utility, sqlc, and pure utils all live under `internal/shared/`.

## Three-Layer Architecture

Each module follows: **handler → service → repository**.

| Layer             | Does                                                                       |
| ----------------- | -------------------------------------------------------------------------- |
| `cmd/app/`        | Setup, register routes, start server                                       |
| `handler.go`      | Receive request, validate input, send response                             |
| `service.go`      | Business logic and decisions                                               |
| `repository.go`   | Compose DB / cache / mailer / storage to fulfill what service asks for     |

Repository is the module's only gateway to infrastructure. A single repository method may compose multiple side effects (DB write + cache invalidation + email) within one operation — service does not orchestrate those calls itself.

## Import Rules

```
module A         →  FORBIDDEN to import module B
module A         →  ALLOWED to import internal/shared/*
shared/utils/    →  FORBIDDEN to import Gin or any HTTP package
shared/web/      →  ALLOWED to import Gin
shared/queries/  →  generated; never edited manually
```

If two or more modules need the same logic, place it in `shared/utils/` (pure) or `shared/web/` (HTTP) — never duplicate.

## Wiring

Each module wires itself in its own `routes.go`: `infrastructure → repository → service → handler`. No global `App` struct, no DI container.

`cmd/app/routes.go` only registers modules. `auth.AuthRoutes` returns `web.AuthMiddlewares` (defined in `internal/shared/web/`); other modules accept that bag without importing `internal/auth/`.

If a module needs more than 3–4 infrastructure handles, group them in a single `shared.Infra` struct rather than growing the parameter list — but only when actually needed.

Constructor errors (`NewService`, `NewRepository`) must propagate to `cmd/app/main.go` and fail loudly at startup. Never discard with `_`.

## Type Placement

| Type                    | Location                                                  |
| ----------------------- | --------------------------------------------------------- |
| HTTP input/output       | `dto.go` in the module                                    |
| Repository parameter    | `repository.go` in the module                             |
| Internal transform type | Same file as the function that uses it, just above it     |
| Cross-module type       | `shared/utils/`                                           |

## Naming

- DTOs: `XxxRequest` / `XxxResponse` (HTTP only — never use `Request` suffix for non-HTTP types).
- Repository params: `XxxInput` or `XxxArgs` (avoid collision with sqlc `XxxParams`).
- If a repository method passes through to a single sqlc query unchanged, reuse the sqlc `XxxParams` directly.

## Context Propagation

Every method from handler down to repository takes `ctx context.Context` as its first parameter. Repository passes `ctx` to all infra calls (DB, cache, mailer). **Non-negotiable.**

## Value vs Pointer Returns

Default to value returns. Use a pointer only when the struct is **larger than ~64 bytes** or when nilability genuinely carries meaning to the caller.

Do **not** signal "not found" with `nil` just to justify a pointer return. Return a module-level sentinel error from repository (e.g. `ErrUserNotFound`) and let service handle it via `errors.Is` — see [Error Handling](#error-handling).

## Error Handling

Translate infra errors to module-level sentinels at the **repository boundary**. Service operates on sentinels only — it must not import `pgx`, `redis`, or other infra packages for error handling.

**Module-level sentinels** live in `errors.go` (or at the top of `repository.go` while still few):

```go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailDuplicate    = errors.New("email already exists")
    ErrUsernameDuplicate = errors.New("username already exists")
)
```

**Repository:** translate infra errors that have **domain meaning** to sentinels. Wrap the rest with context.

```go
// Lookup miss → sentinel
user, err := pg.repo.FindUserByID(ctx, id)
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return queries.User{}, ErrUserNotFound
    }
    return queries.User{}, fmt.Errorf("find user by id: %w", err)
}

// Constraint violation → sentinel (per constraint name)
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    switch pgErr.ConstraintName {
    case "users_email_key":    return ErrEmailDuplicate
    case "users_username_key": return ErrUsernameDuplicate
    }
}
return fmt.Errorf("insert user: %w", err)
```

**Service:** propagate sentinels and wrapped errors upward. Branch on sentinels via `errors.Is` only when service needs to add domain context (e.g. mapping to a different sentinel for a higher-level operation). Never type-switch on infra errors.

**Handler:** map sentinels to HTTP status. Default to 500 for everything else.

```go
switch {
case errors.Is(err, ErrUserNotFound):   return web.NotFound(c, err)
case errors.Is(err, ErrEmailDuplicate): return web.Conflict(c, err)
default:                                return web.InternalError(c, err)
}
```

**Rule of thumb:** create a sentinel only for errors the handler must distinguish. If the result is 500 either way, just wrap with `fmt.Errorf("...: %w", err)` — no sentinel needed.

**Forbidden:** signaling "not found" by returning `(nil, nil)` from repository. Always return a sentinel. See [Value vs Pointer Returns](#value-vs-pointer-returns).

## Transaction Pattern

```go
tx, err := pg.db.Begin(ctx)
if err != nil { return nil, err }
defer tx.Rollback(ctx)

qtx := pg.repo.WithTx(tx)

// Tx step 1: <what & why>
// ...

// Tx step 2: <what & why>
// ...

if err = tx.Commit(ctx); err != nil { return nil, err }
```

Always `defer tx.Rollback()` immediately after `tx.Begin()`. After successful `Commit()`, Rollback is a no-op.

Every `qtx.X(...)` inside a transaction must be preceded by `// Tx step N: <intent>`. Single-query repository methods do not need this.

## Response Envelope

All responses follow:
```json
{ "success": true|false, "data": ..., "error": ..., "meta": { "page": ..., "total": ... } }
```
Use `shared/web` response helpers — never write raw `c.JSON` in handlers.

## Interfaces

Define interfaces only when writing tests that require mocking — typically at the service layer for mocking the repository. Repository tests are integration tests (real DB + real Redis + fake mailer). No speculative interfaces.

## Cross-Module Data Sharing

- All modules import generated types from `internal/shared/queries/` — sqlc is the only legitimate "god package".
- SQL files per module: `query/sqlc_auth.sql`, `query/sqlc_akademik.sql`. Generated code stays in one `queries` package.
- Cross-domain joins belong to one owner module — pick the module whose business domain the result represents.
- Cache keys use a per-module namespace prefix: `auth:session:{user_id}`, `akademik:mahasiswa:{nim}`, `keuangan:tagihan:{id}`.
- If module A needs business logic from module B, define a small interface in module A's repository and have module B implement it.

### Write Discipline

`INSERT` / `UPDATE` / `DELETE` queries hanya boleh dipanggil dari repository modul pemiliknya. `SELECT` bebas lintas modul.

- Pemilik tabel = modul yang business domain-nya bertanggung jawab atas invariant tabel itu (validasi, side effect, audit, cache invalidation).
- Jika modul A perlu memicu write di domain B, A definisikan interface kecil, B implement. Contoh: `akademik` ingin menutup tagihan saat mahasiswa lulus → akademik definisikan `TagihanCloser`, keuangan implement.
- Aturan ini **konvensi, bukan compiler-enforced**: karena `shared/queries/` di-share, secara teknis modul mana pun bisa memanggil query write apa pun. Disiplin dijaga lewat code review. → see [architecture.md#write-discipline-and-the-shared-sqlc-trade-off](architecture.md#write-discipline-and-the-shared-sqlc-trade-off).

## Key Dependencies

| Package                       | Purpose                             |
| ----------------------------- | ----------------------------------- |
| `github.com/gin-gonic/gin`    | HTTP router                         |
| `github.com/jackc/pgx/v5`     | PostgreSQL driver + connection pool |
| `github.com/sqlc-dev/sqlc`    | SQL-to-Go generator (dev tool)      |
| `github.com/pressly/goose/v3` | Database migrations                 |
| `github.com/joho/godotenv`    | `.env` loading                      |
| `golang.org/x/crypto`         | bcrypt password hashing             |
| `github.com/google/uuid`      | UUID generation                     |
