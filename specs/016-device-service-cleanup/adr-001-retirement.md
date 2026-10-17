# ADR 001: Retire on the device lane with an output fence

Status: Proposed implementation decision | 2026-10-16

Updates are lane-owned, but live timer publication runs under the publisher mutex.
Deleting storage first can be undone by an in-flight timer's snapshot write. Closing
first and then attempting deletion would disable a still-existing device on failure.

Route successful bridge removal to the existing device lane. Reuse the publisher's
output mutex to run deletion behind current output and close only when deletion
succeeds. Then cancel/join monitoring and evict the processor generation. Registry
locks never surround callbacks/I/O. An owned child context permanently retires old
references and removes parent-context registrations. No retained tombstones; existing
rediscovery creates a new generation. Publisher.Close remains the common timer and
shutdown-registration cleanup path.

Alternatives rejected: store-first/close-first races above; moving all timer output
onto lanes introduces queue/retry policy beyond this cleanup; per-device operation
locks risk callback re-entry deadlocks and duplicate the existing lane ownership.
