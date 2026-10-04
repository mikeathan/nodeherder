# Storage driver decision (CI repair)

Status: Implementation proposed for review; user authorized scoped replacement.

Bolt v1.3.1 crashes at Bucket.write under Go checkptr during race tests. Replace
the two imports with go.etcd.io/bbolt v1.4.3, whose Go 1.23 minimum preserves our
Go 1.24 baseline. bbolt maintains Bolt API compatibility; verify file compatibility
with a legacy-generated fixture rather than relying on that claim alone.
Sources: [bbolt compatibility](https://github.com/etcd-io/bbolt),
[v1.4.3 module requirements](https://github.com/etcd-io/bbolt/blob/v1.4.3/go.mod).

Alternatives: disabling checkptr hides the failure; vendoring/patching the old
driver adds maintenance; a different database adds unnecessary schema migration.
Keep buckets/keys/files and options unchanged. Required testify/x/sys updates are
transitive; no unrelated dependency updates. No new format features/compaction.

Rollout: stop NodeHerder, back up all configured database files/volumes before
binary/container replacement. Keep old binary/image plus backup. Restore the
stopped-service backup for rollback; fixture tests do not certify every real file.
No deployed files or household data are accessed in this change.
