<template>
  <div class="shell-loading" role="status">
    <span class="sr-only">Loading NodeHerder…</span>
    <aside aria-hidden="true"><span class="nh-skel" /><span class="nh-skel" /><span class="nh-skel" /></aside>
    <div class="loading-main" aria-hidden="true">
      <div class="loading-header"><span class="nh-skel" /></div>
      <div class="loading-content"><span class="nh-skel" /><span class="nh-skel loading-card" /></div>
    </div>
  </div>
</template>
<style scoped>
  .shell-loading { display: grid; grid-template-columns: var(--nh-nav-w) minmax(0, 1fr); height: 100dvh;
    min-height: 100dvh; background: var(--nh-bg); color: var(--nh-text); overflow: hidden; }
  aside { display: flex; flex-direction: column; gap: 1.5rem; padding: 1.5rem 1rem; background: var(--nh-surface); border-right: 1px solid var(--nh-border); }
  .nh-skel { display: block; height: 1.5rem; width: 65%; }
  .loading-main { min-width: 0; }
  .loading-header { height: var(--nh-header-h); display: flex; align-items: center; padding: 0 var(--nh-space); border-bottom: 1px solid var(--nh-border); }
  .loading-header .nh-skel { width: 8rem; }
  .loading-content { display: grid; gap: 1.5rem; padding: 1.5rem; }
  .loading-card { width: 100%; height: 16rem; }
  @media (max-width: 960px) { .shell-loading { grid-template-columns: minmax(0, 1fr); } aside { display: none; } }
</style>
