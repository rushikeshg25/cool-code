# Execution, editing, and completion evidence

Shell, test, lint and formatter results include JSON fields for success, exit
code, timeout, cancellation, command and working directory. Each output stream
is limited to 24,000 bytes with an explicit truncation flag. The environment
label identifies the harness host, not the destination of nested Docker or SSH
commands.

The edit_file tool requires exactly expected_replacements matches (default 1).
Missing, ambiguous or nonpositive counts fail without modifying the file.
This is not a concurrent-edit version guard.

Compaction includes tool calls and results and uses the same user-turn boundary
for summarizing and removing messages. Individual transcript entries are bounded
and marked when truncated. The handoff asks the model to preserve pending
requirements, verification, blockers and Git state. Summaries still depend on the
model and do not guarantee that every requirement is retained.

Workflow guidance asks for consistent project setup, relevant checks, bounded
repair attempts, and completion claims supported by tool evidence. It respects
mode and permission limits and does not automatically run commands.

The shared behavior is covered by deterministic tool and processor regressions,
including session restore followed by a second compaction. CLI proxy integration
and detailed tracing remain exclusive to dev.
