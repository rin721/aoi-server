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
        <p>由 Go 服务托管静态管理台，浏览器直接访问同源 IAM 与探针 API。</p>
      </div>
      <div class="auth-card__form">
        <slot />
      </div>
    </section>
  </main>
</template>
