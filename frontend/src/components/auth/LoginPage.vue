<script setup lang="ts">
  import { useAuth } from '@/mixins/composables/useAuthentication';
  import { ref } from 'vue';

  import GoogleIcon from '@/components/icons/GoogleIcon.vue';

  const error = ref<string>('');
  const loading = ref(false);

  const { user, signIn } = useAuth();

  const handleLogin = async () => {
    loading.value = true;
    error.value = '';
    try {
      await signIn();
    } catch (e) {
      console.error('Login failed', e);
      error.value = 'Login failed. Please try again.';
    } finally {
      loading.value = false;
    }
  };
</script>

<style scoped>
  .login-wrapper {
    height: 20vh;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
  }

  .error-message {
    color: red;
    font-size: 0.875rem;
  }
</style>

<!-- <template>
  <div class="login-wrapper flex flex-column align-items-center justify-content-center gap-3">
    <Button @click="handleLogin" :disabled="loading">
      <span v-if="!loading">Login with Google</span>
      <span v-else>Signing in...</span>
    </Button> 
    <Button @click="handleLogin" :disabled="loading" class="google-login-button flex align-items-center gap-2">
      <GoogleIcon class="google-icon" />
      <span v-if="!loading">Sign in with Google</span>
      <span v-else>Signing in...</span>
    </Button>
    <p v-if="error" class="error-message">{{ error }}</p>
  </div>
</template> -->
<style scoped>
.gsi-material-button {
  -moz-user-select: none;
  -webkit-user-select: none;
  -ms-user-select: none;
  -webkit-appearance: none;
  background-color: white;
  border: 1px solid #747775;
  border-radius: 4px;
  box-sizing: border-box;
  color: #1f1f1f;
  cursor: pointer;
  font-family: 'Roboto', arial, sans-serif;
  font-size: 14px;
  height: 40px;
  letter-spacing: 0.25px;
  outline: none;
  overflow: hidden;
  padding: 0 12px;
  position: relative;
  text-align: center;
  transition: background-color .218s, border-color .218s, box-shadow .218s;
  vertical-align: middle;
  white-space: nowrap;
  width: auto;
  max-width: 400px;
  min-width: min-content;
  display: flex;
  align-items: center;
}

.gsi-material-button .gsi-material-button-icon {
  height: 20px;
  margin-right: 12px;
  width: 20px;
}

.gsi-material-button .gsi-material-button-content-wrapper {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
}

.gsi-material-button .gsi-material-button-contents {
  flex-grow: 1;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: top;
}

.gsi-material-button .gsi-material-button-state {
  transition: opacity .218s;
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  opacity: 0;
}

.gsi-material-button:disabled {
  cursor: default;
  background-color: #ffffff61;
  border-color: #1f1f1f1f;
}

.gsi-material-button:disabled .gsi-material-button-contents,
.gsi-material-button:disabled .gsi-material-button-icon {
  opacity: 38%;
}

.gsi-material-button:not(:disabled):active .gsi-material-button-state,
.gsi-material-button:not(:disabled):focus .gsi-material-button-state {
  background-color: #303030;
  opacity: 12%;
}

.gsi-material-button:not(:disabled):hover {
  box-shadow: 0 1px 2px 0 rgba(60, 64, 67, .30), 0 1px 3px 1px rgba(60, 64, 67, .15);
}

.gsi-material-button:not(:disabled):hover .gsi-material-button-state {
  background-color: #303030;
  opacity: 8%;
}
</style>


<template>
  <div class="login-wrapper flex flex-column align-items-center justify-content-center gap-3">
    <button
      @click="handleLogin"
      :disabled="loading"
      class="gsi-material-button flex align-items-center"
    >
      <div class="gsi-material-button-state"></div>
      <div class="gsi-material-button-content-wrapper">
        <div class="gsi-material-button-icon">
          <GoogleIcon />
        </div>
        <span class="gsi-material-button-contents">
          {{ loading ? 'Signing in...' : 'Sign in with Google' }}
        </span>
      </div>
    </button>
  </div>
</template>