// Package integrations coordinates user-configured external information independently of devices.
//
// A compiled host YAML definition describes a bounded search/select recipe. Setup owns private
// inputs and address choices; a durable Operation pins its definition, generation and exact POST
// bytes before a worker performs HTTP. The worker normalizes a provider response into a typed
// capability and publishes the result only after its repository transaction commits.
//
// Service owns the worker queue, cancellation and cleanup timer. Its mutex protects admission and
// local bookkeeping; provider I/O runs outside that lock. Generation checks reject responses from
// deleted, reconfigured or credential-replaced owners. M1 performs no recurring provider requests:
// only explicit mutations admit work, while GETs project already stored data.
//
// The repository owns encryption and atomic storage; operations owns generic HTTP/extraction;
// capabilities owns waste semantics; transport DTOs intentionally exclude private attempt state.
package integrations
