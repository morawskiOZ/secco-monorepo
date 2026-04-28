<script lang="ts">
  import { page } from '$app/state';
  import { getAuth } from '$lib/stores/auth.svelte';
  import { triggerDeploy } from '$lib/api/deploy';

  const auth = getAuth();

  let mobileOpen = $state(false);
  let deploying = $state(false);
  let deployMessage = $state('');

  const navLinks = [
    { href: '/', label: 'Dashboard', icon: '⌂' },
    { href: '/tresci', label: 'Treści', icon: '✎' },
    { href: '/projekty', label: 'Projekty', icon: '▣' },
    { href: '/proces', label: 'Proces', icon: '⚙' },
    { href: '/images', label: 'Zdjęcia', icon: '▣' }
  ];

  function isActive(href: string): boolean {
    const pathname = page.url.pathname;
    if (href === '/') return pathname === '/';
    return pathname.startsWith(href);
  }

  async function handleDeploy() {
    if (deploying) return;
    deploying = true;
    deployMessage = '';
    try {
      const result = await triggerDeploy();
      deployMessage = 'Deploy uruchomiony!';
      setTimeout(() => { deployMessage = ''; }, 3000);
    } catch (err) {
      deployMessage = 'Błąd deploy!';
      setTimeout(() => { deployMessage = ''; }, 3000);
    } finally {
      deploying = false;
    }
  }

  async function handleLogout() {
    await auth.logout();
  }

  function toggleMobile() {
    mobileOpen = !mobileOpen;
  }

  function closeMobile() {
    mobileOpen = false;
  }
</script>

<!-- Mobile hamburger -->
<button class="hamburger" onclick={toggleMobile} aria-label="Menu">
  <span class="hamburger-line"></span>
  <span class="hamburger-line"></span>
  <span class="hamburger-line"></span>
</button>

<!-- Overlay for mobile -->
{#if mobileOpen}
  <div class="sidebar-overlay" onclick={closeMobile} role="presentation"></div>
{/if}

<aside class="sidebar" class:open={mobileOpen}>
  <div class="sidebar-header">
    <a href="/" class="sidebar-logo" onclick={closeMobile}>Secco CMS</a>
  </div>

  <div class="sidebar-deploy">
    <button
      class="btn btn-primary deploy-btn"
      onclick={handleDeploy}
      disabled={deploying}
    >
      {deploying ? 'Deploying...' : 'Deploy'}
    </button>
    {#if deployMessage}
      <span class="deploy-message">{deployMessage}</span>
    {/if}
  </div>

  <nav class="sidebar-nav">
    {#each navLinks as link}
      <a
        href={link.href}
        class="nav-link"
        class:active={isActive(link.href)}
        onclick={closeMobile}
      >
        <span class="nav-icon">{link.icon}</span>
        {link.label}
      </a>
    {/each}
  </nav>

  <div class="sidebar-footer">
    <button class="nav-link logout-btn" onclick={handleLogout}>
      Wyloguj się
    </button>
  </div>
</aside>

<style>
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    width: var(--sidebar-width);
    height: 100vh;
    background: var(--color-white);
    border-right: 1px solid #e0e0e0;
    display: flex;
    flex-direction: column;
    z-index: 100;
    overflow-y: auto;
  }

  .sidebar-header {
    padding: var(--spacing-lg);
    border-bottom: 1px solid #f0f0f0;
  }

  .sidebar-logo {
    font-size: var(--font-size-lg);
    font-weight: 700;
    color: var(--color-accent);
    text-decoration: none;
  }

  .sidebar-logo:hover {
    text-decoration: none;
  }

  .sidebar-deploy {
    padding: var(--spacing-md) var(--spacing-lg);
    border-bottom: 1px solid #f0f0f0;
  }

  .deploy-btn {
    width: 100%;
    justify-content: center;
    padding: var(--spacing-sm) var(--spacing-md);
    font-size: var(--font-size-sm);
  }

  .deploy-message {
    display: block;
    text-align: center;
    font-size: var(--font-size-xs);
    margin-top: var(--spacing-xs);
    color: var(--color-success);
  }

  .sidebar-nav {
    flex: 1;
    padding: var(--spacing-md) 0;
  }

  .nav-link {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-sm) var(--spacing-lg);
    color: var(--color-text);
    text-decoration: none;
    font-size: var(--font-size-sm);
    font-weight: 500;
    transition: background var(--transition-fast), color var(--transition-fast);
    border: none;
    background: none;
    width: 100%;
    text-align: left;
    cursor: pointer;
  }

  .nav-link:hover {
    background: #f5f5f5;
    text-decoration: none;
  }

  .nav-link.active {
    background: #eef0ff;
    color: var(--color-accent);
    font-weight: 600;
    border-right: 3px solid var(--color-accent);
  }

  .nav-icon {
    width: 20px;
    text-align: center;
    font-size: var(--font-size-base);
  }

  .sidebar-footer {
    padding: var(--spacing-md) 0;
    border-top: 1px solid #f0f0f0;
  }

  .logout-btn {
    color: var(--color-accent-secondary);
    font-family: inherit;
  }

  .logout-btn:hover {
    color: var(--color-error);
  }

  .hamburger {
    display: none;
    position: fixed;
    top: var(--spacing-md);
    left: var(--spacing-md);
    z-index: 200;
    background: var(--color-white);
    border: 1px solid #e0e0e0;
    border-radius: var(--radius-sm);
    padding: var(--spacing-sm);
    flex-direction: column;
    gap: 4px;
    cursor: pointer;
  }

  .hamburger-line {
    display: block;
    width: 20px;
    height: 2px;
    background: var(--color-text);
    border-radius: 1px;
  }

  .sidebar-overlay {
    display: none;
  }

  @media (max-width: 767px) {
    .hamburger {
      display: flex;
    }

    .sidebar {
      transform: translateX(-100%);
      transition: transform var(--transition-normal);
    }

    .sidebar.open {
      transform: translateX(0);
    }

    .sidebar-overlay {
      display: block;
      position: fixed;
      inset: 0;
      background: rgba(0, 0, 0, 0.4);
      z-index: 99;
    }
  }
</style>
