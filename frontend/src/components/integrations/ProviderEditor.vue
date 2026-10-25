<script setup lang="ts">
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ProviderForm from './ProviderForm.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import ProviderDestinationReview from './ProviderDestinationReview.vue';
  import ProviderTestPanel from './ProviderTestPanel.vue';
  import ProviderComparison from './ProviderComparison.vue';
  import { useProviderEditor } from '@/composables/integrations/useProviderEditor';
  import { RouteName } from '@/types/router';
  import './connections.css';
  const {
    key,
    newFile,
    provider,
    yaml,
    documentValue,
    mode,
    dirty,
    validation,
    approve,
    reassociate,
    message,
    busy,
    error,
    helpRoute,
    affectedInstances,
    instances,
    rememberFocus,
    useShipped,
    keepVersion,
    validate,
    switchMode,
    save,
    reload,
    restoreMissing,
    restore,
    copy,
    remove,
    testInstance,
    sample,
    preview,
    test,
  } = useProviderEditor();
</script>

<template>
  <div class="connections connection-stack" @focusin="rememberFocus">
    <UiPageHeader
      title="Configure operations"
      :subtitle="provider?.document.name"
      :back="{ name: RouteName.Connections }"
      back-label="Connections"
      ><template #actions><UiButton :to="helpRoute" icon="unknown">Help</UiButton></template></UiPageHeader
    >
    <ConnectionFeedback :error="error" :busy="busy" />
    <div v-if="message" class="connection-notice" role="status">{{ message }}</div>
    <div v-if="provider?.status === 'host_changed'" class="connection-notice" role="status">
      The host file differs from the active configuration. Review it and choose Reload from host to apply it.
    </div>
    <div v-if="provider?.disk_error" class="connection-notice" role="status">{{ provider.disk_error }}</div>
    <UiCard v-if="provider" title="Provider configuration"
      ><div class="connection-stack">
        <p class="connection-muted">
          Changes affect every connection using this provider. Saved on the host as configs/integrations/{{ key }}.yaml.
          Address choices, tokens and results stay separate.
        </p>
        <ProviderComparison
          v-if="provider.shipped_yaml && !newFile"
          :provider="provider"
          :busy="busy"
          @use-shipped="useShipped"
          @keep-version="keepVersion" />
        <div class="connection-actions">
          <UiButton :pressed="mode === 'form'" :disabled="busy" @click="switchMode('form')">Form</UiButton
          ><UiButton :pressed="mode === 'yaml'" :disabled="busy" @click="switchMode('yaml')">YAML</UiButton
          ><span class="connection-muted" role="status">{{ dirty ? 'Unsaved changes' : 'Saved' }}</span>
        </div>
        <fieldset :disabled="busy" class="connection-form-fields">
          <legend>Provider fields</legend>
          <ProviderForm v-if="mode === 'form' && documentValue" v-model="documentValue" />
          <label v-else class="connection-field"
            ><span>Provider YAML</span
            ><textarea
              id="provider-yaml"
              v-model="yaml"
              class="connection-code"
              spellcheck="false"
              aria-describedby="yaml-note"
              maxlength="65536" />
          </label>
        </fieldset>
        <p id="yaml-note" class="connection-muted">
          Form edits may normalize YAML formatting. Fixed values and comments are part of the shareable file; review
          them before committing it to Git. Save does not run Git commands.
        </p>
        <ProviderDestinationReview
          v-if="validation?.previous_origin && validation.origin !== validation.previous_origin"
          v-model:approved="approve"
          v-model:reassociate="reassociate"
          :validation="validation"
          :instances="affectedInstances"
          :busy="busy" />
        <ProviderTestPanel
          v-if="!newFile"
          v-model:instance="testInstance"
          v-model:sample="sample"
          :instances="instances"
          :preview="preview"
          :busy="busy"
          @test="test" />
        <div class="connection-actions">
          <UiButton :disabled="busy" @click="validate">Validate</UiButton
          ><UiButton variant="primary" :disabled="busy" @click="save">Save changes</UiButton
          ><UiButton v-if="!newFile" :disabled="busy" @click="reload">Reload from host</UiButton
          ><UiButton
            v-if="provider.status === 'missing' && provider.active_hash"
            :disabled="busy"
            @click="restoreMissing"
            >Restore last active file</UiButton
          >
          <UiButton v-if="!newFile && provider.status !== 'missing'" :disabled="busy" @click="restore"
            >Restore previous</UiButton
          >
        </div>
        <div v-if="!newFile" class="connection-actions">
          <UiButton :disabled="busy || provider.status !== 'active'" @click="copy">Copy provider</UiButton
          ><UiButton variant="ghost" :disabled="busy" @click="remove">Remove definition</UiButton>
        </div>
      </div></UiCard
    >
  </div>
</template>
