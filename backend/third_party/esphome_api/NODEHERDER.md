# NodeHerder local ESPHome client patch

Based on `github.com/mycontroller-org/esphome_api` v1.4.0, tag commit
`64fb66c8598cf7f48a9139f8e899bd5b8b1553a3`; original Apache-2.0 license is in
`LICENSE`. Source and dependency history: https://github.com/mycontroller-org/esphome_api.

This is an incomplete local hardening effort, not an independently audited or production-ready
client. The upstream module declares Go 1.26; the fork currently declares Go 1.24.0 after the
full upstream package tree compiled and its current tests passed using NodeHerder's Go 1.24.8.
That evidence is not a guarantee of compatibility with future upstream changes.

Changes so far: close the socket on setup failure; add context-aware, bounded dial/handshake;
synchronize the last-message timestamp; reject plaintext frames over 64 KiB before allocation;
reject truncated decrypted messages; reject concurrent waiters for the same response type; and
unblock response waiters when the reader exits. Close is idempotent and context-bounded, and
encrypted writes serialize nonce advancement with frame delivery. Outbound frames are bounded and
short writes are reported. Handshake reads are bounded, validate protocol selection and do not
consume subsequent application frames. Companion tests exercise these cases with local fake peers, including
a Noise responder that checks concurrent encrypted writes. That responder uses the same Noise
library and is not an independent implementation or proof of device interoperability. More work
is required for encrypted-frame validation, unknown message handling, independent Noise
interoperability, discovery, commands and reconnect.
See `specs/038-iot-control-center/` in the parent repository.

Before updating: compare upstream changes and licenses, review the patch and protocol version,
reproduce security regressions, run this module's tests and race checks with the project's
toolchain, and recheck Docker, CI, SBOM and dependency-scanning behavior. Do not change or discard
the original license. No changes have been published upstream.
