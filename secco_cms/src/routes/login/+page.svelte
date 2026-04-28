<script lang="ts">
  import { goto } from '$app/navigation';
  import { getAuth } from '$lib/stores/auth.svelte';

  const auth = getAuth();

  let username = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (loading) return;

    error = '';
    loading = true;

    const success = await auth.login(username, password);
    if (success) {
      goto('/');
    } else {
      error = 'Nieprawidłowe dane logowania';
    }

    loading = false;
  }
</script>

<div class="login-page">
  <div class="login-card">
    <h1 class="login-title">Secco CMS</h1>
    <form onsubmit={handleSubmit}>
      <div class="form-group">
        <label for="username">Login</label>
        <input
          id="username"
          type="text"
          bind:value={username}
          required
          autocomplete="username"
          disabled={loading}
        />
      </div>
      <div class="form-group">
        <label for="password">Hasło</label>
        <input
          id="password"
          type="password"
          bind:value={password}
          required
          autocomplete="current-password"
          disabled={loading}
        />
      </div>
      {#if error}
        <p class="login-error">{error}</p>
      {/if}
      <button type="submit" class="btn btn-primary login-btn" disabled={loading}>
        {loading ? 'Logowanie...' : 'Zaloguj się'}
      </button>
    </form>
  </div>
</div>

<style>
  .login-page {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: var(--color-bg-page);
  }

  .login-card {
    background: var(--color-white);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    padding: var(--spacing-2xl);
    width: 100%;
    max-width: 400px;
    margin: var(--spacing-md);
  }

  .login-title {
    font-size: var(--font-size-2xl);
    font-weight: 700;
    color: var(--color-accent);
    text-align: center;
    margin-bottom: var(--spacing-xl);
  }

  .form-group {
    margin-bottom: var(--spacing-md);
  }

  .form-group label {
    display: block;
    font-size: var(--font-size-sm);
    font-weight: 600;
    color: var(--color-text);
    margin-bottom: var(--spacing-xs);
  }

  .form-group input {
    width: 100%;
    padding: var(--spacing-sm) var(--spacing-md);
    border: 1px solid #d0d0d0;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-base);
    transition: border-color var(--transition-fast);
  }

  .form-group input:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(46, 49, 146, 0.15);
  }

  .form-group input:disabled {
    background: #f5f5f5;
  }

  .login-error {
    color: var(--color-error);
    font-size: var(--font-size-sm);
    margin-bottom: var(--spacing-md);
    text-align: center;
  }

  .login-btn {
    width: 100%;
    justify-content: center;
    padding: var(--spacing-sm) var(--spacing-md);
    font-size: var(--font-size-base);
  }
</style>
