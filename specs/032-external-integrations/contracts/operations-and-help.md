# Reusable operations, schemas and in-app help

Status: Planning contract; no application page or runtime implemented.
Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md)

## Generic operation types and named requests

The proposed runtime implements HTTP, field extraction, input/selection, finite conditions, bounded polling
and validated save operations. YAML connects them. A request name is an arbitrary identifier,
not a Go function name. No provider-named adapter is needed for the initial HTTP flow; the shared
waste capability validator handles waste meaning separately from transport/setup orchestration.

M1 uses one reusable `search_select` recipe: show search input, run its configured request, display
configured returned choices, run the selection's configured request, optionally check configured
pending status, then save validated data. Refresh independently names its request. This recipe
composes generic operations and is not a general graph/script editor. Only bins ship initially.
Additional recipes/types are implemented only when a real integration requires them.

The connections are explicit, for example:

```yaml
setup:
  type: search_select
  search:
    input: postcode
    request: discover
    options: {items: /candidates, value: /id, label: /label}
  select: {input: property_id, request: submit}
  status:
    request: poll
    pointer: /status
    pending: [queued, running, partial]
    done: [done]
    failed: [failed]
refresh: {request: schedule}
```

Here discover/submit/poll/schedule could all be renamed if their YAML references are updated.
Only operation/recipe type names are defined in code. Unknown request/input names are configuration
errors; the engine never guesses from English names. Missing refresh state disables Refresh with
a clear incomplete-setup explanation instead of guessing a property/token.

Inputs are declared by ID/type/label/required. HTTP query/body/path values reference `input.<id>`
or `state.<id>`. Captures populate named state values from typed response selectors. Reserved
transport state includes ETag; credential references are isolated from ordinary value references.
For lookup/status replies, captures and the status selector read the full decoded body; response_root
selects the subtree consumed by the common output mapper. Choice value/label pointers are relative
to each item in the options array. HTTP errors/304/Retry-After are handled before successful mapping.

M1 captures support string/boolean/number/local-date/UTC-timestamp types, optional absence and
validated literal prefix removal. No arbitrary expressions or recursive templates. Pending status
sets must be disjoint from done/failed; terminal states stop continuation. M1 Check status is manual;
the same operation becomes scheduled in M2. Every POST attempt marked idempotent preserves its
body/key; no implicit POST retry is enabled by merely naming a request `submit`.

## Keep the schema small

There is no separate large schema framework in M1. The [small definition format](definition-format.md)
is represented by ordinary Go structs and a few checks, with matching frontend contract types.
Users edit endpoints, request links and field mappings; they do not write JSON Schema or reproduce
all provider response fields. The format-version marker identifies supported configuration versions.

| Piece | How it stays simple |
| --- | --- |
| YAML configuration | Backend checks required fields, supported methods/types and valid references |
| Provider response | Map only fields we consume; validate their expected strings/lists/dates |
| Waste result | One shared typed result for all providers and colours; examples are developer reference |

A moved field is a mapping edit; a changed URL is a request edit. Unrelated upstream additions are
ignored. Invalid/missing required data retains last-good results and shows the affected field.
A genuinely new operation/auth/transport or output meaning can still need code. Strict YAML key
validation does not mean rejecting unrelated provider-response keys.

The backend validator is authoritative. Form validation adds immediate feedback and displays server
errors with field/YAML positions. Contract/example tests keep Go, TypeScript, documentation and the
form aligned. Do not add a JSON Schema interpreter, auto-generated editor or a second schema language
for ordinary configuration. Generated schemas can be reconsidered later if actual tooling needs them.

## Help page in the operations/configuration screen

Proposed navigation: Connections → Waste collection → provider → Configure operations → Help.
Use an inline Help tab plus `/connections/:type/providers/:provider/help`; the
[navigation contract](connections-navigation.md) defines instance and provider routes. All signed-in
users can read it. Desktop can retain the editor alongside contextual help; mobile opens a full-width
Help tab/route with Back to configuration. Keep draft state in memory while navigating between them,
preserve focus/scroll and do not discard or persist household inputs in browser storage.

Use task-based, plain-English topics: **Set up a connection**, **Choose what to send**,
**Understand the steps**, **Find names and dates**, **Fix a provider change**, **Save your settings**.
Start with what the user wants to achieve, then give numbered steps and one concrete example.
Do not introduce internal input/state namespaces, JSON Pointer terminology or schema definitions
before explaining the job. Put YAML syntax in an optional “For people editing YAML directly” section.

## Guided values and readable explanations

Normal setup still asks only for postcode/address. Configure operations is optional repair work.
Replace code-like read-only controls with the following presentation:

| Label | Behavior and example |
| --- | --- |
| Provider address | Editable base URL; help says “the first part, such as https://whenisbins.com” |
| Request path | Editable remaining path; show how it joins the provider address |
| Reference name (used in YAML) | Editable short ID; explain that the supplied value can stay unchanged |
| Information to send | Group containing actual parameter controls, not an expression string |
| Provider field | Editable provider key, e.g. postcode; its meaning is explained alongside it |
| Use this value | Picker of readable available inputs, e.g. Postcode entered in Setup |
| Value to send every time | Revealed only after choosing A fixed value; for provider constants |
| Saved automatically | Plain explanatory text for a provider-generated lookup/property reference |
| Read the collection results | The list/field locations used to display bin names, dates and colours |
| Preview results | Show a mapped sample; explicitly distinguish sample validation from a live URL test |

For the current input, the picker displays **Postcode entered in Setup** and the form stores the
existing `{ref: input.postcode}` value. **Address selected in Setup** maps to input.property_id.
Later stages may select **Saved lookup ID** / **Saved property reference** when those values are
available. A fixed-value choice stores a literal; input choices store references, never a copied
personal postcode. Pickers list values available at that step, with unavailable ones explained.
Secret references belong to dedicated credential controls, never the ordinary value picker.

The Form view shows a live sentence: “Send the postcode entered in Setup as ‘postcode’.” Changing
the provider key to postal_code changes only that key; the value still comes from Setup. Show a
clear error for duplicate/invalid keys, missing fixed values or unavailable sources. Form and YAML
must preserve key/source/fixed-value edits in both directions and across Help navigation. Native
labels, 44px targets and stacked mobile rows apply. Automatic explanations are paragraphs/cards,
not disabled or read-only text inputs that imply editable configuration.

The production Form must edit supported path/query/body bindings without requiring handwritten
reference expressions. Automatically saved values are explained separately from the controls that
choose where they are used. Complete YAML validation is part of the backend implementation.

The help must answer these practical questions:

1. **Set up/change an address:** choose the provider, enter postcode, select the address and save;
   all bin colours appear together. One installation is one shared home.
2. **Choose what to send:** provider field means the name the service expects; Use this value means
   where it comes from. Worked example: postcode + Postcode entered in Setup. Explain automatic
   IDs as numbers/references the provider supplies and the application remembers.
3. **Understand the steps:** Find addresses → Get collection dates → Check progress if needed →
   Refresh dates. Supplied settings connect these; users do not assemble a workflow. Explain manual
   M1 Check status versus later automatic continuation without implying either already exists.
4. **Find names and dates:** /collections means the top-level list named collections; /data/collections
   means open data then collections. /dates is the date list inside each bin. One map covers all
   colours. Show a small example and a Preview results action before introducing technical names.
5. **Fix a provider change:** edit the affected address/path, rename an input field or move a result
   location. Preview sample results, then validate/save. State that a sample does not verify live
   connectivity. Explain any credential reconnect and preserve last-good operation on failure.
6. **Save your settings:** Form/YAML edit the same host file; Save writes/applies, Reload reads host
   edits, Restore returns to an earlier version. Explain conflicts/errors with a next action, and
   that manual Git versioning excludes credentials, selected addresses and fetched state.

Every reference entry states purpose, supported type/values, required/default behaviour, a synthetic
example and common error. Contextual help links from a field/error target the relevant anchor. Link
the supported format reference and offer a copyable synthetic example/definition download. Bundled help
works without external documentation/network calls and does not require AI assistance.

Keep help content static/trusted, using existing Vue/shared controls rather than rendering provider
HTML or introducing a documentation platform. A proposed IntegrationsHelp.vue owns presentation;
source-owned unit tests verify pure help helpers, while Playwright verifies rendered
navigation, keyboard/focus, mobile widths, error links, search and returning with an unchanged draft.

## Acceptance evidence

Validate the shared example against the typed format and reference rules;
prove renaming request IDs and all corresponding references changes no code; reject a stale/missing
reference. Use fixtures to map zero/one/multiple colours with one collection mapping and show a
field-movement repair. Help must use the exact shipped names/grammar and must not claim unsupported
recipes or operations are available. No application check is claimed passed by this document.

## Draft and navigation ownership

Keep a memory-only provider draft keyed by type/provider and its base disk hash in the feature's
Vuex state. Help navigation must not reload a newer server document over a dirty draft. A changed
server revision offers Compare/Reload; it never silently discards user edits. Invalid raw YAML
preserves its buffer and blocks Form/save until corrected. Reference-name dictionaries use Maps or
own-property checks, so names such as constructor cannot become inherited object values.

AppShell's existing fullPath watcher resets its main scroll on navigation. Do not change that
behavior globally: restore editor field/focus/scroll inside the Connections feature after its route
mount/watch completes, or keep the editor in a stable feature parent while Help opens. Use existing
shared controls/themes and source-owned pure helpers plus browser assertions. Unmount cancels local
status polls; it does not cancel an already admitted provider operation. Logout/session invalidation
clears all feature drafts/setup/private buffers from memory and never writes them to browser storage.

M1 offers exact address-choice setup only. Help must explain unsupported council journeys without
claiming neighbouring-property consent, direct address typing or another recipe is implemented.
