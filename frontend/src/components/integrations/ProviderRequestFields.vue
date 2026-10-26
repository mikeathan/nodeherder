<script setup lang="ts">
  import UiButton from '@/components/ui/UiButton.vue';
  import { useEditableModel } from '@/composables/integrations/useEditableModel';
  import type { ProviderRequest } from '@/types/integrations.type';
  import RequestValues from './RequestValues.vue';
  import { requestPathFields, nextDefinitionKey } from '@/domain/provider-definition';
  const props = defineProps<{ modelValue: ProviderRequest; id: string; choices: { value: string; label: string }[] }>();
  const emit = defineEmits<{
    'update:modelValue': [value: ProviderRequest];
    'rename': [name: string];
    'renameCapture': [oldName: string, name: string];
    'remove': [];
  }>();
  const request = useEditableModel(
    () => props.modelValue,
    (value) => emit('update:modelValue', value)
  );
  function addCapture() {
    const captures = request.value.captures ?? {};
    request.value.captures = {
      ...captures,
      [nextDefinitionKey(captures, 'saved_value')]: { pointer: '/id', type: 'string' },
    };
  }
</script>
<template>
  <div class="connection-form">
    <div class="connection-form">
      <div class="connection-row">
        <label class="connection-field"
          ><span>Request name</span
          ><input :value="id" @change="emit('rename', ($event.target as HTMLInputElement).value)" /></label
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
        :model-value="requestPathFields(request)"
        :choices="choices"
        title="URL path — one segment at a time"
        path-mode
        @update:model-value="request.path = Object.values($event)" />
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
          Lookup and property references are saved automatically for later requests. No address or token belongs in this
          configuration.
        </p>
        <div v-for="(capture, key) in request.captures" :key="key" class="connection-form">
          <div class="connection-row">
            <label class="connection-field"
              ><span>Saved value name</span
              ><input
                :value="key"
                @change="emit('renameCapture', String(key), ($event.target as HTMLInputElement).value)" /></label
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
              ><select :value="capture.when ?? ''" @change="capture.when = ($event.target as HTMLSelectElement).value">
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
        <div><UiButton @click="addCapture()">Add saved value</UiButton></div>
      </fieldset>
      <div><UiButton variant="ghost" @click="emit('remove')">Remove request</UiButton></div>
    </div>
  </div>
</template>
