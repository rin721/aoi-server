<script setup lang="ts">
import { getLocalizedAoiComponentDocs } from "~/data/aoiComponentDocs"
import type { AoiDanmakuItem, AoiDanmakuMode } from "~/types/danmaku"
import type { AoiDocsLocale } from "~/types/docs"
import type { PlayerPlaybackRate } from "~/types/player"
import type { AoiRgbaColor } from "~/utils/aoiColor"

const props = defineProps<{
  name: string
}>()

type DemoCopy = {
  accentColor: string
  actionDemoLabel: string
  apiLabel: string
  apiTabsLabel: string
  balancedDescription: string
  balancedLabel: string
  cancel: string
  captions: string
  cardLabel: string
  chooseImage: string
  clientHeavyMessage: string
  componentDocsTitle: string
  confirm: string
  copy: string
  createCodeLabel: string
  densityLabel: string
  disabledOption: string
  docsCompletionLabel: string
  docsLabel: string
  docsSectionDescription: string
  docsSectionTitle: string
  docsTagsLabel: string
  demoLabel: string
  eventsLabel: string
  fullscreen: string
  genericFallback: string
  gridItems: string[]
  infoCardDescription: string
  internalDocsLink: string
  keepDocsVisible: string
  layerSafeLabel: string
  layeredLabel: string
  lazyImageAlt: string
  lazyMountMessage: string
  like: string
  linkChip: string
  liveLabel: string
  mediaHeavyMessage: string
  metadataChip: string
  moreActions: string
  openDialog: string
  openGallery: string
  openMenu: string
  outlinedDate: string
  outlinedPalette: string
  outlinedTime: string
  palette: string
  playPreview: string
  propsLabel: string
  publishTime: string
  revealMessage: string
  revealStrength: string
  save: string
  saveLabel: string
  scrollableRow: string
  searchDocs: string
  searchPlaceholder: string
  secondaryBlue: string
  selectedChip: string
  share: string
  snapItems: string[]
  softDescription: string
  softLabel: string
  sourceTag: string
  slotsLabel: string
  stateLabel: string
  statusMessage: string
  switchLabel: string
  sunflowerOrange: string
  toolbarLabel: string
  updatedDate: string
  updatedLabel: string
  vividDescription: string
  vividLabel: string
  wrappersLabel: string
}

const demoCopyByLocale: Record<AoiDocsLocale, DemoCopy> = {
  "zh-CN": {
    accentColor: "强调色",
    actionDemoLabel: "文档动作示例",
    apiLabel: "API",
    apiTabsLabel: "API 选项卡",
    balancedDescription: "默认节奏",
    balancedLabel: "均衡",
    cancel: "取消",
    captions: "字幕",
    cardLabel: "卡片",
    chooseImage: "选择图片",
    clientHeavyMessage: "该组件在浏览器中运行。文档渲染它的 API，并把实时工作台保留为客户端组件。",
    componentDocsTitle: "组件文档",
    confirm: "确认",
    copy: "复制",
    createCodeLabel: "AoiButton 示例",
    densityLabel: "密度",
    disabledOption: "禁用选项",
    docsCompletionLabel: "文档完成度",
    docsLabel: "文档",
    docsSectionDescription: "区块标题让文案、图标和动作保持对齐。",
    docsSectionTitle: "文档区块",
    docsTagsLabel: "文档标签",
    demoLabel: "演示",
    eventsLabel: "事件",
    fullscreen: "全屏",
    genericFallback: "Aoi 组件演示",
    gridItems: ["一", "二", "三"],
    infoCardDescription: "卡片可以承载媒体、文案、元信息、动作和链接行为。",
    internalDocsLink: "内部文档链接",
    keepDocsVisible: "保持文档可见",
    layerSafeLabel: "层级安全",
    layeredLabel: "分层",
    lazyImageAlt: "Aoi 渐变占位图",
    lazyMountMessage: "当文档演示进入视口时挂载。",
    like: "喜欢",
    linkChip: "链接",
    liveLabel: "实时",
    mediaHeavyMessage: "媒体重型组件。请结合 API 表和视口懒加载规则使用。",
    metadataChip: "元信息",
    moreActions: "更多动作",
    openDialog: "打开对话框",
    openGallery: "打开图库",
    openMenu: "打开菜单",
    outlinedDate: "描边日期",
    outlinedPalette: "描边色板",
    outlinedTime: "描边时间",
    palette: "色板",
    playPreview: "播放预览",
    propsLabel: "属性",
    publishTime: "发布时间",
    revealMessage: "Reveal 在水合前也保持内容可见。",
    revealStrength: "入场强度",
    save: "保存",
    saveLabel: "保存",
    scrollableRow: "可滚动行",
    searchDocs: "搜索文档",
    searchPlaceholder: "组件...",
    secondaryBlue: "辅助蓝",
    selectedChip: "已选",
    share: "分享",
    snapItems: ["介绍", "规则", "构建"],
    softDescription: "轻柔",
    softLabel: "柔和",
    sourceTag: "源码",
    slotsLabel: "插槽",
    stateLabel: "状态",
    statusMessage: "文档渲染器正在使用 Aoi 反馈组件。",
    switchLabel: "启用路由进度",
    sunflowerOrange: "向日葵橙",
    toolbarLabel: "工具栏",
    updatedDate: "更新日期",
    updatedLabel: "已更新",
    vividDescription: "明亮",
    vividLabel: "鲜明",
    wrappersLabel: "包装组件"
  },
  en: {
    accentColor: "Accent color",
    actionDemoLabel: "Docs action demo",
    apiLabel: "API",
    apiTabsLabel: "API tabs",
    balancedDescription: "Default",
    balancedLabel: "Balanced",
    cancel: "Cancel",
    captions: "Captions",
    cardLabel: "Card",
    chooseImage: "Choose image",
    clientHeavyMessage: "This component runs in the browser. The docs render its API and keep the live workbench client-only.",
    componentDocsTitle: "Component docs",
    confirm: "Confirm",
    copy: "Copy",
    createCodeLabel: "AoiButton example",
    densityLabel: "Density",
    disabledOption: "Disabled option",
    docsCompletionLabel: "Docs completion",
    docsLabel: "Docs",
    docsSectionDescription: "Section headers keep copy, icons, and actions aligned.",
    docsSectionTitle: "Docs section",
    docsTagsLabel: "Docs tags",
    demoLabel: "Demo",
    eventsLabel: "Events",
    fullscreen: "Fullscreen",
    genericFallback: "Aoi component demo",
    gridItems: ["One", "Two", "Three"],
    infoCardDescription: "Cards can carry media, copy, metadata, actions, and link behavior.",
    internalDocsLink: "Internal docs link",
    keepDocsVisible: "Keep docs visible",
    layerSafeLabel: "Layer safe",
    layeredLabel: "Layered",
    lazyImageAlt: "Aoi gradient placeholder",
    lazyMountMessage: "Mounted when the docs demo enters the viewport.",
    like: "Like",
    linkChip: "link",
    liveLabel: "Live",
    mediaHeavyMessage: "Media-heavy component. Use the API table with lazy viewport loading rules.",
    metadataChip: "metadata",
    moreActions: "More actions",
    openDialog: "Open dialog",
    openGallery: "Open gallery",
    openMenu: "Open menu",
    outlinedDate: "Outlined date",
    outlinedPalette: "Outlined palette",
    outlinedTime: "Outlined time",
    palette: "Palette",
    playPreview: "Play preview",
    propsLabel: "Props",
    publishTime: "Publish time",
    revealMessage: "Reveal keeps content visible before hydration.",
    revealStrength: "Reveal strength",
    save: "Save",
    saveLabel: "Save",
    scrollableRow: "Scrollable row",
    searchDocs: "Search docs",
    searchPlaceholder: "components...",
    secondaryBlue: "Secondary blue",
    selectedChip: "selected",
    share: "Share",
    snapItems: ["Intro", "Rules", "Build"],
    softDescription: "Calm",
    softLabel: "Soft",
    sourceTag: "source",
    slotsLabel: "Slots",
    stateLabel: "State",
    statusMessage: "Docs renderer is using Aoi feedback components.",
    switchLabel: "Enable route progress",
    sunflowerOrange: "Sunflower orange",
    toolbarLabel: "Toolbar",
    updatedDate: "Updated date",
    updatedLabel: "Updated",
    vividDescription: "Bright",
    vividLabel: "Vivid",
    wrappersLabel: "Wrappers"
  },
  ja: {
    accentColor: "アクセント色",
    actionDemoLabel: "ドキュメント操作デモ",
    apiLabel: "API",
    apiTabsLabel: "API タブ",
    balancedDescription: "標準",
    balancedLabel: "バランス",
    cancel: "キャンセル",
    captions: "字幕",
    cardLabel: "カード",
    chooseImage: "画像を選択",
    clientHeavyMessage: "このコンポーネントはブラウザーで動作します。ドキュメントでは API を描画し、ライブ作業領域はクライアント側に保ちます。",
    componentDocsTitle: "コンポーネント文書",
    confirm: "確認",
    copy: "コピー",
    createCodeLabel: "AoiButton の例",
    densityLabel: "密度",
    disabledOption: "無効な選択肢",
    docsCompletionLabel: "文書の進捗",
    docsLabel: "文書",
    docsSectionDescription: "セクション見出しは文言、アイコン、操作をそろえます。",
    docsSectionTitle: "文書セクション",
    docsTagsLabel: "文書タグ",
    demoLabel: "デモ",
    eventsLabel: "イベント",
    fullscreen: "全画面",
    genericFallback: "Aoi コンポーネントのデモ",
    gridItems: ["一", "二", "三"],
    infoCardDescription: "カードはメディア、文言、メタ情報、操作、リンク動作を持てます。",
    internalDocsLink: "内部ドキュメントリンク",
    keepDocsVisible: "文書を表示し続ける",
    layerSafeLabel: "レイヤー安全",
    layeredLabel: "階層化",
    lazyImageAlt: "Aoi グラデーションのプレースホルダー",
    lazyMountMessage: "文書デモが表示領域に入るとマウントされます。",
    like: "いいね",
    linkChip: "リンク",
    liveLabel: "ライブ",
    mediaHeavyMessage: "メディア重型コンポーネントです。API 表と表示領域の遅延読み込みルールを使います。",
    metadataChip: "メタ情報",
    moreActions: "その他の操作",
    openDialog: "ダイアログを開く",
    openGallery: "ギャラリーを開く",
    openMenu: "メニューを開く",
    outlinedDate: "アウトライン日付",
    outlinedPalette: "アウトライン色板",
    outlinedTime: "アウトライン時刻",
    palette: "色板",
    playPreview: "プレビュー再生",
    propsLabel: "プロパティ",
    publishTime: "公開時刻",
    revealMessage: "Reveal は hydrate 前も内容を表示したままにします。",
    revealStrength: "表示モーション強度",
    save: "保存",
    saveLabel: "保存",
    scrollableRow: "スクロール行",
    searchDocs: "文書検索",
    searchPlaceholder: "コンポーネント...",
    secondaryBlue: "補助ブルー",
    selectedChip: "選択済み",
    share: "共有",
    snapItems: ["紹介", "規則", "構築"],
    softDescription: "穏やか",
    softLabel: "ソフト",
    sourceTag: "ソース",
    slotsLabel: "スロット",
    stateLabel: "状態",
    statusMessage: "ドキュメント描画器は Aoi フィードバックコンポーネントを使っています。",
    switchLabel: "ルート進行を有効化",
    sunflowerOrange: "ひまわりオレンジ",
    toolbarLabel: "ツールバー",
    updatedDate: "更新日",
    updatedLabel: "更新済み",
    vividDescription: "明るい",
    vividLabel: "鮮明",
    wrappersLabel: "ラッパー"
  }
}

const docsLocale = useAoiDocsLocale()
const copy = computed(() => demoCopyByLocale[docsLocale.value])
const doc = computed(() => getLocalizedAoiComponentDocs(docsLocale.value).find((item) => item.name === props.name))
const selectedMode = ref("solid")
const selectedButtonBox = ref("docs")
const selectedSegment = ref("balanced")
const selectedTab = ref("props")
const selectedSelect = ref("orange")
const checked = ref(true)
const switched = ref(true)
const sliderValue = ref(64)
const colorInput = ref("#ff7d52")
const colorValue = ref<AoiRgbaColor>({ r: 255, g: 125, b: 82, a: 1 })
const textValue = ref("Aoi")
const dateValue = ref("2026-06-10")
const timeValue = ref("20:30")
const dialogOpen = ref(false)
const menuOpen = ref(false)
const lightboxOpen = ref(false)
const currentTime = ref(42)
const isPlaying = ref(false)
const muted = ref(false)
const volume = ref(72)
const rate = ref<PlayerPlaybackRate>(1)
const theater = ref(false)
const fullscreen = ref(false)
const danmakuText = ref("")

const normalizedName = computed(() => props.name)
const menuAnchorId = computed(() => `docs-menu-${props.name.toLowerCase()}`)
const buttonBoxItems = computed(() => [
  { value: "docs", label: copy.value.docsLabel, icon: "book-open" },
  { value: "demo", label: copy.value.demoLabel, icon: "sparkles" },
  { value: "api", label: copy.value.apiLabel, icon: "braces" }
])
const segmentedItems = computed(() => [
  { value: "soft", label: copy.value.softLabel, description: copy.value.softDescription, icon: "feather" },
  { value: "balanced", label: copy.value.balancedLabel, description: copy.value.balancedDescription, icon: "blocks" },
  { value: "vivid", label: copy.value.vividLabel, description: copy.value.vividDescription, icon: "sparkles" }
])
const tabItems = computed(() => [
  { value: "props", label: copy.value.propsLabel, icon: "list" },
  { value: "events", label: copy.value.eventsLabel, icon: "radio" },
  { value: "slots", label: copy.value.slotsLabel, icon: "panel-top" }
])
const selectOptions = computed(() => [
  { value: "orange", label: copy.value.sunflowerOrange },
  { value: "blue", label: copy.value.secondaryBlue },
  { value: "disabled", label: copy.value.disabledOption, disabled: true }
])
const statItems = computed(() => [
  { icon: "blocks", label: copy.value.wrappersLabel, value: "50+" },
  { icon: "shield-check", label: copy.value.layeredLabel, value: "Aoi" },
  { icon: "sparkles", label: copy.value.demoLabel, value: copy.value.liveLabel }
])
const tagItems = computed(() => [
  { icon: "hash", label: copy.value.docsLabel, value: "docs" },
  { icon: "book-open", label: copy.value.sourceTag, to: "/docs/project/repository" },
  { icon: "sparkles", label: "aoi", value: "aoi" }
])
const danmakuItems = computed<AoiDanmakuItem[]>(() => [
  {
    authorName: "Docs",
    body: "Aoi Danmaku",
    color: "#ffffff",
    createdAt: "2026-06-10T00:00:00.000Z",
    id: "docs-danmaku-1",
    mode: "scroll",
    timeSeconds: 4
  },
  {
    authorName: "Aoi",
    body: copy.value.layerSafeLabel,
    color: "#ffdf7e",
    createdAt: "2026-06-10T00:00:00.000Z",
    id: "docs-danmaku-2",
    mode: "top",
    timeSeconds: 8
  }
])
const lightboxItems = computed(() => [
  {
    alt: copy.value.lazyImageAlt,
    description: copy.value.infoCardDescription,
    id: "docs-gradient",
    src: "gradient:docs-lightbox",
    thumbnailSrc: "gradient:docs-lightbox-thumb",
    title: copy.value.componentDocsTitle,
    type: "image" as const
  }
])

function sendDanmaku(payload: { body: string, color: string, mode: AoiDanmakuMode }) {
  danmakuText.value = `${payload.body} (${payload.mode}, ${payload.color})`
}
</script>

<template>
  <AoiSurface class="docs-component-demo" surface="toolbar" padding="md">
    <div v-if="normalizedName === 'AoiButton'" class="docs-component-demo__row">
      <AoiButton icon="sparkles">Solid</AoiButton>
      <AoiButton appearance="soft" intent="info" icon="info">Soft</AoiButton>
      <AoiButton appearance="outline" intent="warning" icon="triangle-alert">Outline</AoiButton>
      <AoiButton appearance="plain" intent="danger" icon="circle-alert">Plain</AoiButton>
    </div>

    <div v-else-if="normalizedName === 'AoiActionBar'" class="docs-component-demo__row">
      <AoiActionBar :label="copy.actionDemoLabel" size="sm" surface>
        <AoiButton size="sm" icon="save">{{ copy.save }}</AoiButton>
        <AoiButton size="sm" appearance="outline" icon="copy">{{ copy.copy }}</AoiButton>
        <AoiIconButton icon="more-horizontal" :label="copy.moreActions" size="sm" />
      </AoiActionBar>
    </div>

    <div v-else-if="normalizedName === 'AoiButtonBox'" class="docs-component-demo__row">
      <AoiButtonBox v-model="selectedButtonBox" :items="buttonBoxItems" :aria-label="copy.docsLabel" />
    </div>

    <div v-else-if="normalizedName === 'AoiIconButton'" class="docs-component-demo__row">
      <AoiIconButton icon="heart" :label="copy.like" appearance="soft" active />
      <AoiIconButton icon="bookmark" :label="copy.saveLabel" appearance="outline" />
      <AoiIconButton icon="share-2" :label="copy.share" />
    </div>

    <div v-else-if="normalizedName === 'AoiLink'" class="docs-component-demo__row">
      <AoiLink to="/docs/project/overview">{{ copy.internalDocsLink }}</AoiLink>
      <AoiLink href="https://nuxt.com" target="_blank" format-url />
    </div>

    <div v-else-if="normalizedName === 'AoiMediaOverlayButton'" class="docs-component-demo__media-buttons">
      <AoiMediaOverlayButton icon="play" :label="copy.playPreview" :fill="false" />
      <AoiMediaOverlayButton icon="captions" :label="copy.captions" :fill="false" active />
      <AoiMediaOverlayButton icon="maximize" :label="copy.fullscreen" :fill="false" />
    </div>

    <div v-else-if="normalizedName === 'AoiCheckbox'" class="docs-component-demo__row">
      <AoiCheckbox v-model="checked" :label="copy.keepDocsVisible" />
      <AoiCheckbox disabled :label="copy.disabledOption" />
    </div>

    <div v-else-if="normalizedName === 'AoiSwitch'" class="docs-component-demo__row">
      <AoiSwitch v-model="switched" :label="copy.switchLabel" />
    </div>

    <div v-else-if="normalizedName === 'AoiTextField'" class="docs-component-demo__grid">
      <AoiTextField v-model="textValue" :label="copy.componentDocsTitle" icon="file-text" />
      <AoiTextField :label="copy.searchDocs" appearance="outlined" icon="search" :placeholder="copy.searchPlaceholder" />
    </div>

    <div v-else-if="normalizedName === 'AoiSelect'" class="docs-component-demo__grid">
      <AoiSelect v-model="selectedSelect" :label="copy.palette" :options="selectOptions" />
      <AoiSelect v-model="selectedSelect" :label="copy.outlinedPalette" appearance="outlined" :options="selectOptions" />
    </div>

    <div v-else-if="normalizedName === 'AoiSegmentedControl'" class="docs-component-demo__block">
      <AoiSegmentedControl v-model="selectedSegment" :items="segmentedItems" :columns="3" :aria-label="copy.densityLabel" />
    </div>

    <div v-else-if="normalizedName === 'AoiTabs'" class="docs-component-demo__block">
      <AoiTabs v-model="selectedTab" :items="tabItems" :aria-label="copy.apiTabsLabel" />
    </div>

    <div v-else-if="normalizedName === 'AoiSlider'" class="docs-component-demo__block">
      <AoiSlider v-model="sliderValue" :label="copy.revealStrength" :min="0" :max="100" />
    </div>

    <div v-else-if="normalizedName === 'AoiChoiceCard'" class="docs-component-demo__grid">
      <AoiChoiceCard value="balanced" :title="copy.balancedLabel" :description="copy.balancedDescription" icon="blocks" selected />
      <AoiChoiceCard value="vivid" :title="copy.vividLabel" :description="copy.vividDescription" icon="sparkles" />
    </div>

    <div v-else-if="normalizedName === 'AoiColorInput'" class="docs-component-demo__grid">
      <AoiColorInput v-model="colorInput" :label="copy.accentColor" />
      <span class="docs-component-demo__swatch" :style="{ background: colorInput }" />
    </div>

    <div v-else-if="normalizedName === 'AoiColorPalette'" class="docs-component-demo__block">
      <AoiColorPalette v-model="colorValue" :label="copy.palette" />
    </div>

    <div v-else-if="normalizedName === 'AoiDateField'" class="docs-component-demo__grid">
      <AoiDateField v-model="dateValue" :label="copy.updatedDate" />
      <AoiDateField v-model="dateValue" :label="copy.outlinedDate" appearance="outlined" />
    </div>

    <div v-else-if="normalizedName === 'AoiTimeField'" class="docs-component-demo__grid">
      <AoiTimeField v-model="timeValue" :label="copy.publishTime" />
      <AoiTimeField v-model="timeValue" :label="copy.outlinedTime" appearance="outlined" />
    </div>

    <div v-else-if="normalizedName === 'AoiFileInput'" class="docs-component-demo__row">
      <AoiFileInput accept="image/*">
        <template #default="{ open }">
          <AoiButton icon="image" appearance="outline" @click="open">{{ copy.chooseImage }}</AoiButton>
        </template>
      </AoiFileInput>
    </div>

    <ClientOnly v-else-if="doc?.demo === 'client-heavy'">
      <AoiStatusMessage intent="info" icon="monitor" :message="copy.clientHeavyMessage" />
    </ClientOnly>

    <div v-else-if="normalizedName === 'AoiChip'" class="docs-component-demo__row">
      <AoiChip :label="copy.selectedChip" icon="check" selected />
      <AoiChip :label="copy.metadataChip" icon="tag" />
      <AoiChip :label="copy.linkChip" icon="arrow-up-right" to="/docs" />
    </div>

    <div v-else-if="normalizedName === 'AoiCodeBlock'" class="docs-component-demo__block">
      <AoiCodeBlock code="<AoiButton icon=&quot;sparkles&quot;>Create</AoiButton>" :label="copy.createCodeLabel" />
    </div>

    <div v-else-if="normalizedName === 'AoiContentGrid'" class="docs-component-demo__block">
      <AoiContentGrid min-width="120px" gap="compact" :mobile-columns="2">
        <AoiSurface v-for="item in copy.gridItems" :key="item" padding="sm">{{ item }}</AoiSurface>
      </AoiContentGrid>
    </div>

    <div v-else-if="normalizedName === 'AoiIcon'" class="docs-component-demo__row">
      <AoiIcon name="sun" :size="28" decorative />
      <AoiIcon name="book-open" :size="28" decorative />
      <AoiIcon name="sparkles" :size="28" decorative />
    </div>

    <div v-else-if="normalizedName === 'AoiInfoCard'" class="docs-component-demo__block">
      <AoiInfoCard layout="inline" interactive to="/docs/components/actions">
        <template #media><AoiIcon name="book-open" :size="28" decorative /></template>
        <template #title>{{ copy.componentDocsTitle }}</template>
        <template #description>{{ copy.infoCardDescription }}</template>
        <template #meta><AoiMetaPill icon="check" label="Aoi" /></template>
      </AoiInfoCard>
    </div>

    <div v-else-if="normalizedName === 'AoiLazyImage'" class="docs-component-demo__image">
      <AoiLazyImage src="gradient:docs-lazy-image" :alt="copy.lazyImageAlt" />
    </div>

    <div v-else-if="normalizedName === 'AoiLazyMount'" class="docs-component-demo__block">
      <AoiLazyMount>
        <AoiStatusMessage intent="success" icon="eye" :message="copy.lazyMountMessage" />
      </AoiLazyMount>
    </div>

    <div v-else-if="normalizedName === 'AoiMetaPill'" class="docs-component-demo__row">
      <AoiMetaPill icon="clock" :label="copy.updatedLabel" value="2026-06-10" />
      <AoiMetaPill icon="layers" :label="copy.layerSafeLabel" />
    </div>

    <AoiReveal v-else-if="normalizedName === 'AoiReveal'" class="docs-component-demo__block" variant="rise">
      <AoiSurface padding="sm">{{ copy.revealMessage }}</AoiSurface>
    </AoiReveal>

    <div v-else-if="normalizedName === 'AoiScrollArea'" class="docs-component-demo__scroll">
      <AoiScrollArea :aria-label="copy.scrollableRow" axis="y" :tabindex="0">
        <p v-for="item in 6" :key="item">{{ copy.scrollableRow }} {{ item }}</p>
      </AoiScrollArea>
    </div>

    <div v-else-if="normalizedName === 'AoiScrollScene' || normalizedName === 'AoiScrollSnapItem'" class="docs-component-demo__snap">
      <AoiScrollSnapItem v-for="item in copy.snapItems" :key="item">
        <AoiSurface padding="sm">{{ item }}</AoiSurface>
      </AoiScrollSnapItem>
    </div>

    <div v-else-if="normalizedName === 'AoiSection'" class="docs-component-demo__block">
      <AoiSection :title="copy.docsSectionTitle" :description="copy.docsSectionDescription" icon="book-open" />
    </div>

    <div v-else-if="normalizedName.includes('Skeleton')" class="docs-component-demo__block">
      <AoiSkeletonGroup layout="stack" :count="2">
        <AoiSkeletonText :lines="2" />
      </AoiSkeletonGroup>
    </div>

    <div v-else-if="normalizedName === 'AoiStatGrid'" class="docs-component-demo__block">
      <AoiStatGrid :items="statItems" :columns="3" />
    </div>

    <div v-else-if="normalizedName === 'AoiSurface'" class="docs-component-demo__grid">
      <AoiSurface surface="card" padding="sm">{{ copy.cardLabel }}</AoiSurface>
      <AoiSurface surface="state" intent="success" padding="sm">{{ copy.stateLabel }}</AoiSurface>
      <AoiSurface surface="toolbar" padding="sm">{{ copy.toolbarLabel }}</AoiSurface>
    </div>

    <div v-else-if="normalizedName === 'AoiTagList'" class="docs-component-demo__block">
      <AoiTagList :items="tagItems" :label="copy.docsTagsLabel" />
    </div>

    <div v-else-if="normalizedName === 'AoiProgress'" class="docs-component-demo__grid">
      <AoiProgress indeterminate />
      <AoiProgress type="circular" indeterminate />
    </div>

    <div v-else-if="normalizedName === 'AoiProgressBar'" class="docs-component-demo__block">
      <AoiProgressBar :value="72" :label="copy.docsCompletionLabel" />
    </div>

    <div v-else-if="normalizedName === 'AoiStatusMessage'" class="docs-component-demo__block">
      <AoiStatusMessage intent="success" icon="circle-check" :message="copy.statusMessage" />
    </div>

    <div v-else-if="normalizedName === 'AoiDialog'" class="docs-component-demo__row">
      <AoiButton appearance="outline" icon="message-square" @click="dialogOpen = true">{{ copy.openDialog }}</AoiButton>
      <AoiDialog v-model:open="dialogOpen">
        <template #headline>{{ copy.openDialog }}</template>
        <p>{{ doc?.usage }}</p>
        <template #actions>
          <AoiButton appearance="plain" @click="dialogOpen = false">{{ copy.cancel }}</AoiButton>
          <AoiButton icon="check" @click="dialogOpen = false">{{ copy.confirm }}</AoiButton>
        </template>
      </AoiDialog>
    </div>

    <div v-else-if="normalizedName === 'AoiMenu'" class="docs-component-demo__row">
      <AoiButton :id="menuAnchorId" appearance="outline" icon="list" @click="menuOpen = !menuOpen">{{ copy.openMenu }}</AoiButton>
      <AoiMenu
        v-model:open="menuOpen"
        :anchor="menuAnchorId"
        :items="[
          { value: 'copy', label: copy.copy, icon: 'copy' },
          { value: 'inspect', label: copy.searchDocs, icon: 'search' }
        ]"
      />
    </div>

    <ClientOnly v-else-if="normalizedName === 'AoiLightboxGallery'">
      <div class="docs-component-demo__row">
        <AoiButton appearance="outline" icon="images" @click="lightboxOpen = true">{{ copy.openGallery }}</AoiButton>
        <AoiLightboxGallery v-model:open="lightboxOpen" :items="lightboxItems" />
      </div>
    </ClientOnly>

    <div v-else-if="normalizedName === 'AoiVideoControls' || normalizedName === 'AoiVideoToolbar' || normalizedName === 'AoiVideoTimeline'" class="docs-component-demo__block">
      <AoiVideoControls
        :current-time="currentTime"
        :duration="180"
        :is-playing="isPlaying"
        :muted="muted"
        :volume-percent="volume"
        :playback-rate="rate"
        :theater-mode="theater"
        :fullscreen="fullscreen"
        @seek="currentTime = $event"
        @toggle-play="isPlaying = !isPlaying"
        @toggle-muted="muted = !muted"
        @toggle-theater="theater = !theater"
        @toggle-fullscreen="fullscreen = !fullscreen"
        @update:volume-percent="volume = $event"
        @update:playback-rate="rate = $event"
      />
    </div>

    <div v-else-if="normalizedName === 'AoiVideoQueueList' || normalizedName === 'AoiWatchLayout' || doc?.demo === 'media-heavy'" class="docs-component-demo__media">
      <AoiIcon name="play-square" :size="30" decorative />
      <span>{{ copy.mediaHeavyMessage }}</span>
    </div>

    <div v-else-if="normalizedName === 'AoiDanmakuComposer'" class="docs-component-demo__block">
      <AoiDanmakuComposer @submit="sendDanmaku" />
      <AoiStatusMessage v-if="danmakuText" intent="success" :message="danmakuText" />
    </div>

    <div v-else-if="normalizedName === 'AoiDanmakuPanel'" class="docs-component-demo__block">
      <AoiDanmakuPanel :items="danmakuItems" :current-time="6" @seek="currentTime = $event" />
    </div>

    <div v-else-if="normalizedName === 'AoiDanmakuLayer'" class="docs-component-demo__danmaku">
      <AoiDanmakuLayer :items="danmakuItems" :current-time="6" playing />
    </div>

    <div v-else class="docs-component-demo__generic">
      <AoiIcon :name="doc ? 'blocks' : 'badge-help'" :size="22" decorative />
      <span>{{ doc?.description || copy.genericFallback }}</span>
    </div>
  </AoiSurface>
</template>

<style scoped>
.docs-component-demo {
  position: relative;
  isolation: isolate;
  overflow: hidden;
}

.docs-component-demo__row,
.docs-component-demo__grid,
.docs-component-demo__block,
.docs-component-demo__media-buttons,
.docs-component-demo__generic,
.docs-component-demo__media {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.docs-component-demo__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.docs-component-demo__block {
  display: grid;
  align-items: stretch;
}

.docs-component-demo__media-buttons {
  min-height: 120px;
  justify-content: center;
  border-radius: var(--aoi-radius-card);
  background:
    linear-gradient(180deg, rgba(0, 0, 0, .05), rgba(0, 0, 0, .58)),
    linear-gradient(135deg, var(--aoi-accent-20), var(--aoi-secondary-50));
}

.docs-component-demo__media,
.docs-component-demo__generic {
  border: 1px dashed var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  background: var(--aoi-surface);
  color: var(--aoi-text-muted);
  line-height: 1.6;
  padding: 14px;
}

.docs-component-demo__swatch {
  display: inline-block;
  width: 48px;
  height: 48px;
  border: 1px solid var(--aoi-border);
  border-radius: var(--aoi-radius-card);
  box-shadow: var(--aoi-shadow-sm);
}

.docs-component-demo__image {
  overflow: hidden;
  width: min(320px, 100%);
  aspect-ratio: 16 / 9;
  border-radius: var(--aoi-radius-card);
}

.docs-component-demo__scroll {
  max-height: 180px;
}

.docs-component-demo__scroll p {
  margin: 0;
  border-bottom: 1px solid var(--aoi-border);
  padding: 10px 0;
}

.docs-component-demo__snap {
  display: grid;
  grid-auto-columns: minmax(160px, 1fr);
  grid-auto-flow: column;
  gap: 10px;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
}

.docs-component-demo__danmaku {
  position: relative;
  min-height: 150px;
  overflow: hidden;
  border-radius: var(--aoi-radius-card);
  background: linear-gradient(135deg, var(--aoi-accent-40), var(--aoi-secondary-50));
}
</style>


