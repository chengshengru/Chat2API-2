// Chat2API 前端主入口
// Wails v2 会自动注入 window.go / window.main 对象
// 我们通过 window.go.main.App 访问 Go 后端

console.log("Chat2API frontend starting...");

document.addEventListener("DOMContentLoaded", () => {
  const app = document.getElementById("app");
  if (!app) return;

  // 基础 UI
  app.innerHTML = `
    <div style="font-family: system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif; padding: 20px; max-width: 1200px; margin: 0 auto;">
      <header style="border-bottom: 1px solid #eee; padding-bottom: 16px; margin-bottom: 20px;">
        <h1 style="margin: 0; color: #1f2937;">Chat2API</h1>
        <p style="color: #6b7280; margin: 4px 0 0 0;">Multi-platform AI Service Unified Management Tool</p>
      </header>

      <div id="status-card" style="background: #f3f4f6; border-radius: 8px; padding: 16px; margin-bottom: 20px;">
        <h3 style="margin: 0 0 8px 0; color: #1f2937;">Application Status</h3>
        <div id="status-content" style="color: #4b5563; font-size: 14px;">
          Loading...
        </div>
      </div>

      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 16px; margin-bottom: 20px;">
        <div id="providers-card" style="background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px;">
          <h3 style="margin: 0 0 12px 0; color: #1f2937;">Providers</h3>
          <div id="providers-content" style="color: #6b7280; font-size: 14px;">Loading...</div>
        </div>

        <div id="accounts-card" style="background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px;">
          <h3 style="margin: 0 0 12px 0; color: #1f2937;">Accounts</h3>
          <div id="accounts-content" style="color: #6b7280; font-size: 14px;">Loading...</div>
        </div>

        <div id="proxy-card" style="background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px;">
          <h3 style="margin: 0 0 12px 0; color: #1f2937;">Proxy Server</h3>
          <div id="proxy-content" style="color: #6b7280; font-size: 14px;">Loading...</div>
          <div style="margin-top: 12px; display: flex; gap: 8px;">
            <button id="btn-start-proxy" style="padding: 6px 16px; background: #22c55e; color: white; border: none; border-radius: 6px; cursor: pointer;">Start</button>
            <button id="btn-stop-proxy" style="padding: 6px 16px; background: #ef4444; color: white; border: none; border-radius: 6px; cursor: pointer;">Stop</button>
            <button id="btn-refresh" style="padding: 6px 16px; background: #3b82f6; color: white; border: none; border-radius: 6px; cursor: pointer;">Refresh</button>
          </div>
        </div>
      </div>

      <div style="background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px;">
        <h3 style="margin: 0 0 12px 0; color: #1f2937;">Recent Logs</h3>
        <pre id="logs-content" style="background: #111827; color: #e5e7eb; padding: 12px; border-radius: 6px; font-size: 12px; overflow-x: auto; max-height: 400px; margin: 0;">Loading...</pre>
      </div>
    </div>
  `;

  // 绑定事件
  document.getElementById("btn-start-proxy").addEventListener("click", async () => {
    try {
      const result = await window.go.main.App.GetConfig();
      const ok = await window.go.main.App.StartProxy(result.ProxyPort, result.ProxyHost);
      refreshStatus();
      alert("Proxy " + (ok ? "started" : "failed to start"));
    } catch (e) {
      alert("Failed to start proxy: " + e);
    }
  });

  document.getElementById("btn-stop-proxy").addEventListener("click", async () => {
    try {
      const ok = await window.go.main.App.StopProxy();
      refreshStatus();
      alert("Proxy " + (ok ? "stopped" : "failed to stop"));
    } catch (e) {
      alert("Failed to stop proxy: " + e);
    }
  });

  document.getElementById("btn-refresh").addEventListener("click", refreshStatus);

  // 定期刷新
  setInterval(refreshStatus, 3000);
  refreshStatus();
});

async function refreshStatus() {
  try {
    const app = window.go.main.App;
    const [info, providers, accounts, proxyStatus, logs] = await Promise.all([
      app.GetAppInfo(),
      app.GetProviders(),
      app.GetAccounts(false),
      app.GetProxyStatus(),
      app.GetRecentLogs(20),
    ]);

    const statusEl = document.getElementById("status-content");
    if (statusEl && info) {
      statusEl.innerHTML = `
        <div><strong>Version:</strong> ${info.version || "1.0.0"}</div>
        <div><strong>Config:</strong> ${info.config ? JSON.stringify(info.config, null, 0) : "N/A"}</div>
      `;
    }

    const providersEl = document.getElementById("providers-content");
    if (providersEl) {
      if (Array.isArray(providers) && providers.length > 0) {
        providersEl.innerHTML = providers.map(p => `<div style="margin-bottom: 6px;">• ${p.Name || p.name || "unknown"} <span style="color: ${p.Enabled ? "#22c55e" : "#ef4444"}">(${p.Enabled ? "enabled" : "disabled"})</span></div>`).join("");
      } else {
        providersEl.innerHTML = "No providers configured";
      }
    }

    const accountsEl = document.getElementById("accounts-content");
    if (accountsEl) {
      if (Array.isArray(accounts) && accounts.length > 0) {
        accountsEl.innerHTML = accounts.map(a => `<div style="margin-bottom: 6px;">• ProviderID: ${a.ProviderID || "N/A"} <span style="color: ${a.Status === "active" ? "#22c55e" : "#6b7280"}">(${a.Status})</span></div>`).join("");
      } else {
        accountsEl.innerHTML = "No accounts configured";
      }
    }

    const proxyEl = document.getElementById("proxy-content");
    if (proxyEl) {
      proxyEl.innerHTML = `
        <div><strong>Running:</strong> <span style="color: ${proxyStatus.IsRunning ? "#22c55e" : "#6b7280"}">${proxyStatus.IsRunning}</span></div>
        <div><strong>Host:</strong> ${proxyStatus.Host || "N/A"}</div>
        <div><strong>Port:</strong> ${proxyStatus.Port || "N/A"}</div>
        <div><strong>Uptime:</strong> ${proxyStatus.Uptime || 0}s</div>
      `;
    }

    const logsEl = document.getElementById("logs-content");
    if (logsEl) {
      if (Array.isArray(logs) && logs.length > 0) {
        logsEl.innerHTML = logs.slice(0, 20).map(l => {
          const time = new Date((l.Timestamp || Date.now() / 1000) * 1000).toLocaleTimeString();
          return `[${time}] [${l.Level || "info"}] ${l.Message || ""}`;
        }).join("\n");
      } else {
        logsEl.textContent = "No logs";
      }
    }
  } catch (e) {
    const logsEl = document.getElementById("logs-content");
    if (logsEl) {
      logsEl.textContent = "Error loading status: " + e;
    }
  }
}
