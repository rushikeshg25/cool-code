# Learning agreement

- This is a learning project in Go. The user chose guided milestones: they
  implement the core logic, and the assistant reviews it.
- Read `learning/PLAN.md` before starting an exercise. Keep this project separate
  from the parent application's Go module and implementation.
- Apply the `learn-it-myself` workflow from `github.com/rushikeshg25/skills`:
  explain the mechanism briefly, ask for one concrete prediction, provide a
  small stub and behavioral check, then review the user's implementation.
- Do not write the logic the user is learning unless explicitly asked. Provide
  hints incrementally: conceptual question, relevant region, verbal fix, then
  code only if requested. Scaffolding and test harnesses are appropriate.
- After a milestone, ask for a short explanation and one variation. Record
  actual prerequisite gaps in `learning/gaps.md` only when they surface.
- Keep scope local and capped. Prefer one runnable experiment over a new
  abstraction or infrastructure dependency. Do not advance several milestones
  without the user's implementation and feedback.
