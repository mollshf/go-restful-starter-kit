# Architecture & Design Rationale

This document explains the *why* behind the rules in [conventions.md](conventions.md). Rules belong there; rationale lives here. Avoid duplicating either way — link instead.

## Modulith with one PostgreSQL database

Single database for all modules. Reasons:

- High referential integrity between domains (akademik ↔ keuangan ↔ auth) makes split-DB joins painful.
- One transaction can span domains where business invariants demand it.
- Backup, migrations, and observability are simpler.

The cost: every module shares the same schema namespace. Mitigated by per-module SQL files in `query/<module>.sql`.

## Why `internal/shared/`?

Top-level `internal/` is reserved exclusively for domain modules (`auth`, `akademik`, `keuangan`, `aset`, ...). Everything that is not a domain — infrastructure wrappers, HTTP utilities, pure utils, sqlc-generated queries — lives under `internal/shared/`.

The redundancy of "internal + shared" is intentional: it communicates intent at a glance ("this is shared by all modules under internal").

## Repository as infrastructure orchestrator

Repository is the module's exclusive gateway to infrastructure. Service expresses *what* the business needs; repository decides *how* infrastructure delivers it.

A single repository method may compose multiple side effects in one operation. For example, `RegisterUser` may insert into DB, invalidate a cache key, and send a welcome email — all within one method. Service does not orchestrate those calls; it asks repository to "register the user" and trusts repository to coordinate the infrastructure.

**Heuristic:** if service ever calls `repo.X()` then `repo.Y()` then `repo.Z()` to complete a single business operation, that orchestration belongs *inside* one repository method, not in service.

**Trade-offs:**

- Repository methods grow as features expand. That is expected.
- Service tests are easy (mock one interface — the repository). Repository tests are integration tests against real DB + real Redis + fake mailer.
- If a caching policy varies per use case (`GetUserFresh` vs `GetUserCached`), expose two methods rather than push the decision into service.

## Why translate errors at the repository boundary

Repository is the only layer that knows about `pgx`, `redis`, and other infrastructure. Error translation follows the same boundary: infra errors that have **domain meaning** are converted to module-level sentinels (`ErrUserNotFound`, `ErrEmailDuplicate`, ...) inside repository. Service operates on sentinels exclusively.

**Why not let service translate?** It would force service to import `pgx` and `pgconn` just to pattern-match on `pgx.ErrNoRows` or `PgError.Code == "23505"`. That couples business logic to the database driver — swapping pgx, or wrapping a query in a different infrastructure (cache fallback, read replica), would ripple into service. With repository-side translation, only repository changes when infra changes.

**Why not return `(nil, nil)` for not-found?** It collapses two distinct outcomes (no row, error during lookup) into the same shape, forcing every caller to do `if x == nil` *and* `if err != nil` — the second check becomes easy to forget. A sentinel makes the contract explicit: errors are always errors, values are always values. It also pairs cleanly with value returns (no need to make a return type a pointer just so you can return `nil`).

**Heuristic for what deserves a sentinel:**

- Yes — the handler must distinguish it: `ErrUserNotFound` (404), `ErrEmailDuplicate` (409), `ErrSlotKelasPenuh` (409), `ErrPrasyaratBelumLulus` (400).
- No — it ends up as 500 either way: connection lost, query timeout, malformed SQL, disk full. Wrap with `fmt.Errorf("...: %w", err)` and let it propagate.

The test is whether *the handler needs to branch on it*. If not, a sentinel adds noise without value.

**Service's role.** Service rarely translates errors; it propagates them. The exception is when service composes multiple repository calls and wants to expose a higher-level domain error — e.g. `Login` may translate `ErrAccountNotFound` *and* `ErrPasswordMismatch` into a single `ErrInvalidCredentials` to avoid leaking which factor failed (a security concern, not an infra concern). That kind of translation is business logic and belongs in service.

**Where sentinels live.** In `errors.go` at the module root once you have more than a handful, or at the top of `repository.go` while still few. They are part of the module's public-ish surface — handler imports them to map HTTP status, service imports them to compose higher-level errors. Keep them in one place.

## Cross-module data sharing — why these mechanisms

Modules never import each other. Two mechanisms enable sharing without coupling:

**1. Shared data shape via sqlc.** All modules import generated types from `internal/shared/queries/`. The DB schema is the contract; modules read it through generated code, never through each other. sqlc is the only legitimate "god package" because it contains zero business logic — it only translates SQL to Go.

**2. Per-module repository wraps infrastructure.** Each module's `repository.go` injects only the infrastructure it needs from `internal/shared/`. Different modules inject different subsets:

```go
// internal/auth/repository.go — needs DB, queries, cache, mailer
type Repository struct {
    db      *pgxpool.Pool
    queries *queries.Queries
    cache   *cache.Redis
    mailer  *mailer.SMTP
}

// internal/akademik/repository.go — needs DB, queries, storage (for transcript PDFs)
type Repository struct {
    db      *pgxpool.Pool
    queries *queries.Queries
    storage *storage.S3
}
```

## Write discipline and the shared sqlc trade-off

The shared `internal/shared/queries/` package gives every module access to every generated query — including writes to tables it does not own. The compiler cannot stop `internal/akademik/` from calling `r.queries.InsertTagihan(...)`. This is a conscious trade-off:

- **What we gain by sharing sqlc:** cross-domain reads stay one query (with JOINs), shared types stay consistent, multi-domain transactions can share one `tx` via `WithTx`. These benefits matter precisely *because* this is a modulith on a single Postgres.
- **What we give up:** structural enforcement of "writes belong to the owner". The discipline becomes a convention.

The rule in [conventions.md#write-discipline](conventions.md#write-discipline) — writes only from owner, reads anywhere — exists because writes carry invariants (validation, audit, cache invalidation, downstream notifications) that the owning module is responsible for. If two modules can write to the same table, those invariants get duplicated or, worse, silently skipped. Reads carry no invariants, so sharing them is safe.

**How we keep the convention honest:**

1. **SQL file per module** (`query/sqlc_auth.sql`, `query/sqlc_akademik.sql`) makes ownership obvious at the source.
2. **Code review** — a call to `r.queries.<OtherModule><Verb>(...)` from outside the owner is the smell to catch.
3. **Cross-module write requests use interfaces** — the asking module defines a small interface, the owner implements it. This forces the write to go through the owner's repository, where invariants live.

**Exit ramp.** If the project later splits into separate services, the natural step is to move each module's SQL into its own sqlc package (`internal/<module>/queries/`). At that point the import rule (modules cannot import each other) extends to queries, and the convention becomes compiler-enforced. We have not done this now because it costs cross-domain JOINs and shared types — costs that are worth paying only when the modulith is actually being broken apart.

## Wiring deep dive

Each module wires itself linearly: infrastructure → repository → service → handler. Service has only one dependency: its repository. Handler has only one dependency: its service.

```go
// cmd/app/routes.go — auth is wired first; its Middlewares bag is passed
// to every other module so they can enforce authentication without
// importing the auth package.
func ApiRoutes(router *gin.Engine, db *pgxpool.Pool, q *queries.Queries) {
    api := router.Group("/api")

    authMW := auth.AuthRoutes(api, db, q)
    akademik.AkademikRoutes(api, db, q, authMW)
    keuangan.KeuanganRoutes(api, db, q, authMW)
}
```

```go
// internal/shared/web/auth_middlewares.go — the type lives here so any
// module can accept it without importing internal/auth.
type AuthMiddlewares struct {
    Session gin.HandlerFunc
    // Permission func(string) gin.HandlerFunc  // future
}
```

```go
// internal/auth/routes.go — auth owns the DB-backed implementation and
// returns the populated bag for cmd/app to distribute.
func AuthRoutes(router *gin.RouterGroup, db *pgxpool.Pool, q *queries.Queries) web.AuthMiddlewares {
    repo    := NewAuthRepository(db, q)
    service := NewAuthService(repo)
    handler := NewAuthHandler(service)

    sessionMW := SessionMiddleware(repo)

    auth := router.Group("/auth")
    auth.POST("/login",  web.Wrap(handler.Login))
    auth.POST("/logout", web.Wrap(handler.Logout))
    auth.GET("/me",      sessionMW, web.Wrap(handler.Me))

    return web.AuthMiddlewares{Session: sessionMW}
}
```

```go
// internal/akademik/routes.go — receives the bag via shared/web; does NOT
// import internal/auth.
func AkademikRoutes(router *gin.RouterGroup, db *pgxpool.Pool, q *queries.Queries, authMW web.AuthMiddlewares) {
    repo    := NewAkademikRepository(db, q)
    service := NewAkademikService(repo)
    handler := NewAkademikHandler(service)

    akademik := router.Group("/akademik", authMW.Session)
    akademik.GET("/mahasiswa", web.Wrap(handler.ListMahasiswa))
}
```

**Why does the auth module return middleware?** Some middleware (session validation, permission checks) need DB access — they hash the cookie token and look it up in `user_sessions`. That makes the *implementation* domain logic owned by `internal/auth/`. The *type* (`web.AuthMiddlewares`) lives in `internal/shared/web/` so other modules can accept it without importing `internal/auth/`. Auth populates it; everyone else consumes it.

**Why a struct instead of a single return value?** Future-proofing for more shared middleware (permission, role check, rate limit per user). Adding a field to `web.AuthMiddlewares` does not break existing callers; changing a single return type would.

**Why no `App` struct?** In this starter kit, service has no cross-cutting dependencies — only its repository. Repository absorbs all infrastructure. Cross-cutting concerns (logging, auth, rate limit) live in middleware, not injected into service. Adding a global wiring layer would be premature abstraction.

**When to introduce `shared.Infra`:** when a module needs more than 3–4 infrastructure handles, group them in a single struct and pass that struct down instead of growing the parameter list. Don't do this preemptively — wait until a module actually has many dependencies.

## Value vs pointer returns

Default to returning by value. Only return a pointer when the type is larger than ~64 bytes (roughly one cache line) or when nilability is genuinely meaningful to the caller.

- **≤ 64 bytes:** return by value. Copy is cheap, the value stays on the stack, no GC pressure, no nil checks at the call site.
- **> 64 bytes:** return a pointer. The copy cost outweighs the heap allocation, and a pointer avoids passing big structs through registers.
- **Lookup misses:** do not signal "not found" with `nil` just to justify a pointer return. Return a module-level sentinel error from repository (e.g. `ErrUserNotFound`) — see [Why translate errors at the repository boundary](#why-translate-errors-at-the-repository-boundary). This keeps the value-return ergonomic and removes the ambiguous `(nil, nil)` state.

The 64-byte threshold is a heuristic, not a hard rule — measure with `unsafe.Sizeof` when unsure. Most sqlc-generated `Row` structs for narrow lookups (id, hash, a few timestamps) fit under it; rows that embed many text columns may not.

## Why mandatory `// Tx step N` comments

Each `qtx.X(...)` call inside a transaction must be preceded by a one-line comment of the form `// Tx step N: <what & why>`. Reasons:

- Reader can scan the steps top-to-bottom and immediately see the business intent of the transaction.
- If a step fails, the comment makes it obvious which side effect is being rolled back.
- Reordering or inserting a step requires updating the numbering — that friction is intentional, it forces re-thinking the sequence.

Example:

```go
qtx := pg.repo.WithTx(tx)

// Tx step 1: insert into users — create the new user identity.
userID, err := qtx.InsertUser(ctx, ...)
if err != nil { ... }

// Tx step 2: insert into user_accounts — attach the password credential to the user.
_, err = qtx.InsertAccount(ctx, ...)
if err != nil { ... }
```

This rule applies only to queries inside a `Begin/Commit` block. Single-query repository methods do not need this comment — the function name already documents the intent.

## Why interfaces are scarce

Define interfaces only when writing tests that require mocking. Reasons:

- Premature interfaces fragment the type system and make navigation harder ("where is this implemented?").
- The standard refactor pattern in Go is "concrete first, extract interface when needed".
- Repository tests are integration tests — no mock needed.
- Service tests mock the repository via a small interface defined next to the test, not in the production code.

If you find yourself defining interfaces "just in case", you are designing for a use case that doesn't yet exist.

## Naming rationale

- `XxxRequest` / `XxxResponse` reserved for HTTP DTOs so the suffix carries a clear signal at the call site.
- `XxxInput` / `XxxArgs` for repository params to avoid name collision with sqlc-generated `XxxParams` — if both lived in the same package with the same name, you'd be forced to alias every import.
- Reuse sqlc `XxxParams` when the repository method is a thin pass-through; wrapping it adds no value and creates two types that must stay in sync.
