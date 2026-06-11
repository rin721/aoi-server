<template>
  <div class="admin-shell">
    <AppSidebar />
    <div class="admin-main">
      <AppTopBar />
      <main class="admin-content">
        <slot />
      </main>
    </div>
  </div>
</template>
