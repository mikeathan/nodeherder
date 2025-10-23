<script setup lang="ts">
  import { useAuth } from '@/mixins/composables/useAuthentication';
  import { ref, onMounted } from 'vue';
  import { getOAuthUrl, waitForOAuthCompletion } from '@/services/auth.service';
  import { navigatePostLogin } from '@/router/navigation';
  import { useRoute } from 'vue-router';
  import { store } from '@/store';
  import GoogleIcon from '@/components/icons/GoogleIcon.vue';

  const error = ref<string>('');
  const loading = ref(false);
  const isProcessing = ref(false);
  const route = useRoute();
  const { user } = useAuth();

  // Check if we're returning from OAuth
  onMounted(async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const authStatus = urlParams.get('auth');

    if (authStatus === 'success') {
      // Clear the URL parameter
      const newUrl = window.location.pathname;
      window.history.replaceState({}, document.title, newUrl);

      // Check if user is authenticated and redirect
      try {
        const response = await fetch('http://localhost:4110/api/auth/me', {
          credentials: 'include',
          method: 'GET',
        });
        if (response.ok) {
          const userData = await response.json();
          if (userData.id && userData.username) {
            const userSession = {
              user: userData,
              isAuthenticated: true,
            };
            store.dispatch('auth/loginUser', userSession);
            navigatePostLogin(route, true);
            return;
          }
        }
      } catch (e) {
        console.error('Failed to verify authentication:', e);
      }

      // If we get here, auth failed
      error.value = 'Authentication verification failed. Please try again.';
    }
  });

  const handleLogin = async () => {
    if (isProcessing.value || loading.value) {
      return;
    }

    isProcessing.value = true;
    loading.value = true;
    error.value = '';

    try {
      // Get OAuth URL and redirect directly in the same window
      const url = await getOAuthUrl();

      // Store the fact that we're doing OAuth so we can handle the return
      sessionStorage.setItem('oauth_in_progress', 'true');
      sessionStorage.setItem('oauth_return_url', window.location.href);

      // Redirect to OAuth URL in the same window
      window.location.href = url;
    } catch (e: any) {
      console.error('Login failed', e);
      error.value = e.message || 'Login failed. Please try again.';
      loading.value = false;
      isProcessing.value = false;
    }
  };
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <!-- Logo -->
      <div class="logo-container">
        <img src="@/assets/images/nodeherder_logo.png" alt="Logo" class="logo" />
      </div>

      <!-- Welcome Text -->
      <div class="welcome-section">
        <h1 class="welcome-title">Welcome Back</h1>
        <p class="welcome-subtitle">Sign in to continue to your account</p>
      </div>

      <!-- Login Button -->
      <div class="button-container">
        <button class="google-button" :disabled="loading" @click.prevent="handleLogin">
          <GoogleIcon class="icon" />
          <span>{{ loading ? 'Signing in...' : 'Sign in with Google' }}</span>
        </button>

        <!-- Error Message -->
        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <!-- Redirect Notice -->
        <p class="redirect-notice">A popup window will open for secure authentication</p>
      </div>

      <!-- Security Badge -->
      <div class="security-badge">
        <svg class="security-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            d="M12 2L4 6V12C4 16.55 7.16 20.74 12 22C16.84 20.74 20 16.55 20 12V6L12 2Z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round" />
          <path
            d="M9 12L11 14L15 10"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round" />
        </svg>
        <span>Secure Login</span>
      </div>
    </div>

    <!-- Footer -->
    <div class="footer">
      <i class="footer-link">Copyright © 2025 NodeHerder</i>
      <!-- <span class="footer-divider">•</span> -->
    </div>
  </div>
</template>

<style scoped>
  .login-container {
    min-height: 100vh;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
    padding: 20px;
    position: relative;
  }

  .login-card {
    background: rgba(30, 30, 30, 0.95);
    border-radius: 24px;
    padding: 48px 40px;
    max-width: 450px;
    width: 100%;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .logo-container {
    display: flex;
    justify-content: center;
    margin-bottom: 32px;
  }

  .logo {
    width: 120px;
    height: 120px;
    object-fit: contain;
  }

  .welcome-section {
    text-align: center;
    margin-bottom: 40px;
  }

  .welcome-title {
    font-size: 28px;
    font-weight: 600;
    color: #e3e3e3;
    margin: 0 0 8px 0;
  }

  .welcome-subtitle {
    font-size: 14px;
    color: #a0a0a0;
    margin: 0;
  }

  .button-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-bottom: 32px;
  }

  .google-button {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    font-weight: 500;
    transition: all 0.2s ease;
    width: 100%;
    background-color: #131314;
    border: 1px solid #747775;
    border-radius: 20px;
    color: #e3e3e3;
    cursor: pointer;
    font-size: 14px;
    height: 48px;
    letter-spacing: 0.25px;
    outline: none;
    padding: 0 24px;
  }

  .google-button:hover:not(:disabled) {
    background-color: #1f1f1f;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    border-color: #8f8f8f;
  }

  .google-button:active:not(:disabled) {
    transform: scale(0.98);
  }

  .google-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .icon {
    width: 20px;
    height: 20px;
  }

  .error-message {
    color: #ff6b6b;
    font-size: 14px;
    text-align: center;
    padding: 12px;
    background: rgba(255, 107, 107, 0.1);
    border-radius: 8px;
    border: 1px solid rgba(255, 107, 107, 0.3);
  }

  .redirect-notice {
    font-size: 12px;
    color: #808080;
    text-align: center;
    margin: 0;
  }

  .security-badge {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px;
    background: rgba(77, 208, 225, 0.1);
    border: 1px solid rgba(77, 208, 225, 0.3);
    border-radius: 12px;
    margin-bottom: 32px;
    color: #4dd0e1;
    font-size: 13px;
  }

  .security-icon {
    width: 18px;
    height: 18px;
    color: #4dd0e1;
  }

  .footer {
    position: absolute;
    bottom: 24px;
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 13px;
  }

  .footer-link {
    color: #808080;
    text-decoration: none;
    transition: color 0.2s;
  }

  .footer-link:hover {
    color: #4dd0e1;
  }

  .footer-divider {
    color: #4a4a4a;
  }

  @media (max-width: 768px) {
    .login-card {
      padding: 32px 24px;
    }

    .welcome-title {
      font-size: 24px;
    }

    .footer {
      flex-direction: column;
      gap: 8px;
    }

    .footer-divider {
      display: none;
    }
  }
</style>
