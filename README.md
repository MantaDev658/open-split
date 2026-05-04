# Open Split

A self-hosted, open-source expense splitting API. Track shared costs, manage groups, simplify debts, and settle up — without sending your data to a third party.

---

## Features

- **JWT authentication** — stateless auth via signed tokens; no session storage
- **Group management** — create groups, add/remove members with outstanding-balance guards
- **Four split strategies** — even, exact, percentage, and proportional shares
- **Penny-perfect math** — a custom `money` engine distributes remainders deterministically; no floating-point drift
- **Debt simplification** — greedy algorithm reduces N bilateral debts to the minimum number of transactions
- **Audit log** — every mutation (expense created/updated/deleted, member added/removed, debt settled) is recorded per group
- **Cursor-based pagination** — all list endpoints are paginated with `?limit` and `?cursor`
- **CSV CLI** — standalone binary for offline expense parsing from a ledger file

---

## Architecture

Go workspace monorepo. Two modules, two binaries.

```
open-split/
├── apps/backend/
│   ├── cmd/api/            # REST API server (main binary)
│   ├── cmd/cli/            # CSV expense parser (standalone binary)
│   └── internal/core/
│       ├── domain/         # Entities, allocation logic, error types
│       ├── application/    # Use-case services (ExpenseService, GroupService, UserService)
│       └── infrastructure/
│           ├── http/       # Handlers split by service, auth middleware
│           ├── postgres/   # Repository implementations, migrations, partition manager
│           └── csv/        # CSV ledger parser
├── libs/shared/money/      # Zero-dependency money type (int64 cents)
└── go.work                 # Go workspace linking both modules
```

Domain layer has no external dependencies. Infrastructure depends on domain, never the reverse.

---

## Prerequisites

- Go 1.21+
- Docker (for the local PostgreSQL instance)
- `make`
- `golang-migrate` for schema migrations: `make setup-migrate`
- `golangci-lint` for linting: `make setup-lint`

---

## Getting Started

```bash
git clone https://github.com/yourusername/opensplit.git
cd opensplit

# Start postgres and apply migrations
make db-up
make migrate-up

# Run the API (listens on :8080)
make run-api
```

Set the following environment variables (or create a `.env` file):

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret used to sign and verify JWT tokens |

Example `.env`:
```
DATABASE_URL=postgresql://postgres:password@localhost:5432/opensplit?sslmode=disable
JWT_SECRET=change-me-in-production
```

---

## API Reference

All endpoints except `/auth/*` require a `Bearer` token in the `Authorization` header.

Currency amounts are always in **cents** (integer). `$10.00` = `10000`.

Paginated endpoints accept `?limit=N` (default 20, max 100) and `?cursor=<RFC3339Nano timestamp>`. Responses include `"next_cursor"` when more pages exist.

### Auth

| Method | Path | Description |
|---|---|---|
| POST | `/auth/register` | Create a user account |
| POST | `/auth/login` | Authenticate and receive a JWT |

### Expenses

| Method | Path | Description |
|---|---|---|
| POST | `/expenses` | Create an expense; payer is set from the JWT |
| GET | `/expenses` | List expenses; filter by `?group_id=` |
| PUT | `/expenses/{id}` | Update an expense |
| DELETE | `/expenses/{id}` | Delete an expense |

### Balances & Settlements

| Method | Path | Description |
|---|---|---|
| GET | `/balances` | Net balances and suggested settlements; filter by `?group_id=` |
| GET | `/friends/{user_id}/balances` | Bilateral friend balances for the authenticated user |
| POST | `/settlements` | Record a debt payment between two users |

### Users

| Method | Path | Description |
|---|---|---|
| GET | `/users` | List all users |
| PUT | `/users/{id}` | Update display name |
| DELETE | `/users/{id}` | Soft-delete a user |

### Groups

| Method | Path | Description |
|---|---|---|
| POST | `/groups` | Create a group; creator is set from the JWT |
| GET | `/groups` | List groups for the authenticated user |
| PUT | `/groups/{id}` | Rename a group |
| DELETE | `/groups/{id}` | Delete a group |
| POST | `/groups/{id}/members` | Add a member |
| DELETE | `/groups/{id}/members/{user_id}` | Remove a member (blocked if outstanding balance) |
| GET | `/groups/{id}/activity` | Paginated audit log for a group |

### Request bodies

**POST /auth/register**
```json
{ "id": "alice", "display_name": "Alice", "password": "..." }
```

**POST /expenses**
```json
{
  "group_id": "optional-group-uuid",
  "description": "Dinner",
  "total_cents": 9000,
  "split_type": "EVEN",
  "splits": [
    { "user_id": "alice" },
    { "user_id": "bob" }
  ]
}
```

**POST /settlements**
```json
{ "receiver_id": "bob", "amount_cents": 5000, "group_id": "optional" }
```

---

## Split Strategies

| Strategy | `split_type` | `splits` format | Example |
|---|---|---|---|
| Even | `EVEN` | `[{"user_id": "alice"}, ...]` | Total divided equally; remainders go to first participant |
| Exact | `EXACT` | `[{"user_id": "alice", "value": 2000}, ...]` | Values must sum to `total_cents` |
| Percentage | `PERCENT` | `[{"user_id": "alice", "value": 25}, ...]` | Values must sum to 100 |
| Shares | `SHARES` | `[{"user_id": "alice", "value": 2}, ...]` | Proportional to share counts |

---

## CLI Tool

Parse a CSV ledger file and print settle-up instructions:

```bash
make run-cli
# or directly:
cd apps/backend && go run cmd/cli/main.go -file=../../test_expenses.csv
```

CSV format — one expense per row, minimum 7 columns:

```
Date,Description,Category,TotalCents,Payer,Strategy,Participants...
2026-04-01,Dinner,Food,9000,Alice,EVEN,Alice,Bob,Charlie
2026-04-02,Hotel,Lodging,30000,Charlie,EXACT,Alice:10000,Bob:10000,Charlie:10000
```

Participant columns follow strategy syntax: `Name` for EVEN, `Name:Value` for EXACT/PERCENT/SHARES.

---

## Development

```
make check          # build + lint (run before every commit)
make test-unit      # unit tests with coverage
make test-race      # race detector
make test-fuzz      # 30s fuzz run on domain package
make test-integration  # postgres integration tests (requires TEST_DB_URL)
make test           # all of the above (spins up Docker)
make migrate-up     # apply pending migrations
make migrate-down   # roll back last migration
```

The linter (`golangci-lint`) enforces doc comments on all exported identifiers, error wrapping, and a no-shadow rule. Run `make check` before opening a PR.

---

## Contributing

1. Fork the repo and create a branch from `main`.
2. Make your changes. `make check` must pass clean.
3. Include tests for any new behaviour.
4. Open a pull request — describe what changed and why.

---

## License

[AGPL-3.0](LICENSE)
