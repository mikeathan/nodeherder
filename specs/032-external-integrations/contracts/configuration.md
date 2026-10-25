# Host YAML definitions and a shared editing UI

Status: M1 implementation contract; later-milestone provisions remain deferred. See [implementation review](../implementation-review.md).
Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md)

## User decisions and scope

Every authenticated household user can add, view, edit, test, refresh, reload and remove integrations
and their definitions. There are no feature-specific roles, allowlists or private per-user setups.
Keep the existing login/session boundary; this does not expose editing to unauthenticated callers.
Users share the configured household and can continue each other's setups. created_by/updated_by
is audit metadata only. Revisions, not user ownership, prevent simultaneous edits losing work.

The user explicitly requires desktop/mobile simplicity and UI-editable provider configuration.
Recommend YAML files on the host as the definition source of truth. The user favours configurations
similar to Home Assistant; this design does not claim Home Assistant stores every UI integration
as user-editable YAML. Its docs distinguish UI configuration and file configuration.
[Home Assistant configuration](https://www.home-assistant.io/docs/configuration/)

## What lives where

| Location | Contents | Version control |
| --- | --- | --- |
| configs/integrations/<type>/<provider>.yaml | Provider identity/capability, explicit setup/refresh links, requests, typed mappings and non-secret defaults | Designed to commit/share after review |
| DATA_DIR/integrations.db | Household instances, encrypted property/binding/setup/credential records, runtime cache/health, active compiled definition hashes and retained revisions | Runtime/private; exclude from Git |
| External INTEGRATION_KEYRING_FILE | Keys for private DB records | Never commit; backup separately |
| UI's unsaved edit buffer | Current form/YAML changes, expected file hash | Memory only; no localStorage |

The accepted example path is configs/integrations/waste_collection/whenisbins.yaml. Type/provider
IDs match directory/document identity; callers never supply arbitrary filesystem paths.
The definition key is the (integration_type, provider id) pair; identical provider filenames in
different type folders cannot overwrite the same definition/runtime binding.
Default definition directory is relative to the backend working directory, matching the existing
configs/automations convention. Optional INTEGRATIONS_CONFIG_DIR overrides the root. Current backend
Compose mounts ./backend/configs at /backend/configs, so the default host directory persists without
adding a new volume. Deployment docs must verify it is writable. Unlike EnsureDir's generic fallback,
never silently save integration definitions in a temporary directory: report config_directory_unwritable.

Supported configuration controls keep selected addresses, credentials and fetched data outside
provider definitions. Fixed literals/comments manually pasted into YAML can still contain private
data; validation cannot guarantee their semantic absence. Explain that these values are exported
and require intentional review; never log raw definitions. Credentials use a declared secret slot bound privately per instance. A user
can view/edit the instance's normal settings through the UI; those settings are intentionally outside
the shareable provider YAML. Export only definitions, not decrypted household records. No arbitrary
filesystem/file-browser API and no secret-file editor are needed.

Ship reviewed example/default definitions. On first installation copy a missing default into the
writable config directory explicitly; updates never overwrite an existing user's file. Show “update
available” and a diff when a shipped default changes. Keep stable IDs independent of file display name.

## One source of truth, two editors

The connection detail page has Results, Setup, Configure operations and Help, usable by every signed-in
user. Normal setup stays postcode → address → save; no endpoint knowledge is needed. Configure operations has:

- **Form**: labelled requests/fields/mappings, inline errors, useful defaults and help.
- **YAML**: editable text of the same document, with validation messages showing line/column/field.
- Validate, Test, Save changes, Reload from host and Restore previous actions.
- A Help tab/page within Configure operations, including contextual field/error links and a quick
  start/operation reference; see the [help contract](operations-and-help.md).

Switching views operates on the same parsed document. Do not create separate form and YAML stores.
Use a node-based YAML representation to retain comments/field order where feasible. Form edits may
normalize formatting; preview those changes and do not promise byte-identical round trips. If YAML
is invalid, keep the buffer available and block Form/save until corrected instead of discarding it.
No hidden unsupported fields: strict parsing rejects them with a useful message.

On mobile use one-column forms, collapsible request sections, full-width inputs, touch targets at
least 44px, clear labels and reachable save/error feedback. The YAML panel can scroll internally;
the page must not overflow horizontally. Users can repair supported fields through Form without
typing YAML on a phone. Test at 390px and 1280px, keyboard/focus, offline/pending and concurrent-edit
conflicts. New editor dependencies need demonstrated accessibility/bundle benefit; a native textarea
with line-numbered validation is sufficient for the first release.

## Safe file saving and activation

Saving is a configuration action, never a Git commit. The UI gets exact YAML bytes plus content_hash,
active_hash and revision/status. PUT submits YAML and the expected disk hash (If-Match). All application
writers for one definition are serialized. Re-read disk before writing; a detected external/UI change
returns 409 with a diff/reload choice rather than overwriting silently.

A new definition uses validated fixed-directory type/provider IDs and If-None-Match: * so it cannot replace an
existing file accidentally. A Copy definition action can reuse a supported recipe/capability.
Reload definitions rescans only this configured directory, including files added on the host;
it does not install code or fetch packages. Unsupported recipe/capability names show a useful error.

Validate size/encoding/YAML/schema/references/typed output/origin rules and compile before changing
the file. Draft edits are held in the browser until Save; Test validates supplied draft in memory
without activating it. A test that submits provider work uses the normal idempotency/limit contract.

Persist a revision journal containing the previous/new hash and compiled snapshot before the file
rename. Write a temporary file in the same directory, flush it, atomically rename into place and
flush the directory where supported; restrict definition IDs/path traversal and reject symlink paths.
Do not accept a browser-supplied absolute path. After rename, commit the new active compiled hash
and instance association updates in one DB transaction. Report success only when saved and applied.

The filesystem and DB are not one atomic transaction. If file write succeeds but activation fails,
keep last-good active behaviour and return an explicit saved_not_applied result with current disk/
active hashes. On startup or explicit Reload, reconcile journal/file hashes, validate/compile, and
finish activation or retain last-good with an error. No silent rollback of a user's externally edited
file. Pending operations pin their original compiled definition; an update affects new operations.
Keep current and previous good source/compiled revisions plus any revision referenced by pending
work. Prune unreferenced history after seven days. Restore previous writes another reviewed YAML
revision through the same path; it does not replay provider requests.

Reload reads host YAML and follows identical validation/activation rules without rewriting the file.
M1 needs startup load and an explicit Reload action, not a background file watcher. Invalid edits
leave the previous active definition running with a clear disk_error/active status. A missing file
blocks new runs after reload; retain existing resource data and offer restore. File removal is not
equivalent to confirmed deletion of instances or private data. UI Remove must identify affected
instances and coordinate disable/deletion first.

One backend process owns runtime state. Disk-hash conflict checks protect app writers and detect
ordinary host changes; they cannot make an arbitrary external editor participate in an atomic CAS.
Document that simultaneous external write/Git checkout and UI saving is unsupported: pause app edits
or stop the service for bulk changes, then Reload. No false claim of a filesystem-wide edit lock.

## Plain-language editor

Use the [operations/help contract](operations-and-help.md) for guided labels and value pickers.
“Provider field” names the API parameter; “Use this value” selects an input such as Postcode entered
in Setup or a provider-saved reference. Fixed constants reveal a separate text field. Generated
lookup/property values are explained as saved automatically, not displayed as read-only input
expressions. Ordinary setup does not require this screen. Keep advanced YAML syntax in its own view.

## Git and backups

The default files are ordinary YAML: users can diff/commit them using their existing Git workflow,
or download a definition and commit it elsewhere. No git add/commit/push, GitHub permission or network
account is required by the application. Local revision recovery is not Git history. Review diff for
accidental literal personal data before committing; the UI schema permits secret references only in
auth positions and rejects plaintext tokens/passwords in those positions.

Back up YAML definitions with the consistent runtime DB snapshot and separately protected keyring.
On restore compare referenced hashes, preserve last-good compiled revisions, and show any definition
mismatch for Reload. Runtime caches/credentials/property bindings are not reconstructed from Git.
Do not add broad ignore rules hiding the shareable definitions. Runtime DB/history/key files stay
outside the committable definition tree. Update deployment/ignore examples in the implementation
phase; M1 deployment instructions are in [docs/connections.md](../../../docs/connections.md).

## YAML grammar and execution limits

YAML is serialization, not executable code. One document, string keys, known fields and typed scalars;
reject duplicate keys, custom tags, anchors/aliases/merge keys and unsupported scalar coercions in v1.
Use true/false booleans, quote date-like strings, and preserve date-only semantics. Bound bytes,
depth and node count before decoding. Reject unknown fields rather than silently ignoring typos.
No !include, !secret custom tag, Jinja, shell substitution or runtime package loading. Use an explicit
secret_slot/ref field and typed value references.

The [illustrative WhenIsBins definition](whenisbins.definition.example.yaml) shows the intended
shape and contains no live household values. It is planning data, not a live-verified default.
Generic operations handle HTTP/idempotency/cache mechanics, explicit YAML links select requests,
and the common waste validator handles domain meaning. The sample demonstrates the property_id
branch. The [typed format](definition-format.md) defines the recognized fields; cross-reference,
quality and provider-edge fixtures remain required before implementation.

The existing backend dependency graph already includes gopkg.in/yaml.v3 v3.0.1. Verify its strict/node
behaviour and Go-toolchain compatibility before making it direct; consider a stable maintained v3
successor separately. Do not adopt an unreleased v4 or perform a repository-wide parser migration
for this feature. KnownFields alone does not enforce every node/resource rule above.
[yaml.v3 API](https://pkg.go.dev/gopkg.in/yaml.v3),
[maintainer version guidance](https://raw.githubusercontent.com/yaml/go-yaml/main/README.md)

The [operations/help](operations-and-help.md) and [execution contract](execution-and-security.md)
fix supported operations, value references,
header/status/body parsing, finite mappings and credential isolation. A new YAML definition can
reuse a supported recipe/capability. Endpoint/field/schema repairs within those primitives need
no build/release. An entirely new auth protocol/transport/transform/capability still needs reviewed
code; YAML alone cannot implement arbitrary APIs. Broaden reusable adapters when real later
integrations show what is shared; avoid a universal workflow language up front.

## Verification

Source ownership: configuration.go/configuration_test.go for YAML grammar/compile,
repository/integration_definitions.go/integration_definitions_test.go for host I/O/journal/hash
recovery, existing integrations repository companion for atomic active revisions, and frontend
configuration service/domain companions for form/YAML parity. Browser tests cover both viewports.

Required evidence: two signed-in users can edit/continue shared state; anonymous denied; UI/UI and
ordinary host-edit conflicts; invalid YAML/unknown fields/aliases; read-only volume/no temp fallback;
disk-full/rename/activation failure and crash recovery; definition removal/restore; pending-revision
pinning; no automatic Git operation; no secret/private-state export; mobile Form repair and YAML
buffer preservation. Actual M1 checks are recorded in the implementation review.


Runtime DB/key path selection must also reject temporary fallback, not only definition writes.
The [runtime contract](runtime-lifecycle.md) defines strict storage, schema migration, feature-degraded
startup and joined cleanup. [API/recovery](api-and-recovery.md) defines strong preconditions,
origin-review tickets, credential/control paths and the missing refresh renewal branch.

## Default packaging and build artifact

Current backend/Dockerfile copies only the compiled binary into the runtime image. A definition
left only in specs/ or an unmounted source directory will not ship with a binary-only install.
Proposed single reviewed seed lives at backend/configs/integrations/waste_collection/whenisbins.yaml;
a small backend/configs/integrations/embed.go embeds that YAML for the optional feature registry.
The same file remains an ordinary committable host definition when the repo's configs volume is
mounted. Do not maintain a second hand-edited provider copy. The planning example is reference data,
not the runtime asset. embed_test.go owns the default-asset/format parity checks.

A new empty writable config directory receives missing embedded seeds by exclusive atomic create.
Existing host files are never overwritten; upgrades offer their existing diff/update flow. Newly
added valid host providers are scanned without recompiling. New operation/capability types still
require code. Only the initial waste type ships; embedding a seed is not a provider-named executor.
A selected reviewed seed approves its stated data destination through ordinary setup; new or changed
origins still follow the explicit credential-review rules. No host/YAML file authorizes its own secrets.


First-release deployment is confirmed personal Docker use. See the explicit mount/key/backup
paths in [execution/security](execution-and-security.md#personal-docker-deployment-first-release).
Existing config/data mounts remain; add only the separate read-only external secrets-directory mount
during implementation. No Compose file, key or running container is changed by these planning edits.
