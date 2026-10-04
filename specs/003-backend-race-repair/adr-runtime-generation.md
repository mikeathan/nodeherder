# Ready automation generations

Status: implementation decision for FR-20, authorized direction by user; no policy
ratification or exception approval.

Disk loading currently exposes unconfigured actions, and unchanged schedule jobs
retain the old automation instance. A reload lock would drop physical events or
block callback re-entry. Configuring inside storage's lock has the same deadlock risk.

Keep one engine runtime registry of prepared generations; storage is a persistence
boundary, not an execution registry. Prepare transactional handler changes, activate,
persist when required, publish, then retire old handler resources. Rollback leaves
the last working generation active. Reads snapshot pointers and release locks before
callbacks. Concurrent management writes fail busy; they are not queued or retried.

Cost: a runtime map and brief pointer-read lock; staging temporarily owns old/new
schedules. Old accepted evaluations/delayed actions may finish; MQTT acknowledgement
and global event-context isolation are unchanged. Reload publishes independent IDs,
not an all-household transaction. Invalid disk scans fail closed without removals;
individual configuration failures retain their previous runtime and schedule.

Handler preparation adds an internal lifecycle contract (activate/rollback/complete)
so failed configuration or persistence cannot leave replacement timers running.
Existing Process remains a convenience wrapper. No dependency/schema migration.

User-approved explicit-disable exception: an observed disabled Z2M source removes
its ready entry before handler retirement. Cleanup errors remain observable/retryable
on reload but cannot restore execution. Preserve disk recipes for successful later
re-enable. Ordinary failures retain working generations; accepted commands remain
non-joining. No new polling or device-state policy.
