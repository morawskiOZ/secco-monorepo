<script lang="ts">
  import '$lib/styles/tokens.css';
  import '../app.css';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { getAuth } from '$lib/stores/auth.svelte';
  import Sidebar from '$lib/components/Sidebar.svelte';

  let { children } = $props();

  const auth = getAuth();

  $effect(() => {
    auth.checkAuth();
  });

  $effect(() => {
    if (auth.checking) return;

    const pathname = page.url.pathname;
    const isLoginPage = pathname === '/login';

    if (!auth.authenticated && !isLoginPage) {
      goto('/login');
    } else if (auth.authenticated && isLoginPage) {
      goto('/');
    }
  });
</script>

{#if auth.checking}
  <div class="spinner-container">
    <div class="spinner"></div>
  </div>
{:else if !auth.authenticated}
  {@render children()}
{:else if page.url.pathname === '/login'}
  {@render children()}
{:else}
  <div class="admin-layout">
    <Sidebar />
    <main class="admin-main">
      {@render children()}
    </main>
  </div>
{/if}
