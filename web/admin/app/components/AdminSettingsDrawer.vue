<script setup lang="ts">
defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  "update:open": [value: boolean]
}>()

const { accentPresets, reset, state } = useAdminUiPreferences()

const themeItems = [
  { icon: "sun", label: "浅色", value: "light" },
  { icon: "moon", label: "深色", value: "dark" },
  { icon: "monitor", label: "跟随系统", value: "system" }
]

const densityItems = [
  { icon: "rows-3", label: "舒适", value: "comfortable" },
  { icon: "list", label: "紧凑", value: "compact" }
]

const infoLayoutItems = [
  { icon: "layout-grid", label: "自适应", value: "adaptive" },
  { icon: "columns-3", label: "瀑布流", value: "masonry" },
  { icon: "list", label: "列表", value: "list" }
]
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="admin-drawer-layer">
      <button class="admin-drawer-layer__scrim" type="button" aria-label="关闭设置" @click="emit('update:open', false)" />
      <aside class="admin-settings-drawer" aria-label="后台设置">
        <header class="admin-settings-drawer__header">
          <div>
            <strong>后台设置</strong>
            <span>外观、布局和可访问性偏好</span>
          </div>
          <AoiIconButton icon="x" label="关闭设置" @click="emit('update:open', false)" />
        </header>

        <section class="admin-settings-drawer__section">
          <h3>主题模式</h3>
          <AoiSegmentedControl v-model="state.theme" :items="themeItems" aria-label="主题模式" :columns="3" />
        </section>

        <section class="admin-settings-drawer__section">
          <h3>主题色</h3>
          <div class="admin-color-grid">
            <button
              v-for="preset in accentPresets"
              :key="preset.value"
              class="admin-color-swatch"
              :class="{ 'admin-color-swatch--active': state.accent === preset.value }"
              type="button"
              :style="{ '--admin-swatch': preset.accent60 }"
              @click="state.accent = preset.value"
            >
              <span />
              {{ preset.label }}
            </button>
          </div>
        </section>

        <section class="admin-settings-drawer__section">
          <h3>界面密度</h3>
          <AoiSegmentedControl v-model="state.density" :items="densityItems" aria-label="界面密度" :columns="2" />
        </section>

        <section class="admin-settings-drawer__section">
          <h3>信息面板布局</h3>
          <AoiSegmentedControl v-model="state.infoLayout" :items="infoLayoutItems" aria-label="信息面板布局" :columns="3" />
        </section>

        <section class="admin-settings-drawer__section admin-toggle-list">
          <AoiSwitch v-model="state.contrast" label="高对比度" />
          <AoiSwitch v-model="state.reducedMotion" label="减少动效" />
          <AoiSwitch v-model="state.watermark" label="显示水印" />
        </section>

        <footer class="admin-settings-drawer__footer">
          <AoiButton appearance="soft" icon="rotate-ccw" intent="neutral" @click="reset">
            重置偏好
          </AoiButton>
        </footer>
      </aside>
    </div>
  </Teleport>
</template>


