# File discovery pages

The glob tool accepts offset (zero-based, default 0) and limit (default 200,
range 1–500). Results are sorted by path after applying privacy/ignore filters.
When more results exist, the response includes the next offset.

Example: {"pattern":"**/*.go","offset":200,"limit":200}

Offsets past the end return an empty page. Pagination assumes the matching file
set stays unchanged between calls; adding or deleting files can shift offsets.
This bounds the model response, not the underlying glob scan or its memory use.
