<script setup lang="ts">
  import { computed } from 'vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import {
    activeAlertCountLabel,
    alertLevelLabel,
    alertSeverityLabel,
    alertSeverityTone,
  } from '@/domain/integrations/weather-alerts';
  import type { WeatherAlerts } from '@/types/integrations/weather-alerts';
  import { formatLocalDateTime } from '@/utils/date.utils';

  const props = defineProps<{ alerts: WeatherAlerts }>();
  const warnings = computed(() => props.alerts.warnings);
  const status = computed(() => activeAlertCountLabel(warnings.value.length));
  const stale = computed(() => props.alerts.freshness === 'stale');
</script>

<template>
  <div class="alert-results">
    <section class="alert-panel" aria-label="Weather-alert summary">
      <p class="alert-muted">{{ alerts.attribution }}</p>
      <div class="alert-summary">
        <div>
          <span class="alert-status">{{ status }}</span>
          <small>{{ alerts.status === 'active' ? 'Warnings in force' : 'Confirmed clear' }}</small>
        </div>
        <UiIcon name="warn" class="alert-badge" />
      </div>
      <p class="alert-muted">Captured {{ formatLocalDateTime(alerts.captured_at) }}</p>
      <p class="alert-muted">Last checked {{ formatLocalDateTime(alerts.last_checked_at) }}</p>
      <p v-if="stale" role="status">This alert data is out of date.</p>
    </section>

    <p v-if="!warnings.length" class="alert-empty">No active warnings were returned.</p>

    <article
      v-for="warning in warnings"
      :key="warning.warning_id"
      class="alert-warning"
      :data-tone="alertSeverityTone(warning.severity)">
      <header>
        <span class="alert-severity" :data-tone="alertSeverityTone(warning.severity)">
          {{ alertSeverityLabel(warning.severity) }}
        </span>
        <span class="alert-level">{{ alertLevelLabel(warning.level) }}</span>
        <span v-if="warning.type" class="alert-type">{{ warning.type }}</span>
      </header>
      <h2>{{ warning.headline }}</h2>
      <dl class="alert-window">
        <div>
          <dt>From</dt>
          <dd>{{ formatLocalDateTime(warning.valid_from) }}</dd>
        </div>
        <div>
          <dt>Until</dt>
          <dd>{{ formatLocalDateTime(warning.valid_to) }}</dd>
        </div>
      </dl>
      <ul v-if="warning.areas.length" class="alert-areas" aria-label="Affected areas">
        <li v-for="area in warning.areas" :key="area">{{ area }}</li>
      </ul>
      <p v-if="warning.description" class="alert-description">{{ warning.description }}</p>
      <section v-if="warning.instruction" class="alert-instruction">
        <h3>What to do</h3>
        <p>{{ warning.instruction }}</p>
      </section>
    </article>
  </div>
</template>

<style scoped>
  .alert-results {
    display: grid;
    gap: 1rem;
  }
  .alert-panel,
  .alert-warning {
    min-width: 0;
    padding: 1.25rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
  }
  .alert-warning {
    border-left: 3px solid var(--nh-border);
  }
  .alert-warning[data-tone='minor'] {
    border-left-color: var(--nh-info);
  }
  .alert-warning[data-tone='moderate'] {
    border-left-color: var(--nh-warn);
  }
  .alert-warning[data-tone='severe'] {
    border-left-color: var(--nh-danger);
  }
  .alert-summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    margin: 0.75rem 0 1rem;
  }
  .alert-summary > div {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .alert-status {
    font-size: 1.5rem;
    font-weight: 700;
  }
  .alert-summary small {
    color: var(--nh-text-muted);
  }
  .alert-badge {
    width: 2.5rem;
    height: 2.5rem;
    color: var(--nh-warn);
  }
  .alert-muted {
    margin: 0.5rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .alert-empty {
    padding: 1.5rem;
    border: var(--nh-border-w) dashed var(--nh-border);
    border-radius: var(--nh-radius-l);
    color: var(--nh-text-muted);
    text-align: center;
  }
  .alert-warning header {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.6rem;
  }
  .alert-severity {
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--nh-text-muted);
  }
  .alert-severity[data-tone='minor'] {
    color: var(--nh-info);
  }
  .alert-severity[data-tone='moderate'] {
    color: var(--nh-warn);
  }
  .alert-severity[data-tone='severe'] {
    color: var(--nh-danger);
  }
  .alert-level,
  .alert-type {
    padding: 0.15rem 0.45rem;
    border-radius: 999px;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    font-size: 0.7rem;
  }
  .alert-warning h2 {
    margin: 0.75rem 0 0;
    font-size: 1.15rem;
    line-height: 1.35;
  }
  .alert-window {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem 1.5rem;
    margin: 1rem 0 0;
  }
  .alert-window div {
    display: flex;
    gap: 0.35rem;
  }
  .alert-window dt {
    color: var(--nh-text-muted);
  }
  .alert-window dd {
    margin: 0;
    font-weight: 600;
  }
  .alert-areas {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    list-style: none;
    margin: 0.85rem 0 0;
    padding: 0;
  }
  .alert-areas li {
    padding: 0.2rem 0.5rem;
    border-radius: 999px;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
  .alert-description {
    margin: 0.9rem 0 0;
    line-height: 1.55;
  }
  .alert-instruction {
    margin-top: 1rem;
    padding: 0.85rem 1rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
  }
  .alert-instruction h3 {
    margin: 0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--nh-text-muted);
  }
  .alert-instruction p {
    margin: 0.45rem 0 0;
    line-height: 1.5;
  }
  @media (max-width: 700px) {
    .alert-panel,
    .alert-warning {
      padding: 1rem;
    }
  }
</style>
