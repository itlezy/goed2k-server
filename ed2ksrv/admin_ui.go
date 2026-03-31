package ed2ksrv

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type adminUIData struct {
	ServerName   string
	Description  string
	TokenEnabled bool
}

var adminPageTemplate = template.Must(template.New("admin-ui").Parse(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.ServerName}} Control</title>
  <link rel="stylesheet" href="/app.css">
</head>
<body data-token-enabled="{{.TokenEnabled}}">
  <div class="ambient ambient-a"></div>
  <div class="ambient ambient-b"></div>
  <main class="shell">
    <header class="hero">
      <div>
        <p class="eyebrow">eD2k Server Console</p>
        <h1>{{.ServerName}}</h1>
        <p class="hero-copy">{{.Description}}</p>
      </div>
      <div class="hero-actions">
        <button id="refreshButton" type="button">Refresh</button>
        <button id="logoutButton" type="button" class="ghost">Logout</button>
      </div>
    </header>

    <section id="loginPanel" class="panel login-panel hidden">
      <div>
        <p class="panel-kicker">Admin Access</p>
        <h2>登录管理界面</h2>
        <p class="muted">输入 X-Admin-Token 对应的令牌后，页面会通过管理 API 拉取统计、客户端、共享文件和审计日志。</p>
      </div>
      <form id="loginForm" class="login-form">
        <label for="adminToken">Admin Token</label>
        <input id="adminToken" name="adminToken" type="password" autocomplete="current-password" placeholder="输入管理令牌">
        <button type="submit">Login</button>
      </form>
    </section>

    <section id="dashboard" class="dashboard hidden">
      <section class="stats-grid">
        <article class="stat-card accent-a">
          <span>当前客户端</span>
          <strong id="statClients">0</strong>
          <small id="statConnections">总连接 0</small>
        </article>
        <article class="stat-card accent-b">
          <span>共享文件</span>
          <strong id="statFiles">0</strong>
          <small id="statRegistered">注册 0 / 移除 0</small>
        </article>
        <article class="stat-card accent-c">
          <span>搜索请求</span>
          <strong id="statSearches">0</strong>
          <small id="statSearchEntries">结果项 0</small>
        </article>
        <article class="stat-card accent-d">
          <span>网络流量</span>
          <strong id="statTraffic">0 B</strong>
          <small id="statPackets">入 0 / 出 0</small>
        </article>
      </section>

      <section class="content-grid">
        <article class="panel wide-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Clients</p>
              <h2>在线客户端</h2>
            </div>
            <label class="compact-field">
              <span>搜索</span>
              <input id="clientSearch" type="search" placeholder="名称、地址、Hash">
            </label>
          </div>
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>名称</th>
                  <th>远端地址</th>
                  <th>监听端点</th>
                  <th>最后活跃</th>
                </tr>
              </thead>
              <tbody id="clientsBody"></tbody>
            </table>
          </div>
          <div class="pager">
            <button id="clientsPrev" type="button" class="ghost">上一页</button>
            <span id="clientsPageInfo">第 1 页</span>
            <button id="clientsNext" type="button" class="ghost">下一页</button>
          </div>
        </article>

        <article class="panel wide-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Files</p>
              <h2>共享文件</h2>
            </div>
            <div class="head-actions">
              <label class="compact-field">
                <span>搜索</span>
                <input id="fileSearch" type="search" placeholder="文件名、Hash">
              </label>
              <label class="compact-field">
                <span>类型</span>
                <input id="fileTypeFilter" type="search" placeholder="Audio / Video / Iso">
              </label>
              <button id="deleteSelectedButton" type="button" class="ghost danger">批量删除</button>
            </div>
          </div>
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th></th>
                  <th>文件名</th>
                  <th>类型</th>
                  <th>大小</th>
                  <th>来源</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody id="filesBody"></tbody>
            </table>
          </div>
          <div class="pager">
            <button id="filesPrev" type="button" class="ghost">上一页</button>
            <span id="filesPageInfo">第 1 页</span>
            <button id="filesNext" type="button" class="ghost">下一页</button>
          </div>
        </article>

        <article class="panel form-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Register</p>
              <h2>新增共享文件</h2>
            </div>
          </div>
          <form id="fileForm" class="file-form">
            <label>
              <span>Hash</span>
              <input name="hash" type="text" maxlength="32" required placeholder="32 位 ED2K Hash">
            </label>
            <label>
              <span>文件名</span>
              <input name="name" type="text" required placeholder="example.iso">
            </label>
            <label>
              <span>大小</span>
              <input name="size" type="number" min="0" required placeholder="4096">
            </label>
            <label>
              <span>类型</span>
              <input name="file_type" type="text" placeholder="Iso">
            </label>
            <label>
              <span>扩展名</span>
              <input name="extension" type="text" placeholder="iso">
            </label>
            <label>
              <span>来源主机</span>
              <input name="host" type="text" placeholder="127.0.0.1">
            </label>
            <label>
              <span>来源端口</span>
              <input name="port" type="number" min="1" max="65535" placeholder="4662">
            </label>
            <button type="submit">保存共享文件</button>
          </form>
        </article>

        <article class="panel detail-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Inspect</p>
              <h2>详情</h2>
            </div>
          </div>
          <pre id="detailView" class="detail-view">选择客户端或文件后，这里显示完整 JSON。</pre>
        </article>

        <article class="panel audit-panel wide-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Audit</p>
              <h2>操作审计日志</h2>
            </div>
          </div>
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>时间</th>
                  <th>动作</th>
                  <th>资源</th>
                  <th>标识</th>
                  <th>来源</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody id="auditBody"></tbody>
            </table>
          </div>
          <div class="pager">
            <button id="auditPrev" type="button" class="ghost">上一页</button>
            <span id="auditPageInfo">第 1 页</span>
            <button id="auditNext" type="button" class="ghost">下一页</button>
          </div>
        </article>
      </section>
    </section>

    <div id="toast" class="toast hidden"></div>
  </main>
  <script src="/app.js"></script>
</body>
</html>`))

func (s *Server) handleAdminUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/clients/") && !strings.HasPrefix(r.URL.Path, "/files/") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := adminUIData{
		ServerName:   s.cfg.ServerName,
		Description:  s.cfg.ServerDescription,
		TokenEnabled: strings.TrimSpace(s.cfg.AdminToken) != "",
	}
	if err := adminPageTemplate.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("render admin ui: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) handleAdminJS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/app.js" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = w.Write([]byte(adminUIScript))
}

func (s *Server) handleAdminCSS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/app.css" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write([]byte(adminUIStyles))
}

const adminUIScript = `(() => {
  const body = document.body;
  const tokenEnabled = body.dataset.tokenEnabled === 'true';
  const loginPanel = document.getElementById('loginPanel');
  const dashboard = document.getElementById('dashboard');
  const toast = document.getElementById('toast');
  const refreshButton = document.getElementById('refreshButton');
  const logoutButton = document.getElementById('logoutButton');
  const loginForm = document.getElementById('loginForm');
  const fileForm = document.getElementById('fileForm');
  const clientSearch = document.getElementById('clientSearch');
  const fileSearch = document.getElementById('fileSearch');
  const fileTypeFilter = document.getElementById('fileTypeFilter');
  const detailView = document.getElementById('detailView');
  const clientsBody = document.getElementById('clientsBody');
  const filesBody = document.getElementById('filesBody');
  const auditBody = document.getElementById('auditBody');
  const tokenInput = document.getElementById('adminToken');
  const deleteSelectedButton = document.getElementById('deleteSelectedButton');

  const state = {
    clientsPage: 1,
    filesPage: 1,
    auditPage: 1,
    clientsPerPage: 10,
    filesPerPage: 10,
    auditPerPage: 10,
    selectedFiles: new Set()
  };

  let adminToken = tokenEnabled ? (window.localStorage.getItem('goed2k_admin_token') || '') : '';

  function setToast(message, isError) {
    toast.textContent = message;
    toast.classList.remove('hidden', 'error');
    if (isError) {
      toast.classList.add('error');
    }
    window.clearTimeout(setToast.timer);
    setToast.timer = window.setTimeout(() => toast.classList.add('hidden'), 2600);
  }

  function setAuthState(authenticated) {
    loginPanel.classList.toggle('hidden', authenticated);
    dashboard.classList.toggle('hidden', !authenticated);
    logoutButton.classList.toggle('hidden', !tokenEnabled || !authenticated);
    refreshButton.disabled = !authenticated;
  }

  async function apiFetch(path, options) {
    const requestOptions = options || {};
    const headers = new Headers(requestOptions.headers || {});
    if (adminToken) {
      headers.set('X-Admin-Token', adminToken);
    }
    if (requestOptions.body && !headers.has('Content-Type')) {
      headers.set('Content-Type', 'application/json');
    }
    const response = await fetch(path, { ...requestOptions, headers });
    const payload = await response.json().catch(() => ({ ok: false, error: 'invalid response' }));
    if (!response.ok || !payload.ok) {
      throw new Error(payload.error || ('request failed: ' + response.status));
    }
    return payload;
  }

  function formatBytes(value) {
    const size = Number(value || 0);
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let index = 0;
    let current = size;
    while (current >= 1024 && index < units.length - 1) {
      current /= 1024;
      index += 1;
    }
    return current.toFixed(index === 0 ? 0 : 1) + ' ' + units[index];
  }

  function formatTime(value) {
    if (!value) {
      return '-';
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString('zh-CN', { hour12: false });
  }

  function updatePager(idPrefix, meta) {
    const currentPage = Number(meta.page || 1);
    const perPage = Number(meta.per_page || 10);
    const total = Number(meta.total || 0);
    const totalPages = Math.max(1, Math.ceil(total / perPage));
    document.getElementById(idPrefix + 'PageInfo').textContent = '第 ' + currentPage + ' / ' + totalPages + ' 页';
    document.getElementById(idPrefix + 'Prev').disabled = currentPage <= 1;
    document.getElementById(idPrefix + 'Next').disabled = currentPage >= totalPages;
  }

  function showDetail(data) {
    detailView.textContent = JSON.stringify(data, null, 2);
  }

  function pushRoute(path) {
    if (window.location.pathname !== path) {
      window.history.pushState({}, '', path);
    }
  }

  async function loadStats() {
    const payload = await apiFetch('/api/stats');
    const data = payload.data;
    document.getElementById('statClients').textContent = data.current_clients;
    document.getElementById('statConnections').textContent = '总连接 ' + data.total_connections;
    document.getElementById('statFiles').textContent = data.current_files;
    document.getElementById('statRegistered').textContent = '注册 ' + data.files_registered + ' / 移除 ' + data.files_removed;
    document.getElementById('statSearches').textContent = data.search_requests;
    document.getElementById('statSearchEntries').textContent = '结果项 ' + data.search_result_entries;
    document.getElementById('statTraffic').textContent = formatBytes((data.inbound_bytes || 0) + (data.outbound_bytes || 0));
    document.getElementById('statPackets').textContent = '入 ' + data.inbound_packets + ' / 出 ' + data.outbound_packets;
  }

  async function loadClients() {
    const query = new URLSearchParams({
      page: String(state.clientsPage),
      per_page: String(state.clientsPerPage),
      sort: 'last_seen_at'
    });
    if (clientSearch.value.trim()) {
      query.set('search', clientSearch.value.trim());
    }
    const payload = await apiFetch('/api/clients?' + query.toString());
    const data = payload.data;
    clientsBody.innerHTML = '';
    data.forEach((client) => {
      const row = document.createElement('tr');
      row.innerHTML =
        '<td>' + client.client_id + '</td>' +
        '<td><a class="row-link" href="/clients/' + client.client_id + '">' + (client.client_name || '-') + '</a></td>' +
        '<td>' + (client.remote_address || '-') + '</td>' +
        '<td>' + (client.listen_endpoint || '-') + '</td>' +
        '<td>' + formatTime(client.last_seen_at) + '</td>';
      row.querySelector('a').addEventListener('click', function (event) {
        event.preventDefault();
        pushRoute('/clients/' + client.client_id);
        loadClientDetail(client.client_id).catch((error) => setToast(error.message, true));
      });
      row.addEventListener('click', function () { showDetail(client); });
      clientsBody.appendChild(row);
    });
    if (!data.length) {
      clientsBody.innerHTML = '<tr><td colspan="5" class="empty">没有在线客户端</td></tr>';
    }
    updatePager('clients', payload.meta || {});
  }

  async function loadClientDetail(clientID) {
    const payload = await apiFetch('/api/clients/' + encodeURIComponent(String(clientID)));
    showDetail(payload.data);
  }

  async function deleteFile(hash) {
    await apiFetch('/api/files/' + encodeURIComponent(hash), { method: 'DELETE' });
    state.selectedFiles.delete(hash);
    setToast('共享文件已删除');
    await refreshAll();
  }

  async function batchDeleteSelected() {
    const hashes = Array.from(state.selectedFiles);
    if (!hashes.length) {
      setToast('没有选中的共享文件', true);
      return;
    }
    await apiFetch('/api/files/batch-delete', {
      method: 'POST',
      body: JSON.stringify({ hashes: hashes })
    });
    state.selectedFiles.clear();
    setToast('批量删除完成');
    await refreshAll();
  }

  async function loadFiles() {
    const query = new URLSearchParams({
      page: String(state.filesPage),
      per_page: String(state.filesPerPage),
      sort: 'sources'
    });
    if (fileSearch.value.trim()) {
      query.set('search', fileSearch.value.trim());
    }
    if (fileTypeFilter.value.trim()) {
      query.set('file_type', fileTypeFilter.value.trim());
    }
    const payload = await apiFetch('/api/files?' + query.toString());
    const data = payload.data;
    filesBody.innerHTML = '';
    data.forEach((file) => {
      const checked = state.selectedFiles.has(file.hash) ? ' checked' : '';
      const row = document.createElement('tr');
      row.innerHTML =
        '<td><input class="file-selector" data-hash="' + file.hash + '" type="checkbox"' + checked + '></td>' +
        '<td><a class="row-link" href="/files/' + file.hash + '">' + (file.name || '-') + '</a></td>' +
        '<td>' + (file.file_type || '-') + '</td>' +
        '<td>' + formatBytes(file.size) + '</td>' +
        '<td>' + (file.sources || 0) + '</td>' +
        '<td><button type="button" class="ghost danger">删除</button></td>';
      row.querySelector('a').addEventListener('click', function (event) {
        event.preventDefault();
        pushRoute('/files/' + file.hash);
        loadFileDetail(file.hash).catch((error) => setToast(error.message, true));
      });
      row.querySelector('.file-selector').addEventListener('change', function (event) {
        if (event.target.checked) {
          state.selectedFiles.add(file.hash);
        } else {
          state.selectedFiles.delete(file.hash);
        }
      });
      row.querySelector('button').addEventListener('click', function (event) {
        event.stopPropagation();
        deleteFile(file.hash).catch((error) => setToast(error.message, true));
      });
      row.addEventListener('click', function (event) {
        if (event.target.tagName !== 'BUTTON' && event.target.tagName !== 'INPUT' && event.target.tagName !== 'A') {
          showDetail(file);
        }
      });
      filesBody.appendChild(row);
    });
    if (!data.length) {
      filesBody.innerHTML = '<tr><td colspan="6" class="empty">没有共享文件</td></tr>';
    }
    updatePager('files', payload.meta || {});
  }

  async function loadFileDetail(hash) {
    const payload = await apiFetch('/api/files/' + encodeURIComponent(hash));
    showDetail(payload.data);
  }

  async function loadAudit() {
    const query = new URLSearchParams({
      page: String(state.auditPage),
      per_page: String(state.auditPerPage)
    });
    const payload = await apiFetch('/api/audit?' + query.toString());
    const data = payload.data;
    auditBody.innerHTML = '';
    data.forEach((entry) => {
      const row = document.createElement('tr');
      row.innerHTML =
        '<td>' + formatTime(entry.time) + '</td>' +
        '<td>' + entry.action + '</td>' +
        '<td>' + entry.resource + '</td>' +
        '<td>' + (entry.resource_id || '-') + '</td>' +
        '<td>' + (entry.remote_addr || '-') + '</td>' +
        '<td>' + entry.status + '</td>';
      row.addEventListener('click', function () { showDetail(entry); });
      auditBody.appendChild(row);
    });
    if (!data.length) {
      auditBody.innerHTML = '<tr><td colspan="6" class="empty">暂无审计日志</td></tr>';
    }
    updatePager('audit', payload.meta || {});
  }

  async function loadRouteDetail() {
    const path = window.location.pathname;
    if (path.indexOf('/clients/') === 0) {
      await loadClientDetail(path.slice('/clients/'.length));
      return;
    }
    if (path.indexOf('/files/') === 0) {
      await loadFileDetail(path.slice('/files/'.length));
    }
  }

  async function refreshAll() {
    try {
      await Promise.all([loadStats(), loadClients(), loadFiles(), loadAudit()]);
      await loadRouteDetail();
      setAuthState(true);
    } catch (error) {
      if (tokenEnabled && /unauthorized/i.test(error.message)) {
        setAuthState(false);
        setToast('令牌无效，请重新登录', true);
        return;
      }
      setToast(error.message, true);
    }
  }

  loginForm.addEventListener('submit', async function (event) {
    event.preventDefault();
    adminToken = tokenInput.value.trim();
    window.localStorage.setItem('goed2k_admin_token', adminToken);
    await refreshAll();
  });

  fileForm.addEventListener('submit', async function (event) {
    event.preventDefault();
    const form = new FormData(fileForm);
    const port = Number(form.get('port') || 0);
    const host = String(form.get('host') || '').trim();
    const payload = {
      hash: String(form.get('hash') || '').trim().toUpperCase(),
      name: String(form.get('name') || '').trim(),
      size: Number(form.get('size') || 0),
      file_type: String(form.get('file_type') || '').trim(),
      extension: String(form.get('extension') || '').trim(),
      endpoints: host && port ? [{ host: host, port: port }] : []
    };
    try {
      await apiFetch('/api/files', { method: 'POST', body: JSON.stringify(payload) });
      fileForm.reset();
      setToast('共享文件已保存');
      await refreshAll();
    } catch (error) {
      setToast(error.message, true);
    }
  });

  refreshButton.addEventListener('click', refreshAll);
  logoutButton.addEventListener('click', function () {
    adminToken = '';
    window.localStorage.removeItem('goed2k_admin_token');
    tokenInput.value = '';
    setAuthState(false);
    setToast('已退出登录');
  });
  deleteSelectedButton.addEventListener('click', function () {
    batchDeleteSelected().catch((error) => setToast(error.message, true));
  });

  document.getElementById('clientsPrev').addEventListener('click', function () {
    if (state.clientsPage > 1) {
      state.clientsPage -= 1;
      loadClients().catch((error) => setToast(error.message, true));
    }
  });
  document.getElementById('clientsNext').addEventListener('click', function () {
    state.clientsPage += 1;
    loadClients().catch((error) => setToast(error.message, true));
  });
  document.getElementById('filesPrev').addEventListener('click', function () {
    if (state.filesPage > 1) {
      state.filesPage -= 1;
      loadFiles().catch((error) => setToast(error.message, true));
    }
  });
  document.getElementById('filesNext').addEventListener('click', function () {
    state.filesPage += 1;
    loadFiles().catch((error) => setToast(error.message, true));
  });
  document.getElementById('auditPrev').addEventListener('click', function () {
    if (state.auditPage > 1) {
      state.auditPage -= 1;
      loadAudit().catch((error) => setToast(error.message, true));
    }
  });
  document.getElementById('auditNext').addEventListener('click', function () {
    state.auditPage += 1;
    loadAudit().catch((error) => setToast(error.message, true));
  });

  let clientSearchTimer = 0;
  let fileSearchTimer = 0;
  clientSearch.addEventListener('input', function () {
    state.clientsPage = 1;
    window.clearTimeout(clientSearchTimer);
    clientSearchTimer = window.setTimeout(function () {
      loadClients().catch((error) => setToast(error.message, true));
    }, 180);
  });
  [fileSearch, fileTypeFilter].forEach(function (node) {
    node.addEventListener('input', function () {
      state.filesPage = 1;
      window.clearTimeout(fileSearchTimer);
      fileSearchTimer = window.setTimeout(function () {
        loadFiles().catch((error) => setToast(error.message, true));
      }, 180);
    });
  });
  window.addEventListener('popstate', function () {
    loadRouteDetail().catch((error) => setToast(error.message, true));
  });

  if (tokenEnabled) {
    tokenInput.value = adminToken;
    if (adminToken) {
      refreshAll();
    } else {
      setAuthState(false);
    }
  } else {
    setAuthState(true);
    refreshAll();
  }
})();`

const adminUIStyles = `:root {
  --bg: #f5efe4;
  --panel: rgba(255, 251, 245, 0.88);
  --panel-strong: rgba(255, 251, 245, 0.96);
  --ink: #182028;
  --muted: #5f6b72;
  --line: rgba(24, 32, 40, 0.12);
  --accent: #0e8f70;
  --accent-2: #d95d39;
  --accent-3: #2a5f9e;
  --shadow: 0 24px 60px rgba(42, 43, 56, 0.12);
  --radius: 24px;
}
* { box-sizing: border-box; }
body {
  margin: 0;
  min-height: 100vh;
  color: var(--ink);
  background:
    radial-gradient(circle at top left, rgba(14, 143, 112, 0.18), transparent 28%),
    radial-gradient(circle at 85% 10%, rgba(217, 93, 57, 0.16), transparent 24%),
    linear-gradient(180deg, #f7f0e7 0%, #ede2d5 100%);
  font-family: "IBM Plex Sans", "Segoe UI", sans-serif;
}
button, input { font: inherit; }
.ambient { position: fixed; border-radius: 999px; filter: blur(20px); opacity: 0.38; pointer-events: none; }
.ambient-a { width: 320px; height: 320px; top: -80px; right: -40px; background: rgba(14, 143, 112, 0.16); }
.ambient-b { width: 280px; height: 280px; left: -60px; bottom: 8%; background: rgba(42, 95, 158, 0.14); }
.shell { position: relative; max-width: 1320px; margin: 0 auto; padding: 32px 20px 60px; }
.hero { display: flex; justify-content: space-between; gap: 24px; align-items: end; margin-bottom: 24px; }
.eyebrow, .panel-kicker { margin: 0 0 10px; letter-spacing: 0.12em; text-transform: uppercase; font-size: 12px; color: var(--accent-3); }
h1, h2 { margin: 0; font-family: "Space Grotesk", "IBM Plex Sans", sans-serif; }
h1 { font-size: clamp(2.4rem, 4vw, 4.6rem); line-height: 0.95; }
h2 { font-size: 1.4rem; }
.hero-copy, .muted { color: var(--muted); max-width: 60ch; }
.hero-actions, .head-actions { display: flex; gap: 12px; align-items: center; }
button { border: 0; border-radius: 999px; padding: 12px 18px; background: var(--ink); color: #fff; cursor: pointer; transition: transform 140ms ease, opacity 140ms ease, background 140ms ease; }
button:hover { transform: translateY(-1px); }
button:disabled { opacity: 0.45; cursor: not-allowed; transform: none; }
button.ghost { background: rgba(24, 32, 40, 0.08); color: var(--ink); }
button.danger { background: rgba(217, 93, 57, 0.12); color: #7e2f19; }
.panel { background: var(--panel); backdrop-filter: blur(14px); border: 1px solid rgba(255, 255, 255, 0.4); border-radius: var(--radius); box-shadow: var(--shadow); padding: 22px; }
.login-panel { max-width: 720px; display: grid; grid-template-columns: 1.2fr 1fr; gap: 24px; align-items: center; }
.login-form, .file-form { display: grid; gap: 14px; }
label { display: grid; gap: 8px; color: var(--muted); }
input { width: 100%; border: 1px solid var(--line); background: var(--panel-strong); border-radius: 16px; padding: 12px 14px; color: var(--ink); }
.dashboard { display: grid; gap: 20px; }
.stats-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; }
.stat-card { position: relative; overflow: hidden; padding: 22px; border-radius: 22px; min-height: 150px; box-shadow: var(--shadow); color: #fff; }
.stat-card::after { content: ""; position: absolute; inset: auto -18% -40% auto; width: 180px; height: 180px; border-radius: 999px; background: rgba(255, 255, 255, 0.12); }
.stat-card span, .stat-card small, .stat-card strong { position: relative; z-index: 1; }
.stat-card strong { display: block; margin: 16px 0 12px; font-size: 2.6rem; font-family: "Space Grotesk", sans-serif; }
.accent-a { background: linear-gradient(135deg, #0e8f70, #126b6d); }
.accent-b { background: linear-gradient(135deg, #d95d39, #9d3d2a); }
.accent-c { background: linear-gradient(135deg, #2a5f9e, #1f4571); }
.accent-d { background: linear-gradient(135deg, #3c4454, #252b36); }
.content-grid { display: grid; grid-template-columns: 1.3fr 1.3fr 0.9fr; gap: 20px; }
.wide-panel { grid-column: span 2; }
.form-panel, .detail-panel { min-height: 420px; }
.audit-panel { grid-column: span 3; }
.panel-head { display: flex; justify-content: space-between; gap: 16px; margin-bottom: 18px; align-items: start; }
.compact-field { min-width: 180px; }
.compact-field span { font-size: 12px; text-transform: uppercase; letter-spacing: 0.08em; }
.table-wrap { overflow: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 14px 12px; border-bottom: 1px solid var(--line); text-align: left; vertical-align: middle; }
tr { transition: background 140ms ease; }
tr:hover { background: rgba(14, 143, 112, 0.06); }
.empty { text-align: center; color: var(--muted); }
.row-link { color: var(--accent-3); text-decoration: none; }
.row-link:hover { text-decoration: underline; }
.pager { display: flex; gap: 12px; justify-content: flex-end; align-items: center; padding-top: 14px; }
.detail-view { margin: 0; height: calc(100% - 52px); min-height: 320px; padding: 16px; background: rgba(24, 32, 40, 0.92); color: #d2ffe6; border-radius: 18px; overflow: auto; font-family: "IBM Plex Mono", monospace; font-size: 13px; }
.toast { position: fixed; right: 20px; bottom: 20px; padding: 14px 18px; border-radius: 16px; background: rgba(24, 32, 40, 0.92); color: #fff; box-shadow: var(--shadow); }
.toast.error { background: rgba(126, 47, 25, 0.94); }
.hidden { display: none !important; }
@media (max-width: 1100px) {
  .stats-grid, .content-grid, .login-panel { grid-template-columns: 1fr 1fr; }
  .wide-panel, .audit-panel { grid-column: span 2; }
}
@media (max-width: 760px) {
  .shell { padding: 20px 14px 40px; }
  .hero, .panel-head, .hero-actions, .head-actions, .login-panel, .stats-grid, .content-grid { grid-template-columns: 1fr; display: grid; }
  .wide-panel, .audit-panel { grid-column: span 1; }
  .hero-actions, .head-actions, .pager { justify-content: stretch; }
  button { width: 100%; }
}`
