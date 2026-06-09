<script lang="ts">
  import { onMount } from 'svelte';

  const BACKEND_URL = import.meta.env.VITE_BACKEND_URL ?? 'http://localhost:3000';
  const DASHBOARD_ID = import.meta.env.VITE_DASHBOARD_ID ?? '';
  // All Superset traffic proxied through Go backend root — browser never talks to :8088
  const SUPERSET_PROXY_URL = BACKEND_URL;

  let container: HTMLDivElement;
  let error = $state('');
  let loading = $state(true);

  // Fake KPI data to show alongside the embed
  const kpis = [
    { label: 'Total Revenue',  value: '$17.1M', delta: '+12%', up: true  },
    { label: 'Total Orders',   value: '2,000',  delta: '+8%',  up: true  },
    { label: 'Avg Order Value',value: '$8,551',  delta: '-3%',  up: false },
    { label: 'Active Regions', value: '4',      delta: '—',    up: true  },
  ];

  async function fetchGuestToken(): Promise<string> {
    const url = `${BACKEND_URL}/api/guest-token?dashboard_id=${DASHBOARD_ID}`;
    const res = await fetch(url, { method: 'POST' });
    if (!res.ok) throw new Error(`Guest token request failed: ${res.status}`);
    const data = await res.json();
    return data.token;
  }

  onMount(async () => {
    if (!DASHBOARD_ID) {
      error = 'Set VITE_DASHBOARD_ID in your .env';
      loading = false;
      return;
    }
    try {
      const sdk = await import('@superset-ui/embedded-sdk');
      const embedDashboard = sdk.embedDashboard ?? (sdk as any).default?.embedDashboard;
      await embedDashboard({
        id: DASHBOARD_ID,
        supersetDomain: SUPERSET_PROXY_URL,
        mountPoint: container,
        fetchGuestToken,
        dashboardUiConfig: {
          hideTitle: true,        // we show our own title
          filters: { expanded: false },
        },
      });
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      loading = false;
    }
  });
</script>

<!-- ── Top nav ─────────────────────────────────────────────────── -->
<nav>
  <span class="brand">📊 SalesBI</span>
  <ul>
    <li class="active">Overview</li>
    <li>Products</li>
    <li>Regions</li>
    <li>Settings</li>
  </ul>
  <span class="user">Admin ▾</span>
</nav>

<!-- ── Page layout ────────────────────────────────────────────── -->
<div class="page">

  <!-- Sidebar -->
  <aside>
    <p class="section-label">Dashboards</p>
    <ul>
      <li class="active">Sales Overview</li>
      <li>Inventory</li>
      <li>Marketing</li>
      <li>Finance</li>
    </ul>
    <p class="section-label" style="margin-top:2rem">Filters</p>
    <label>
      Year
      <select><option>2024</option><option>2023</option></select>
    </label>
    <label>
      Region
      <select><option>All</option><option>North</option><option>South</option><option>East</option><option>West</option></select>
    </label>
  </aside>

  <!-- Main content -->
  <main>
    <div class="page-header">
      <div>
        <h1>Sales Overview</h1>
        <p class="subtitle">Live data from PostgreSQL · refreshed on demand</p>
      </div>
      <button class="btn">Export ↓</button>
    </div>

    <!-- KPI cards (our own UI, not Superset) -->
    <div class="kpi-grid">
      {#each kpis as k}
        <div class="kpi-card">
          <span class="kpi-label">{k.label}</span>
          <span class="kpi-value">{k.value}</span>
          <span class="kpi-delta" class:up={k.up} class:down={!k.up}>{k.delta}</span>
        </div>
      {/each}
    </div>

    <!-- Section title above the embed -->
    <div class="section-header">
      <h2>Charts</h2>
      <span class="badge">Powered by Superset</span>
    </div>

    <!-- ✦ Superset embed — fixed height, part of the page ✦ -->
    {#if error}
      <div class="error">{error}</div>
    {:else if loading}
      <div class="embed-placeholder">Loading dashboard…</div>
    {/if}
    <div class="embed-box" bind:this={container}></div>

    <!-- Something below the embed -->
    <div class="footer-note">
      Data source: <code>salesdb.sales</code> · 2,000 rows · last seeded on container start
    </div>
  </main>
</div>

<style>
  /* ── Reset ── */
  :global(*, *::before, *::after) { box-sizing: border-box; }
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background: #f0f2f5;
    color: #1a1a2e;
  }

  /* ── Nav ── */
  nav {
    position: sticky;
    top: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    gap: 2rem;
    padding: 0 2rem;
    height: 52px;
    background: #1a1a2e;
    color: #fff;
  }
  .brand { font-weight: 700; font-size: 1.1rem; margin-right: auto; }
  nav ul { display: flex; gap: 1.5rem; list-style: none; margin: 0; padding: 0; }
  nav li { font-size: 0.85rem; opacity: 0.7; cursor: pointer; }
  nav li.active { opacity: 1; border-bottom: 2px solid #6366f1; padding-bottom: 2px; }
  .user { font-size: 0.85rem; opacity: 0.8; cursor: pointer; }

  /* ── Page layout ── */
  .page {
    display: grid;
    grid-template-columns: 200px 1fr;
    min-height: calc(100vh - 52px);
  }

  /* ── Sidebar ── */
  aside {
    background: #fff;
    border-right: 1px solid #e5e7eb;
    padding: 1.5rem 1rem;
  }
  .section-label {
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #9ca3af;
    margin: 0 0 0.5rem 0.5rem;
  }
  aside ul { list-style: none; padding: 0; margin: 0 0 1rem; }
  aside li {
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    font-size: 0.85rem;
    cursor: pointer;
    color: #374151;
  }
  aside li.active { background: #ede9fe; color: #6366f1; font-weight: 600; }
  aside label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.8rem;
    color: #6b7280;
    margin-bottom: 0.75rem;
  }
  aside select {
    padding: 0.3rem 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.85rem;
  }

  /* ── Main ── */
  main { padding: 1.5rem 2rem; overflow-y: auto; }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.25rem;
  }
  h1 { margin: 0; font-size: 1.4rem; font-weight: 700; }
  .subtitle { margin: 0.2rem 0 0; font-size: 0.8rem; color: #9ca3af; }

  .btn {
    padding: 0.4rem 1rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    background: #fff;
    font-size: 0.85rem;
    cursor: pointer;
  }

  /* ── KPI cards ── */
  .kpi-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 1.5rem;
  }
  .kpi-card {
    background: #fff;
    border-radius: 10px;
    padding: 1rem 1.25rem;
    box-shadow: 0 1px 3px rgba(0,0,0,0.07);
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .kpi-label { font-size: 0.75rem; color: #9ca3af; font-weight: 500; }
  .kpi-value { font-size: 1.5rem; font-weight: 700; color: #111827; }
  .kpi-delta { font-size: 0.78rem; font-weight: 600; }
  .kpi-delta.up   { color: #10b981; }
  .kpi-delta.down { color: #ef4444; }

  /* ── Section header ── */
  .section-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }
  h2 { margin: 0; font-size: 1rem; font-weight: 600; }
  .badge {
    font-size: 0.7rem;
    padding: 0.15rem 0.5rem;
    background: #ede9fe;
    color: #6366f1;
    border-radius: 999px;
    font-weight: 600;
  }

  /* ── Embed box — fixed height, NOT fullscreen ── */
  .embed-box {
    width: 100%;
    height: 600px;           /* change this to whatever fits your page */
    border-radius: 10px;
    overflow: hidden;
    background: #fff;
    box-shadow: 0 1px 3px rgba(0,0,0,0.07);
    margin-bottom: 1rem;
  }
  :global(.embed-box iframe) {
    width: 100%;
    height: 100%;
    border: none;
  }

  .embed-placeholder {
    height: 600px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #9ca3af;
    background: #fff;
    border-radius: 10px;
    margin-bottom: 1rem;
  }

  .error {
    padding: 1rem;
    background: #fff8e1;
    border: 1px solid #f9a825;
    border-radius: 8px;
    margin-bottom: 1rem;
    font-size: 0.85rem;
  }

  /* ── Footer note ── */
  .footer-note {
    font-size: 0.75rem;
    color: #9ca3af;
    text-align: center;
    padding: 1rem 0 2rem;
  }
</style>
