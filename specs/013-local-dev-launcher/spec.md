# Local development launcher

Status: Complete

FR-1: One command from any working directory builds the native backend and runs it
alongside the frontend development server, using existing configuration.
FR-2: Ctrl+C/SIGTERM or either service exiting stops all owned process trees;
startup failures return nonzero. Shutdown is bounded to five seconds before SIGKILL.
FR-3: No dependency installs, Docker changes, environment-file edits, or new packages.
Scope: macOS/Linux development tooling only; backend hot reload is excluded.
Acceptance: isolated fake commands verify working directories, command ordering,
failure propagation, signal cleanup, and descendant cleanup without household access.
