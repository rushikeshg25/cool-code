# Build a PostgreSQL connection pooler in Go

## Intent and finish line

Build enough of a PgBouncer-like pooler to explain how connection reuse works,
where it breaks, and how to verify it. You write the core logic; the assistant
sets exercises, supplies checks and scaffolding, and reviews the result.

The final demonstration uses **20 clients sharing at most 3 PostgreSQL backend
connections** under a deliberately limited workload. Show correct query results,
transactions staying on one backend, waiting when capacity is exhausted, and
recovery after a backend connection dies. Explain the ownership transitions in
the logs. These are proposed lab numbers, not expected production traffic.

Suggested budget: **six weeks at 4–6 hours/week**, with a review on
**2026-10-27** if starting now. This is a planning estimate, not a commitment.
If the core takes longer, reduce scope and keep a working session pooler as a
valid intermediate achievement.

## Scope

- One Go process, one local PostgreSQL 17 instance, one configured database and
  role, and a fixed pool size. Pin the database version for reproducible checks.
- Use Go's standard library for networking, framing, and synchronization. Use
  `psql` and, later, a Go PostgreSQL client as test clients. Do not put
  `database/sql` or an existing pooling library in charge of the pool itself.
- Begin with loopback-only connections, explicit `sslmode=disable`, and a
  disposable database configured for local trust authentication. This is a lab
  restriction; supporting real authentication is a later exercise.
- Initial transaction pooling supports simple-query messages and one outstanding
  query cycle per client. Use a controlled test workload. Extended protocol,
  pipelining, COPY, replication, and cross-transaction session features are outside
  that milestone's compatibility promise.
- No distributed coordinator, Kubernetes, web dashboard, failover, SQL parser,
  or production deployment. PostgreSQL remains responsible for storage and
  transaction durability; the pooler never automatically retries a query whose
  execution outcome is unknown.

## The concept that holds everything together

A **client connection** is an application's connection to our process. A
**backend connection** is our connection to PostgreSQL. A **lease** gives one
client exclusive use of one backend. The **pool** owns backend capacity and
waiting clients.

In session mode, a client holds the lease for its entire connection. In
transaction mode, it holds the lease through a transaction and gives it back
between transactions. A long-lived idle client can therefore stop occupying a
backend in transaction mode.

```text
psql / test clients -> Go pooler -> limited backend connections -> PostgreSQL

backend: idle -> leased -> cleaning -> idle
                    |          |
                    +----------+-> discarded
```

Key invariants:

1. A backend has at most one owner.
2. Capacity counts idle, leased, cleaning, and connections being established;
   concurrent dials must not exceed the configured limit.
3. `ReadyForQuery(I)` means the backend is outside a transaction;
   `ReadyForQuery(T)` and `ReadyForQuery(E)` require keeping the lease.
4. Idle transaction status is necessary but insufficient for reuse: the full
   response cycle must be consumed and the chosen reset must finish successfully.
5. Broken or uncertain connections are discarded. Cleanup failure never returns
   a backend to the idle pool.

## Milestones

### 0. Establish the baseline — about one session

**Learn:** what a PostgreSQL session is and what connection pooling saves.

Create the independent Go module and a reproducible local database setup. Use two
`psql` sessions to run `SELECT pg_backend_pid()`, `BEGIN`, `COMMIT`, `ROLLBACK`,
and a failing statement inside a transaction. Observe them through a separate
admin connection to `pg_stat_activity`.

**Done when:** you can explain why two connected clients normally occupy two
backends and why an error inside an explicit transaction needs recovery.

**First prediction:** if two `psql` sessions connect directly to PostgreSQL, will
their `pg_backend_pid()` values match, and will either value change after COMMIT?

### 1. Build a transparent TCP proxy — about two sessions

**Learn:** byte streams, partial reads and writes, bidirectional I/O, EOF, and
goroutine ownership.

Accept a client, dial one backend, and relay both directions. Define how each
goroutine terminates when its peer disconnects. At this stage, PostgreSQL still
handles the entire startup and authentication exchange.

**Done when:** `psql` can run queries through the proxy and disconnect cleanly.
Repeated connect/disconnect cycles and a killed backend leave no accumulating
connections or stuck goroutines. This milestone intentionally has one backend
per client.

### 2. Understand and frame the wire protocol — about two sessions

**Learn:** framing over TCP, network byte order, protocol phases, and bounded
allocation.

Implement a small codec. Startup packets use a length and protocol/request code;
normal messages use a one-byte type and a length that includes the length field
but excludes the type byte. Handle short reads with exact-read semantics. Reject
invalid or oversized lengths before allocating.

Trace message types and sizes, not passwords or arbitrary query contents.
Observe `Query`, `ErrorResponse`, and `ReadyForQuery`, including I/T/E states.
Handle or explicitly reject startup variants; never decode an SSL request as an
ordinary typed message.

**Done when:** checks cover fragmented headers/bodies, several messages in one
read, truncated input, malformed lengths, and bounded fuzz input. Explain why
one `Read` cannot be assumed to return one PostgreSQL message.

### 3. Build a reusable session pool — about three sessions

**Learn:** resource ownership, bounded capacity, waiting, and sanitizing state.

This is the major architectural change: the pooler now handles client startup
and separately establishes reusable backend sessions. For the fixed local trust
configuration, provide the required startup responses and validate the supported
database, role, and startup parameters. Authentication cannot simply be replayed
from an earlier client's exchange.

Do not expose a reusable backend's real PID/secret in client BackendKeyData.
Until cancellation routing is implemented, use a separate frontend identity and
close unsupported CancelRequest connections without forwarding them. Document
that query cancellation is unavailable in this milestone.

Start with a pool of one. Hold the lease until the client disconnects. Intercept
client termination instead of forwarding it and closing the reusable backend.
Use a conservative cleanup policy: discard a connection on an unexpected or busy
disconnect; only reset a known idle connection with `DISCARD ALL`, consuming all
reset responses before making it available. Discard on reset failure.

Add a bounded wait queue and acquisition deadline. Define how a cancelled waiter
and a simultaneously available lease are resolved without losing capacity.

**Done when:** client B waits while client A is connected; after A exits normally,
B can reuse the same backend PID. A session-setting or temporary-table change
from A does not leak into B. Capacity stays bounded under concurrent acquisition.

### 4. Add transaction pooling — about three sessions

**Learn:** a protocol state machine and the difference between client lifetime
and transaction lifetime.

Acquire on a simple-query cycle. Forward its responses, and use PostgreSQL's
`ReadyForQuery` state to decide ownership; do not infer transaction completion
from SQL text or `CommandComplete`. An implicit autocommit query can finish a
lease; an explicit transaction keeps it across query cycles.

For this learning version, reset after each completed idle cycle before reuse.
This intentionally trades performance and session compatibility for a smaller
model. It is not PgBouncer's default transaction-mode reset policy. SQL session
state is not preserved across leases. Do not pretend a keyword blacklist can
fully enforce a stateless SQL subset.

Reject unsupported typed frontend messages explicitly. Simple-query SQL can
still enter COPY mode: on an unsupported COPY response, terminate the client
conversation and discard the backend safely. Maintenance/reset responses must
never leak into the next client's query results.

**Done when:** with a pool of one, A runs BEGIN and a failing query; B continues
waiting through both T and E states. A's ROLLBACK reaches I, cleanup completes,
and B proceeds. Also test COMMIT, autocommit, and multiple SQL statements inside
one simple Query message. An error alone never triggers release.

### 5. Make failures observable — about two sessions

**Learn:** backpressure, liveness, timeouts, and the difference between losing a
connection and knowing whether a transaction committed.

Exercise backend death, client death mid-transaction, a full queue, a timed-out
waiter, a slow reader, malformed input, and shutdown with leased connections.
Discard uncertain backends and allow later acquisitions to create replacements.
Drain on shutdown with a bounded deadline.

Add lightweight counters for idle/leased/cleaning backends, waiting clients,
acquisition timeouts, and discarded connections. Use controlled tests and Go's
race detector; use channels or explicit events rather than arbitrary sleeps to
coordinate test scenarios.

**Done when:** failure scenarios complete within test deadlines, preserve the
ownership/capacity invariants, and leave the pool usable by a fresh client.

### 6. Demonstrate and compare — about two sessions

**Learn:** measuring the tradeoff instead of assuming pooling makes SQL faster.

Run the final 20-client/3-backend workload. Compare direct PostgreSQL, our pooler,
and PgBouncer in equivalent modes using the supported simple-query workload.
Measure throughput, connection setup time, queue wait, end-to-end p50/p95 latency,
backend count, and memory. Separate connection-churn tests from long-lived-client
tests and record configuration and reset-policy differences. Exclude the admin
observer connection when checking the pool's backend limit.

**Done when:** save reproducible commands, results, and a short explanation of
where time went. Correctness and understanding are the goal; no arbitrary
throughput target is required.

## Keep the design small

Let modules emerge as behavior demands them:

- `wire`: read/write bounded protocol frames over an I/O stream.
- `pool`: acquire an exclusive lease with a deadline and finish it safely;
  capacity accounting, waiters, reset, and discard stay behind this interface.
- `proxy`: own the client conversation and transaction state transitions.

Choose concrete method signatures during their milestone. Avoid exposing pool
internals through getters or introducing interchangeable adapters before a real
second implementation is useful. Test behavior through each module's interface.

## How each learning session works

1. Explain one mechanism and its invariant in a few minutes.
2. Ask you for one concrete prediction before running anything.
3. Provide a small TODO and a behavioral check; you implement the logic.
4. Run checks and review correctness, especially prediction mismatches.
5. Compare the relevant behavior with PostgreSQL or PgBouncer.
6. Try one variation and explain the idea back in three lines.

Review asks: who owns this connection, what releases it, what happens if I/O
fails here, and what experiment proves the answer? Add missing prerequisites to
`learning/gaps.md` only when they actually surface.

## Overbuild check

**Verdict: build as a capped learning project.** The custom pool is justified by
the learning goal. For an application needing pooling today, PgBouncer or a
driver's pool is the simpler alternative.

The proposed lab has 20 clients, 3 backends, a tiny fixture database, and no
production traffic or durability obligation beyond PostgreSQL itself. Query rate
and latency are measurements to gather, not assumed requirements. A local
PostgreSQL process or container supplies the database; a single Go binary supplies
the pooler; console output supplies observability. Revisit richer tooling only
when a concrete exercise cannot be observed or reproduced with these tools.

**Cost assumption:** use existing hardware, with no paid infrastructure. Expected,
idle, and 10x-client experiments have zero additional cloud charges under this
setup; electricity, existing hardware, and any existing software licenses are
excluded. At 10x load, keep the backend cap and measure queuing or rejection.
Budget less than one hour/week for environment maintenance.

**Pause criteria:** the project requires paid hosting, setup consumes more than
two sessions, or the six-week review arrives without a repeatable session-pooling
demo. Reduce scope to the last passing milestone, preserve the explanation and
checks, and reconsider the next experiment. Keep the local run instructions,
trace, and benchmark results; stop the local database when finished.

## Stretch exercises, after the core demo

Choose one at a time: extended query/Sync error recovery; prepared statements
across reassignment; frontend versus backend SCRAM authentication; TLS; correctly
routed CancelRequest with protection against stale cancellation; or multiple
database/user pools. COPY and pipelining each need their own protocol exercise.
These are separate projects in correctness, not a checklist for the initial MVP.

## Primary references

See [the verified protocol notes](PROTOCOL_RESEARCH.md) for source-linked facts
and their design implications.

- [PostgreSQL protocol overview](https://www.postgresql.org/docs/17/protocol-overview.html)
- [PostgreSQL message flow](https://www.postgresql.org/docs/17/protocol-flow.html)
- [PostgreSQL message formats](https://www.postgresql.org/docs/17/protocol-message-formats.html)
- [DISCARD](https://www.postgresql.org/docs/17/sql-discard.html)
- [PgBouncer pooling compatibility](https://www.pgbouncer.org/features.html)
- [PgBouncer configuration and reset policy](https://www.pgbouncer.org/config.html)
