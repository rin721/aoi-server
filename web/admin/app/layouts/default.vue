<script setup lang="ts">
const settingsOpen = ref(false)
</script>

<template>
  <div class="admin-shell">
    <AdminSidebar />
    <div class="admin-main">
      <AdminTopBar @open-settings="settingsOpen = true" />
      <AdminVisitedTabs />
      <main class="admin-content">
        <slot />
      </main>
    </div>
    <AdminMobileNav />
    <AdminSettingsDrawer v-model:open="settingsOpen" />
  </div>
</template>


