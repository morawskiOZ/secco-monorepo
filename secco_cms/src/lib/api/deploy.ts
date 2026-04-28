export async function triggerDeploy(): Promise<{ snapshot_url: string; run_url: string }> {
  const res = await fetch('/api/deploy', { method: 'POST' });
  if (!res.ok) throw new Error('Deploy failed');
  return res.json();
}

export async function getDeployStatus(): Promise<{ status: string; last_deploy: string | null }> {
  const res = await fetch('/api/deploy/status');
  if (!res.ok) throw new Error('Failed to get deploy status');
  return res.json();
}
