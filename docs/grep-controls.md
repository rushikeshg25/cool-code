# Search controls

The grep tool accepts literal (default false), ignoreCase (default false), and
limit (default 100, range 1–200). Existing pattern, path and include fields retain
their meanings; include remains a regular expression over file basenames.

Example: {"pattern":"user[id]","literal":true,"ignoreCase":true,"limit":20}

Each matching line is shortened to 1,000 UTF-8-safe bytes. Result and line
truncation are marked explicitly. Narrow path, include or pattern when results
are truncated. Privacy and ignore rules still apply. File reads stop after the
result cap is reached, though directory traversal may still continue.
