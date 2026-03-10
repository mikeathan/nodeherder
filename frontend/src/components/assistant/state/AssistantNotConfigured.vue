<template>
  <div
    class="absolute top-0 left-0 w-full h-full flex flex-column align-items-center justify-content-center px-4 pb-8 fadein z-1">
    <div class="text-center">
      <i class="pi pi-cog text-6xl text-color-secondary mb-4"></i>
      <h2 class="text-xl font-normal text-color-secondary m-0">Assistant API URL not configured</h2>
      <p class="text-color-secondary mt-2">
        Please configure the Assistant API URL in
        <router-link to="/settings" class="text-primary no-underline hover:underline cursor-pointer"
          >Settings</router-link
        >.
      </p>
    </div>
  </div>
</template>
