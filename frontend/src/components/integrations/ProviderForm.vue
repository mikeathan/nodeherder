<script setup lang="ts">
  import { computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { RouteName } from '@/types/router';
  const route = useRoute();
  import UiButton from '@/components/ui/UiButton.vue';
  import RequestValues from './RequestValues.vue';
  import { humanLabel, valueChoices } from '@/domain/integrations';
  import type { ConnectionValue, ProviderDefinition } from '@/types/integrations.type';
  // The parent owns this editable document; updates are intentionally passed through v-model.
  const props = defineProps<{ modelValue: ProviderDefinition }>();
  const emit = defineEmits<{ 'update:modelValue': [value: ProviderDefinition] }>();
  const d = computed({ get: () => props.modelValue, set: (value) => emit('update:modelValue', value) });
  const choices = computed(() => valueChoices(d.value));
  const requestIDs = computed(() => Object.keys(d.value.requests));
  function paths(id: string): Record<string, ConnectionValue> {
    return Object.fromEntries(d.value.requests[id].path.map((v, i) => [`segment_${i + 1}`, v]));
  }
  function setPaths(id: string, values: Record<string, ConnectionValue>) {
    d.value.requests[id].path = Object.values(values);
  }
  function addRequest() {
    let i = 1;
    while (`request_${i}` in d.value.requests) i++;
    d.value.requests[`request_${i}`] = { method: 'GET', path: ['v1'], query: {}, captures: {} };
  }
  function renameRequest(old: string, name: string) {
    if (!/^[a-z][a-z0-9_]{0,63}$/.test(name) || name in d.value.requests) return;
    d.value.requests = Object.fromEntries(
      Object.entries(d.value.requests).map(([key, v]) => [key === old ? name : key, v])
    );
    for (const step of [d.value.setup.search, d.value.setup.select, d.value.setup.status, d.value.refresh])
      if (step.request === old) step.request = name;
    if (d.value.refresh.renewal_request === old) d.value.refresh.renewal_request = name;
  }
  function addCapture(id: string) {
    const q = d.value.requests[id];
    q.captures ??= {};
    let i = 1;
    while (`saved_value_${i}` in q.captures) i++;
    q.captures[`saved_value_${i}`] = { pointer: '/id', type: 'string' };
  }
  function renameCapture(request: string, old: string, name: string) {
    const captures = d.value.requests[request].captures!;
    if (name && !(name in captures)) {
      captures[name] = captures[old];
      delete captures[old];
    }
  }
</script>
<template>
  <div class="connection-form">
    <p class="connection-muted">
      These shared settings describe how the provider works. Your address and token are saved separately in Setup. Field
      locations use a slash, for example /collections; you only need to change them when the provider changes its
      response.
    </p>
    <fieldset class="connection-form">
      <legend>Provider</legend>
      <RouterLink :to="{ name: RouteName.ConnectionHelp, params: route.params, query: route.query, hash: '#repair' }"
        >Help changing a URL</RouterLink
      ><label class="connection-field"
        ><span>Display name</span><input id="name" v-model="d.name" maxlength="120" /></label
      ><label class="connection-field"
        ><span>Provider URL</span><input id="base_url" v-model="d.base_url" type="url"
      /></label>
      <div class="connection-row">
        <label class="connection-field"
          ><span>Authentication</span
          ><select v-model="d.auth.mode">
            <option value="none">No token</option>
            <option value="optional_bearer">Optional API token</option>
            <option value="bearer">Required API token</option>
          </select></label
        ><label v-if="d.auth.mode !== 'none'" class="connection-field"
          ><span>Credential slot name</span><input v-model="d.auth.secret_slot"
        /></label>
      </div>
      <p class="connection-muted">
        The provider ID and supported output type stay the same when repairing this file. Use Copy provider to create a
        separate definition.
      </p>
    </fieldset>
    <details>
      <summary>Setup inputs</summary>
      <div class="connection-form">
        <div v-for="(input, key) in d.inputs" :key="key" class="connection-row">
          <label class="connection-field"
            ><span>{{ humanLabel(String(key)) }} label</span><input v-model="input.label" /></label
          ><label class="connection-field"
            ><span>Input type</span
            ><select v-model="input.type">
              <option value="postcode">Postcode</option>
              <option value="select">Address choice</option>
              <option value="text">Text</option>
            </select></label
          ><label><input v-model="input.required" type="checkbox" /> Required</label>
        </div>
      </div>
    </details>
    <h2>Requests</h2>
    <RouterLink :to="{ name: RouteName.ConnectionHelp, params: route.params, query: route.query, hash: '#values' }"
      >Help choosing request values</RouterLink
    >
    <p class="connection-muted">
      Each request has a name used by the setup and refresh steps below. Renaming it here updates those links.
    </p>
    <details v-for="(request, id) in d.requests" :id="`requests.${id}`" :key="id" tabindex="-1">
      <summary>{{ humanLabel(String(id)) }} · {{ request.method }}</summary>
      <div class="connection-form">
        <div class="connection-row">
          <label class="connection-field"
            ><span>Request name</span
            ><input :value="id" @change="renameRequest(String(id), ($event.target as HTMLInputElement).value)" /></label
          ><label class="connection-field"
            ><span>Method</span
            ><select v-model="request.method">
              <option>GET</option>
              <option>POST</option>
            </select></label
          >
        </div>
        <RequestValues
          :id="`requests.${id}.path`"
          :model-value="paths(String(id))"
          :choices="choices"
          title="URL path — one segment at a time"
          path-mode
          @update:model-value="setPaths(String(id), $event)" />
        <RequestValues
          :model-value="request.query ?? {}"
          :choices="choices"
          title="URL parameters"
          @update:model-value="request.query = $event" />
        <RequestValues
          v-if="request.method === 'POST'"
          :model-value="request.body ?? {}"
          :choices="choices"
          title="Submitted fields"
          @update:model-value="request.body = $event" />
        <label v-if="request.method === 'POST'"
          ><input v-model="request.idempotency" type="checkbox" /> Reuse a submission key when safely retrying this
          request</label
        ><label v-else
          ><input v-model="request.conditional" type="checkbox" /> Ask whether the saved response has changed</label
        >
        <label class="connection-field"
          ><span>Result location</span
          ><input v-model="request.response_root" placeholder="Leave empty when the result is the whole response"
        /></label>
        <fieldset class="connection-form">
          <legend>Values saved from the response</legend>
          <p class="connection-muted">
            Lookup and property references are saved automatically for later requests. No address or token belongs in
            this configuration.
          </p>
          <div v-for="(capture, key) in request.captures" :key="key" class="connection-form">
            <div class="connection-row">
              <label class="connection-field"
                ><span>Saved value name</span
                ><input
                  :value="key"
                  @change="renameCapture(String(id), String(key), ($event.target as HTMLInputElement).value)" /></label
              ><label class="connection-field"
                ><span>Provider field location</span><input v-model="capture.pointer"
              /></label>
            </div>
            <div class="connection-row">
              <label class="connection-field"
                ><span>Value type</span
                ><select v-model="capture.type">
                  <option value="string">Text</option>
                  <option value="boolean">True or false</option>
                  <option value="number">Number</option>
                  <option value="local-date">Calendar date</option>
                  <option value="utc-timestamp">Timestamp</option>
                </select></label
              ><label class="connection-field"
                ><span>Save when</span
                ><select
                  :value="capture.when ?? ''"
                  @change="capture.when = ($event.target as HTMLSelectElement).value">
                  <option value="">The response arrives</option>
                  <option value="complete">The lookup finishes successfully</option>
                </select></label
              >
            </div>
            <label class="connection-field"
              ><span>Remove this required prefix from the provider identifier</span
              ><input v-model="capture.strip_prefix" placeholder="For example p:" /></label
            ><label><input v-model="capture.optional" type="checkbox" /> This field may be absent</label>
            <div>
              <UiButton variant="ghost" @click="delete request.captures![String(key)]">Remove saved value</UiButton>
            </div>
          </div>
          <div><UiButton @click="addCapture(String(id))">Add saved value</UiButton></div>
        </fieldset>
        <div><UiButton variant="ghost" @click="delete d.requests[String(id)]">Remove request</UiButton></div>
      </div>
    </details>
    <div><UiButton @click="addRequest">Add request</UiButton></div>
    <details>
      <summary>Connect the setup and refresh steps</summary>
      <div class="connection-form">
        <label
          v-for="(step, key) in {
            search: d.setup.search,
            select: d.setup.select,
            status: d.setup.status,
            refresh: d.refresh,
          }"
          :key="key"
          class="connection-field"
          ><span>{{
            key === 'search'
              ? 'Find addresses'
              : key === 'select'
                ? 'Retrieve selected address'
                : key === 'status'
                  ? 'Check lookup status'
                  : 'Read saved schedule'
          }}</span
          ><select v-model="step.request">
            <option v-for="id in requestIDs" :key="id" :value="id">{{ humanLabel(id) }}</option>
          </select></label
        ><label class="connection-field"
          ><span>Request a fresh lookup when the saved schedule expires</span
          ><select v-model="d.refresh.renewal_request">
            <option value="">No renewal request</option>
            <option v-for="id in requestIDs" :key="id" :value="id">{{ humanLabel(id) }}</option>
          </select></label
        >
        <div class="connection-row">
          <label class="connection-field"
            ><span>Search input</span
            ><select v-model="d.setup.search.input">
              <option v-for="(input, key) in d.inputs" :key="key" :value="key">{{ input.label }}</option>
            </select></label
          ><label class="connection-field"
            ><span>Address choice input</span
            ><select v-model="d.setup.select.input">
              <option v-for="(input, key) in d.inputs" :key="key" :value="key">{{ input.label }}</option>
            </select></label
          >
        </div>
        <label v-for="(value, key) in d.setup.search.options" :key="key" class="connection-field"
          ><span>Address list: {{ humanLabel(String(key)) }}</span
          ><input v-model="d.setup.search.options[key]"
        /></label>
        <div v-if="d.setup.search.expect" class="connection-row">
          <label class="connection-field"
            ><span>Required input field location</span><input v-model="d.setup.search.expect.pointer" /></label
          ><label class="connection-field"
            ><span>Supported input value</span><input v-model="d.setup.search.expect.value"
          /></label>
        </div>
        <label class="connection-field"
          ><span>Lookup status field location</span><input v-model="d.setup.status.pointer" /></label
        ><label v-for="key in ['pending', 'done', 'failed'] as const" :key="key" class="connection-field"
          ><span>{{ humanLabel(key) }} status values — separated by commas</span
          ><input
            :value="d.setup.status[key].join(', ')"
            @change="
              d.setup.status[key] = ($event.target as HTMLInputElement).value
                .split(',')
                .map((s) => s.trim())
                .filter(Boolean)
            "
        /></label>
      </div>
    </details>
    <details>
      <summary>Collection fields</summary>
      <RouterLink :to="{ name: RouteName.ConnectionHelp, params: route.params, query: route.query, hash: '#fields' }"
        >Help mapping collection fields</RouterLink
      >
      <div class="connection-form">
        <p class="connection-muted">
          Map the collection list once. The same fields are used for every returned bin; no separate configuration is
          needed for each colour.
        </p>
        <label class="connection-field"
          ><span>Collection list location</span
          ><input id="mapping.collections.items" v-model="d.mapping.collections.items"
        /></label>
        <div class="connection-row">
          <label v-for="(value, key) in d.mapping.collections.fields" :key="key" class="connection-field"
            ><span>{{ humanLabel(String(key)) }}</span
            ><input :id="`mapping.collections.fields.${key}`" v-model="d.mapping.collections.fields[key]"
          /></label>
        </div>
        <h3>Schedule information</h3>
        <div class="connection-row">
          <label v-for="(value, key) in d.mapping.metadata" :key="key" class="connection-field"
            ><span>{{ humanLabel(String(key)) }}</span
            ><input :id="`mapping.metadata.${key}`" v-model="d.mapping.metadata[key]"
          /></label>
        </div>
      </div>
    </details>
    <details>
      <summary>Request limits and errors</summary>
      <div class="connection-form">
        <label class="connection-field"
          ><span>Minimum seconds between status checks</span
          ><input v-model.number="d.policy.poll_min_seconds" type="number" min="5" /></label
        ><label class="connection-field"
          ><span>Minimum seconds between new lookups</span
          ><input v-model.number="d.policy.renewal_min_seconds" type="number" min="86400" /></label
        ><label class="connection-field"
          ><span>Provider's documented submission-key lifetime in seconds</span
          ><input v-model.number="d.policy.idempotency_window_seconds" type="number" min="0" /></label
        ><label class="connection-field"
          ><span>Provider error code location</span><input v-model="d.errors.code"
        /></label>
        <div v-for="(rule, i) in d.errors.rules" :key="i" class="connection-form">
          <div class="connection-row">
            <label class="connection-field"
              ><span>HTTP error status</span
              ><input v-model.number="rule.status" type="number" min="400" max="599" /></label
            ><label class="connection-field"
              ><span>Exact provider problem code</span><input v-model="rule.code"
            /></label>
          </div>
          <label class="connection-field"
            ><span>Recovery action</span
            ><select v-model="rule.action">
              <option value="renewal">Request a fresh lookup</option>
              <option value="restart_setup">Ask to repeat setup</option>
              <option value="reconnect">Ask to replace the credential</option>
              <option value="unsupported">Explain that this setup is unsupported</option>
            </select></label
          >
          <div><UiButton variant="ghost" @click="d.errors.rules.splice(i, 1)">Remove error rule</UiButton></div>
        </div>
        <div>
          <UiButton @click="d.errors.rules.push({ status: 404, code: '', action: 'restart_setup' })"
            >Add error rule</UiButton
          >
        </div>
      </div>
    </details>
  </div>
</template>
