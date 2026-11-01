# ESPHome client evaluation: v1.4.0

Date: 2026-10-30 | Candidate: `github.com/mycontroller-org/esphome_api@v1.4.0` | Go: 1.26.1

## Setup and result

An isolated temporary Go module used the candidate directly. Its test peer bound `127.0.0.1:0`; no household device or external service was contacted by the tests. Run: `go test -race -count=1 ./...`. **Result: failed**, as expected for two adversarial lifecycle probes:

1. `GetClient` with an invalid key returned an error after connecting; the accepted peer observed a read timeout after 150 ms instead of EOF. The client did not provide a handle to the caller to close that socket.
2. A fake plaintext peer sent sensor-state messages while a test goroutine read `LastMessageAt`. Go's race detector identified a write in `pkg/client/client.go:275` concurrent with the getter at line 161.
3. `go test -count=1 -run '^TestOversizedFrame$'` sent a maximal unsigned frame length from a loopback peer. The reader panicked at `pkg/connection/conn_plaintext.go:99` (`makeslice: len out of range`), terminating the test process. This validates a malformed plaintext peer can crash a process using the client; it does not characterize encrypted framing.

The test source and temporary module are at `/private/tmp/nodeherder-esphome-eval/`; they are not part of NodeHerder's normal suite. The selected client and production integration remain undecided. The run did not establish Noise interoperability, real firmware compatibility, reconnect correctness or command coverage.

## Source review

In v1.4.0, `GetClient` uses `net.DialTimeout`, returns directly on transport-construction or handshake failure, and starts its reader only after handshake. `Close` waits for a protocol response before stopping its reader. Plaintext framing reads a peer-supplied unsigned length into an allocation before an explicit bound, with the demonstrated panic above. The release registers a response waiter before sending; an earlier observation about the web-rendered main branch does not apply to v1.4.0. Cancellation, concurrent request handling and successful/failed Noise handshakes remain to be tested.

## Acceptance gate for any client

- Context cancellation interrupts DNS/dial, handshake, reads and shutdown; every failure closes owned sockets and joins workers.
- Incoming frame lengths and decoded metadata are bounded before allocation; unknown messages do not corrupt state or silently stop supported subscriptions.
- Noise NNpsk0 authenticates the supplied key and never silently downgrades; plaintext is an explicit configuration choice.
- Entity discovery and state subscriptions recover after disconnect without duplicating sessions; the required commands retain confirmation semantics.
- Local fake-peer success and failure cases, malformed frames and concurrency tests pass with `-race`; hardware compatibility remains separately recorded evidence.

The repository may wrap a client only if the wrapper can actually enforce these properties; a wrapper cannot reliably repair resource ownership or unbounded allocations hidden inside a dependency. A patched dependency/fork, alternative library or independently implemented transport each require a scoped dependency ADR and maintenance cost comparison.

## Maturity constraint

The user requires a mature and stable solution. As of 2026-10-30 no evaluated Go client has demonstrated that standard: the preferred package fails the probes above, the newer `go-aioesphomeapi` is untagged and explicitly partial, and `Richard87/esphome-apiclient` has limited observable history and adoption. Do not select any merely because it exists. The upstream [aioesphomeapi](https://github.com/esphome/aioesphomeapi) client is actively maintained and used by Home Assistant, but it is Python and its integration would need a deliberate deployment/IPC decision. These assessments are a snapshot, not guarantees of security or completeness.
