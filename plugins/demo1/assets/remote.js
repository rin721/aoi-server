export async function mount(container, context) {
  const userEmail = escapeHTML(context.user?.email || "-")
  const organizationName = escapeHTML(context.organization?.name || "-")

  container.innerHTML = `
    <div class="demo1-plugin">
      <header>
        <strong>Demo1 Sidecar</strong>
        <span>${context.plugin.version}</span>
      </header>
      <section>
        <p>Current user: <b>${userEmail}</b></p>
        <p>Current organization: <b>${organizationName}</b></p>
      </section>
      <button type="button" data-refresh>Refresh sidecar data</button>
      <pre data-output>Loading...</pre>
    </div>
  `

  const style = document.createElement("style")
  style.textContent = `
    .demo1-plugin {
      display: grid;
      gap: 16px;
      min-height: 100%;
      background: var(--aoi-surface, #fff);
      color: var(--aoi-text, #172126);
      padding: 20px;
    }
    .demo1-plugin header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-bottom: 1px solid var(--aoi-border, #d7e2e5);
      padding-bottom: 12px;
    }
    .demo1-plugin button {
      width: fit-content;
      border: 1px solid var(--aoi-border, #d7e2e5);
      border-radius: var(--aoi-radius-control, 8px);
      background: var(--aoi-accent-60, #137b83);
      color: white;
      font: inherit;
      font-weight: 800;
      padding: 9px 12px;
    }
    .demo1-plugin pre {
      overflow: auto;
      border: 1px solid var(--aoi-border, #d7e2e5);
      border-radius: var(--aoi-radius-control, 8px);
      background: rgba(19, 123, 131, .08);
      padding: 12px;
    }
  `
  container.appendChild(style)

  const output = container.querySelector("[data-output]")
  const button = container.querySelector("[data-refresh]")

  async function refresh() {
    output.textContent = "Loading..."
    try {
      const data = await context.request("/api/hello")
      output.textContent = JSON.stringify(data, null, 2)
    } catch (error) {
      output.textContent = error?.message || String(error)
    }
  }

  button.addEventListener("click", refresh)
  await refresh()

  return () => {
    button.removeEventListener("click", refresh)
    container.innerHTML = ""
  }
}

function escapeHTML(value) {
  return String(value).replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "\"": "&quot;",
    "'": "&#39;"
  })[char])
}
