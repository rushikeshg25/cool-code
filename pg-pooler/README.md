# PostgreSQL pooler, built to learn

A learning project in Go: build a TCP proxy, understand the PostgreSQL wire
protocol, then reuse database connections safely across clients.

**Working agreement:** you implement the core logic; the assistant explains,
provides small exercises and checks, and reviews your work. Implementation has
not started.

Start with [the learning plan](learning/PLAN.md). Each milestone has an
observable result and a correctness check. This directory will become its own
Go module when the first exercise is scaffolded; it is separate from the parent
`cool-code` application.

## Learning goals

- Understand TCP streams, partial reads and writes, and connection lifecycles.
- Decode PostgreSQL startup and query messages.
- Manage a bounded pool with exclusive connection ownership and waiting clients.
- Explain session pooling versus transaction pooling.
- Handle failed transactions, cleanup, disconnects, and backpressure.
- Compare observable behavior and performance with PgBouncer.

## Milestones

| Step | Build or experiment | Evidence it works |
| --- | --- | --- |
| 0 | Observe direct PostgreSQL sessions | Explain backend PIDs and transaction states |
| 1 | Transparent TCP proxy | Run queries through it with `psql`; disconnect cleanly |
| 2 | Wire-protocol codec | Handle fragmented messages and reject malformed lengths |
| 3 | Session connection pool | Reuse a backend after disconnect without leaking session state |
| 4 | Transaction pooling | Keep transactions pinned; release only after safe cleanup |
| 5 | Failure handling | Preserve capacity and recover after client or backend failure |
| 6 | Demonstration and comparison | Run 20 clients through at most 3 backends and compare with PgBouncer |

## Initial scope

Use Go, one local PostgreSQL 17 instance, one database and role, and a fixed pool
size. Begin with a disposable, loopback-only lab using trust authentication and
no TLS. The first transaction-pooling exercise uses simple queries with one
outstanding query cycle per client.

The central rule is that a backend has at most one owner. PostgreSQL's
`ReadyForQuery` states `T` (in a transaction) and `E` (failed transaction) keep
the connection pinned. State `I` makes it eligible for reuse only after the
response cycle and cleanup finish successfully.

TLS, SCRAM authentication, extended queries, prepared statements, cancellation
routing, COPY, and multiple database/user pools are later exercises. The initial
goal is a small, understandable learning implementation, not production
compatibility.

## How we work

For each milestone, the assistant explains one mechanism and asks for a concrete
prediction. You implement a small exercise; the assistant supplies behavioral
checks, reviews correctness, and offers incremental hints. Finish with one
variation and a short explanation in your own words.

Suggested effort is six weeks at 4–6 hours per week, adjustable to your pace.
Keep everything local with no paid infrastructure. A working session pooler is
a useful stopping point if transaction pooling needs more time.

## Further details

- [Full plan and acceptance criteria](learning/PLAN.md)
- [Protocol research and primary sources](learning/PROTOCOL_RESEARCH.md)
- [Learning agreement for coding assistants](AGENTS.md)

## Skills used

The approach uses
[learn-it-myself](https://github.com/rushikeshg25/skills/blob/2412dc3ddc3cb541837372881b3334b67421c9db/skills/learn-it-myself/SKILL.md)
and
[overbuild-check](https://github.com/rushikeshg25/skills/blob/2412dc3ddc3cb541837372881b3334b67421c9db/skills/overbuild-check/SKILL.md)
from `rushikeshg25/skills`. These were read and applied to the plan, not installed
globally.
