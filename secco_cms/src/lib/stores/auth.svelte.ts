let authenticated = $state(false);
let checking = $state(true);

async function checkAuth(): Promise<void> {
  checking = true;
  try {
    const res = await fetch('/api/auth/check', {
      credentials: 'same-origin'
    });
    authenticated = res.ok;
  } catch {
    authenticated = false;
  } finally {
    checking = false;
  }
}

async function login(username: string, password: string): Promise<boolean> {
  try {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify({ username, password })
    });
    if (res.ok) {
      authenticated = true;
      return true;
    }
    return false;
  } catch {
    return false;
  }
}

async function logout(): Promise<void> {
  try {
    await fetch('/api/auth/logout', {
      method: 'POST',
      credentials: 'same-origin'
    });
  } finally {
    authenticated = false;
  }
}

export function getAuth() {
  return {
    get authenticated() {
      return authenticated;
    },
    get checking() {
      return checking;
    },
    checkAuth,
    login,
    logout
  };
}
