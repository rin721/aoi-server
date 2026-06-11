<template>
  <main class="auth-layout">
    <section class="auth-card">
      <div class="auth-card__visual">
        <div class="auth-card__brand">
          <span class="auth-card__mark">A</span>
          <div>
            <h1>Aoi Admin</h1>
            <p>面向 go-scaffold 的 IAM 管理台，聚合组织、角色、会话和审计工作流。</p>
          </div>
        </div>
        <div class="auth-card__visual-panel" aria-hidden="true">
          <span class="auth-card__visual-line" />
          <span class="auth-card__visual-code">GET /api/v1/me</span>
          <span class="auth-card__visual-code">RBAC · Audit · Session</span>
        </div>
      </div>
      <div class="auth-card__form">
        <slot />
      </div>
    </section>
  </main>
</template>


