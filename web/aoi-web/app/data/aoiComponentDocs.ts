import type {
  AoiComponentDoc,
  AoiComponentDocApiRow,
  AoiComponentDocCategory,
  AoiComponentDocCategoryMeta,
  AoiDocsLocale
} from "~/types/docs"

function row(name: string, type: string, description: string, defaultValue?: string): AoiComponentDocApiRow {
  return { name, type, description, defaultValue }
}

function component(
  name: string,
  category: AoiComponentDocCategory,
  description: string,
  usage: string,
  props: AoiComponentDocApiRow[],
  events: AoiComponentDocApiRow[] = [],
  slots: AoiComponentDocApiRow[] = [row("default", "slot", "Main content slot.")],
  notes: string[] = [],
  source = `app/components/aoi/${name}.vue`,
  demo = name
): AoiComponentDoc {
  return {
    category,
    demo,
    description,
    events,
    name,
    notes,
    props,
    slots,
    source,
    usage
  }
}

export const aoiComponentCategories: AoiComponentDocCategoryMeta[] = [
  {
    id: "actions",
    title: "Actions",
    description: "Buttons, links, action bars, and command-style navigation.",
    icon: "mouse-pointer-click"
  },
  {
    id: "forms",
    title: "Forms",
    description: "Input, selection, upload, color, date, image, and editor controls.",
    icon: "sliders-horizontal"
  },
  {
    id: "layout-content",
    title: "Layout And Content",
    description: "Surface, grid, section, metadata, lazy loading, skeleton, and scroll primitives.",
    icon: "layout-template"
  },
  {
    id: "feedback",
    title: "Feedback",
    description: "Status, progress, loading, and state feedback components.",
    icon: "badge-info"
  },
  {
    id: "overlays",
    title: "Overlays",
    description: "Dialog, menu, gallery, and context-menu layer wrappers.",
    icon: "panel-top-open"
  },
  {
    id: "media-player",
    title: "Media Player",
    description: "Video player layout, controls, timeline, toolbar, and queue pieces.",
    icon: "play-square"
  },
  {
    id: "danmaku-motion-rich-text",
    title: "Danmaku, Motion, Rich Text",
    description: "Danmaku composition, motion staging, and Markdown-rich editing surfaces.",
    icon: "sparkles"
  }
]

export const aoiComponentDocs: AoiComponentDoc[] = [
  component("AoiActionBar", "actions", "Groups related actions in a responsive toolbar surface.", "Use for page-level or panel-level command rows.", [
    row("label", "string", "Accessible label for the action group."),
    row("size", "\"sm\" | \"md\"", "Button density inside the action row.", "\"md\""),
    row("surface", "boolean", "Adds the toolbar surface treatment.", "false")
  ], [], [row("default", "slot", "AoiButton or AoiIconButton actions.")]),
  component("AoiButton", "actions", "Material Web button wrapper with Aoi appearance, intent, icon, loading, and link support.", "Use for all button-like commands and button-style navigation.", [
    row("appearance", "\"solid\" | \"soft\" | \"outline\" | \"plain\" | \"elevated\"", "Visual form of the action.", "\"solid\""),
    row("intent", "\"primary\" | \"secondary\" | \"neutral\" | \"success\" | \"warning\" | \"danger\" | \"info\"", "Semantic meaning and color mapping.", "\"primary\""),
    row("size", "\"sm\" | \"md\" | \"lg\"", "Control height scale.", "\"md\""),
    row("icon / trailingIcon", "string", "Lucide icon names rendered through AoiIcon."),
    row("to / href", "RouteLocationRaw", "Delegates navigation to AoiLink."),
    row("loading / disabled", "boolean", "Disables interaction and optionally shows spinner.")
  ], [row("click", "MouseEvent", "Emitted when the command is activated.")]),
  component("AoiButtonBox", "actions", "Selectable button-group wrapper for single or multiple values.", "Use when a compact set of choices should look like action buttons.", [
    row("modelValue", "string | string[]", "Selected value or selected values."),
    row("items", "AoiButtonBoxItem[]", "Value, label, icon, and disabled entries."),
    row("multiple", "boolean", "Allows more than one selected value.", "false"),
    row("ariaLabel", "string", "Accessible group label.")
  ], [
    row("update:modelValue", "string | string[]", "Emitted when selection changes."),
    row("change", "string | string[]", "Selection change notification.")
  ]),
  component("AoiIconButton", "actions", "Icon-only command or navigation button with Aoi intent and appearance.", "Use when the icon is the primary affordance and a label remains available to assistive tech.", [
    row("icon", "string", "Lucide icon name."),
    row("label", "string", "Required accessible label."),
    row("appearance", "\"standard\" | \"filled\" | \"tonal\" | \"outlined\"", "Material icon-button form.", "\"standard\""),
    row("size", "\"sm\" | \"md\" | \"lg\"", "Icon button dimensions.", "\"md\""),
    row("active", "boolean", "Applies active visual state.", "false"),
    row("to / href", "RouteLocationRaw", "Delegates navigation to AoiLink.")
  ], [row("click", "MouseEvent", "Emitted on activation.")]),
  component("AoiLink", "actions", "NuxtLink facade that centralizes internal, external, and formatted URL behavior.", "Use for all text, card, tag, and navigation links.", [
    row("to / href", "RouteLocationRaw", "Internal or external navigation target."),
    row("custom", "boolean", "Exposes NuxtLink slot props without rendering an anchor.", "false"),
    row("formatUrl", "boolean", "Splits absolute URLs into protocol, host, and suffix spans.", "false"),
    row("target / rel / noRel", "string | boolean", "External link relationship controls.")
  ], [], [row("default", "slot | slot props", "Link label, or custom slot props when custom is true.")]),
  component("AoiMediaOverlayButton", "actions", "Compact icon button styled for media overlays.", "Use inside video, image, and preview surfaces.", [
    row("icon", "string", "Lucide icon name."),
    row("label", "string", "Accessible label."),
    row("active", "boolean", "Selected/toggled visual state.", "false"),
    row("disabled", "boolean", "Disables the overlay command.", "false")
  ], [row("click", "MouseEvent", "Emitted on activation.")], []),
  component("AoiCheckbox", "forms", "Material checkbox wrapper with Aoi model binding.", "Use for binary settings and checklist rows.", [
    row("modelValue", "boolean", "Checked state."),
    row("disabled", "boolean", "Disables interaction.", "false"),
    row("label", "string", "Optional visible label.")
  ], [row("update:modelValue", "boolean", "Emitted when checked state changes.")]),
  component("AoiChoiceCard", "forms", "Clickable card option for richer radio-like selection.", "Use for settings choices that need title, description, icon, and selected state.", [
    row("value", "string", "Choice identifier."),
    row("title", "string", "Primary choice label."),
    row("description", "string", "Supporting copy."),
    row("icon", "string", "Optional Lucide icon."),
    row("selected", "boolean", "Applies selected styling.", "false"),
    row("disabled", "boolean", "Disables the choice.", "false")
  ], [row("select", "string", "Emitted with the selected value.")]),
  component("AoiColorInput", "forms", "Hex color text input with Aoi field styling.", "Use for direct color entry paired with palette controls.", [
    row("modelValue", "string", "Hex color value."),
    row("label", "string", "Field label."),
    row("disabled", "boolean", "Disables editing.", "false")
  ], [row("update:modelValue", "string", "Emitted when the color string changes.")]),
  component("AoiColorPalette", "forms", "RGBA color palette with hue, channel, alpha, and mode controls.", "Use for advanced color customization.", [
    row("modelValue", "AoiRgbaColor", "RGBA object."),
    row("label", "string", "Palette label."),
    row("disabled", "boolean", "Disables all controls.", "false")
  ], [row("update:modelValue", "AoiRgbaColor", "Emitted when color channels change.")]),
  component("AoiDateField", "forms", "Date input wrapper with filled/outlined Material Web variants.", "Use for date-only settings and metadata.", [
    row("modelValue", "string", "ISO-like date value."),
    row("label", "string", "Field label."),
    row("appearance", "\"filled\" | \"outlined\"", "Field visual form.", "\"filled\""),
    row("min / max", "string", "Native date constraints."),
    row("disabled", "boolean", "Disables editing.", "false")
  ], [row("update:modelValue", "string", "Emitted on date changes.")]),
  component("AoiFileInput", "forms", "Accessible file picker facade.", "Use where a native file input is needed behind an Aoi command surface.", [
    row("accept", "string", "Accepted file types."),
    row("multiple", "boolean", "Allows multiple files.", "false"),
    row("disabled", "boolean", "Disables selection.", "false")
  ], [row("change", "File[]", "Emitted with selected files.")], [row("default", "{ open }", "Slot receives an open function for a custom trigger.")]),
  component("AoiImageClipboard", "forms", "Client-side image clipboard and paste helper.", "Use for image workflows that accept paste, drag, and local image metadata.", [
    row("label", "string", "Workbench label."),
    row("disabled", "boolean", "Disables paste/drop handling.", "false")
  ], [row("change", "payload", "Emitted when image clipboard content changes.")], [row("default", "slot", "Optional workbench content.")], ["Client-only behavior; do not use during SSR."], "app/components/aoi/AoiImageClipboard.vue", "client-heavy"),
  component("AoiImageCropperWorkbench", "forms", "Client-only cropper workbench backed by cropperjs.", "Use for local image crop and preview workflows.", [
    row("src", "string", "Image source."),
    row("aspectRatio", "number", "Crop aspect ratio."),
    row("disabled", "boolean", "Disables crop interaction.", "false")
  ], [row("crop", "payload", "Emitted with crop results.")], [row("actions", "slot", "Optional cropper actions.")], ["Client-only and media-heavy; wrap in ClientOnly."], "app/components/aoi/AoiImageCropperWorkbench.client.vue", "client-heavy"),
  component("AoiRichTextEditor", "danmaku-motion-rich-text", "Client-only Tiptap Markdown editor with document JSON sync.", "Use for rich authoring where Markdown is the exchange format.", [
    row("modelValue", "string", "Markdown content."),
    row("document", "AoiRichTextDocument | null", "Optional Tiptap JSON document."),
    row("label / placeholder / supportingText", "string", "Field copy."),
    row("maxLength", "number", "Optional character limit."),
    row("toolbar", "\"document\" | \"none\"", "Toolbar mode.", "\"document\"")
  ], [
    row("update:modelValue", "string", "Markdown output."),
    row("update:document", "AoiRichTextDocument", "Tiptap JSON output."),
    row("change", "AoiRichTextChangePayload", "Combined Markdown, JSON, and text payload.")
  ], [], ["Client-only; keep unsafe links and images validated by the component."], "app/components/aoi/AoiRichTextEditor.client.vue", "client-heavy"),
  component("AoiSegmentedControl", "forms", "Button or tab-like segmented choice control.", "Use for mutually exclusive modes with compact labels and icons.", [
    row("modelValue", "string", "Selected value."),
    row("items", "AoiSegmentedItem[]", "Value, label, description, icon, accent, disabled."),
    row("columns", "2 | 3 | \"auto\"", "Grid column behavior.", "\"auto\""),
    row("selectionRole", "\"button\" | \"tab\"", "ARIA selection pattern.", "\"button\"")
  ], [
    row("update:modelValue", "string", "Emitted when selected value changes."),
    row("change", "string", "Selection change notification.")
  ]),
  component("AoiSelect", "forms", "Material select wrapper with Aoi field appearance and layer positioning.", "Use for option sets that need a menu instead of segmented buttons.", [
    row("modelValue", "string", "Selected value."),
    row("options", "AoiSelectOption[]", "Value, label, and disabled entries."),
    row("label", "string", "Field label."),
    row("appearance", "\"filled\" | \"outlined\"", "Field visual form.", "\"filled\""),
    row("menuPositioning", "\"absolute\" | \"fixed\" | \"popover\"", "Overlay strategy.", "\"popover\"")
  ], [row("update:modelValue", "string", "Emitted when selection changes.")]),
  component("AoiSlider", "forms", "Material slider wrapper with Aoi label, compact, and contrast modes.", "Use for numeric settings.", [
    row("modelValue", "number", "Slider value."),
    row("min / max / step", "number", "Numeric bounds and step.", "0 / 100 / 1"),
    row("label / ariaLabel", "string", "Visible and accessible labels."),
    row("contrast", "\"default\" | \"inverse\"", "Color contrast mode.", "\"default\""),
    row("compact", "boolean", "Reduces field spacing.", "false")
  ], [
    row("update:modelValue", "number", "Emitted during input."),
    row("change", "number", "Emitted on committed change.")
  ]),
  component("AoiSwitch", "forms", "Material switch wrapper with v-model.", "Use for binary settings where on/off state is the object.", [
    row("modelValue", "boolean", "Switch state."),
    row("disabled", "boolean", "Disables interaction.", "false"),
    row("label", "string", "Optional visible label.")
  ], [row("update:modelValue", "boolean", "Emitted when switch state changes.")]),
  component("AoiTextField", "forms", "Material text field wrapper with icons, counters, and filled/outlined variants.", "Use for text, password, search, and textarea-like fields.", [
    row("modelValue", "string", "Field value."),
    row("label / placeholder / supportingText", "string", "Field copy."),
    row("appearance", "\"filled\" | \"outlined\"", "Field visual form.", "\"filled\""),
    row("type", "string", "Native input type."),
    row("icon / trailingIcon", "string", "Lucide icon names."),
    row("maxLength", "number", "Optional counter limit.")
  ], [
    row("update:modelValue", "string", "Emitted on value changes."),
    row("change", "string", "Committed change notification.")
  ]),
  component("AoiTimeField", "forms", "Time input wrapper with filled/outlined Material variants.", "Use for time-only settings.", [
    row("modelValue", "string", "Time value."),
    row("label", "string", "Field label."),
    row("appearance", "\"filled\" | \"outlined\"", "Field visual form.", "\"filled\""),
    row("disabled", "boolean", "Disables editing.", "false")
  ], [row("update:modelValue", "string", "Emitted on time changes.")]),
  component("AoiChip", "layout-content", "Small tag, filter, or status chip with optional icon and selection.", "Use for compact metadata and selectable labels.", [
    row("label", "string", "Chip label."),
    row("icon", "string", "Optional Lucide icon."),
    row("selected", "boolean", "Applies selected styling.", "false"),
    row("disabled", "boolean", "Disables interaction.", "false"),
    row("to / href", "RouteLocationRaw", "Optional link target.")
  ], [row("click", "MouseEvent", "Emitted when clicked.")], []),
  component("AoiCodeBlock", "layout-content", "Scrollable code block with Aoi tokens and native scroll directive.", "Use for source snippets and JSON previews.", [
    row("code", "string", "Code text."),
    row("fallback", "string", "Text shown when code is empty."),
    row("label", "string", "Accessible label."),
    row("reveal", "AoiRevealProp", "Optional reveal motion.", "false"),
    row("scrollNative", "boolean", "Uses native scroll directive.", "true")
  ], [], []),
  component("AoiContentGrid", "layout-content", "Responsive content grid wrapper.", "Use for repeated cards, chips, and docs panels.", [
    row("minWidth", "string", "Grid item minimum width."),
    row("gap", "\"normal\" | \"compact\" | \"video\"", "Grid spacing preset.", "\"normal\""),
    row("mobileColumns", "1 | 2", "Mobile column behavior.")
  ]),
  component("AoiIcon", "layout-content", "Local Lucide icon renderer through @nuxt/icon.", "Use instead of inline SVGs for product icons.", [
    row("name", "string", "Lucide icon name."),
    row("size", "number | string", "Icon size."),
    row("decorative", "boolean", "Hides icon from assistive tech.", "false"),
    row("label", "string", "Accessible label when not decorative.")
  ], [], []),
  component("AoiInfoCard", "layout-content", "Reusable card shell with optional media, metadata, actions, and link behavior.", "Use for list cards that need consistent Aoi surfaces.", [
    row("as", "string", "Rendered root tag.", "\"article\""),
    row("layout", "\"inline\" | \"stack\"", "Media/copy layout.", "\"stack\""),
    row("density", "\"default\" | \"compact\"", "Spacing density.", "\"default\""),
    row("to / href", "RouteLocationRaw", "Optional card link target."),
    row("interactive / selected", "boolean", "Interaction and selected state.")
  ], [], [
    row("media", "slot", "Media or icon area."),
    row("title / subtitle / description / meta", "slot", "Structured copy slots."),
    row("actions", "slot", "Trailing actions.")
  ]),
  component("AoiLazyImage", "layout-content", "Lazy image wrapper with gradient placeholder support.", "Use for media cards and thumbnails that should not eagerly bind real image URLs.", [
    row("src", "string", "Image source."),
    row("alt", "string", "Image alt text."),
    row("placeholder", "string", "Optional placeholder."),
    row("ratio", "string", "Aspect ratio.")
  ], [row("load / error", "Event", "Image lifecycle notifications.")]),
  component("AoiLazyMount", "layout-content", "Viewport-aware lazy mounting boundary.", "Use for heavy client components and media previews.", [
    row("rootMargin", "string", "IntersectionObserver root margin."),
    row("fallbackVisible", "boolean", "Renders content when observer is unavailable.", "true")
  ], [row("visible", "void", "Emitted when content becomes visible.")]),
  component("AoiMetaPill", "layout-content", "Metadata pill for compact facts and counts.", "Use inside cards, headers, and media metadata.", [
    row("icon", "string", "Optional Lucide icon."),
    row("label", "string", "Pill label."),
    row("value", "string | number", "Optional value.")
  ], [], []),
  component("AoiReveal", "layout-content", "Wrapper component for reusable viewport reveal motion.", "Use when an existing component already has hover transforms.", [
    row("as", "string", "Rendered tag.", "\"div\""),
    row("variant", "AoiRevealVariant", "Reveal effect."),
    row("index / stagger", "number", "Delay calculation."),
    row("disabled", "boolean", "Forces immediate visibility.")
  ]),
  component("AoiScrollArea", "layout-content", "Scrollable area wrapper that cooperates with Aoi scroll directives.", "Use for contained long lists and panels.", [
    row("ariaLabel", "string", "Accessible region label."),
    row("axis", "\"x\" | \"y\"", "Scroll axis.", "\"x\""),
    row("overscroll", "AoiScrollOverscroll", "Overscroll behavior.", "\"contain\""),
    row("rubberBand", "boolean", "Enables Aoi edge resistance.", "true"),
    row("snap", "boolean", "Enables configured scroll snapping.", "false")
  ]),
  component("AoiScrollScene", "layout-content", "Horizontal or staged scroll interaction container.", "Use for rich scroll demos and carefully bounded storytelling sections.", [
    row("items", "unknown[]", "Scene item data."),
    row("enabled", "boolean", "Toggles scene behavior.", "true"),
    row("snap", "boolean", "Enables snap behavior.")
  ], [row("change", "number", "Active scene index.")]),
  component("AoiScrollSnapItem", "layout-content", "Single item wrapper for scroll-snap scenes.", "Use inside AoiScrollScene or native snap lists.", [
    row("as", "string", "Rendered tag.", "\"section\""),
    row("active", "boolean", "Active item state.")
  ]),
  component("AoiSection", "layout-content", "Section primitive with icon, eyebrow, heading, description, actions, and layout modes.", "Use as the default wrapper for page sections and docs regions.", [
    row("title / description / eyebrow", "string", "Header copy."),
    row("icon", "string", "Optional Lucide icon."),
    row("layout", "\"stack\" | \"grid\" | \"inline\" | \"split\"", "Section layout.", "\"stack\""),
    row("level", "2 | 3 | 4", "Heading level.", "2"),
    row("reveal", "AoiRevealProp", "Optional reveal motion.", "\"rise\"")
  ], [], [row("title / actions / default", "slot", "Custom heading, actions, and section body.")]),
  component("AoiSkeleton", "layout-content", "Primitive skeleton block.", "Use as the base loading placeholder.", [
    row("shape", "string", "Skeleton shape preset."),
    row("width / height", "string | number", "Explicit dimensions."),
    row("animated", "boolean", "Enables shimmer.", "true")
  ], [], []),
  component("AoiSkeletonGroup", "layout-content", "Skeleton layout group for rows, grids, and stacks.", "Use to compose predictable loading states.", [
    row("layout", "\"stack\" | \"row\" | \"grid\" | \"inline\" | \"custom\"", "Group layout.", "\"stack\""),
    row("count", "number", "Generated skeleton count."),
    row("gap", "string | number", "Group gap.")
  ]),
  component("AoiSkeletonText", "layout-content", "Multi-line text skeleton.", "Use for loading paragraphs and card metadata.", [
    row("lines", "number", "Number of text lines.", "3"),
    row("animated", "boolean", "Enables shimmer.", "true")
  ], [], []),
  component("AoiStatGrid", "layout-content", "Responsive grid of icon, label, and value stats.", "Use for compact dashboard-like summaries.", [
    row("items", "AoiStatItem[]", "Stat entries."),
    row("columns", "number", "Preferred column count.")
  ], [], []),
  component("AoiSurface", "layout-content", "Core surface primitive for card, panel, toolbar, code, and state containers.", "Use before inventing local card styles.", [
    row("as", "string", "Rendered tag.", "\"div\""),
    row("surface", "\"plain\" | \"panel\" | \"card\" | \"state\" | \"code\" | \"toolbar\"", "Surface kind.", "\"card\""),
    row("padding", "\"none\" | \"sm\" | \"md\" | \"lg\"", "Padding preset.", "\"md\""),
    row("intent", "AoiIntent", "Semantic tint.", "\"neutral\""),
    row("interactive / selected", "boolean", "Interactive and selected states.")
  ]),
  component("AoiTagList", "layout-content", "Wraps metadata tags using AoiChip/AoiLink semantics.", "Use for tag lists on videos and cards.", [
    row("items", "AoiTagItem[]", "Tag entries with label, icon, value, to, or href."),
    row("label", "string", "Accessible list label.")
  ], [row("select", "string", "Emitted when a non-link tag is selected.")], []),
  component("AoiProgress", "feedback", "Material linear/circular progress wrapper.", "Use for indeterminate or determinate loading indicators.", [
    row("type", "\"linear\" | \"circular\"", "Progress style.", "\"linear\""),
    row("value", "number", "Determinate progress value."),
    row("indeterminate", "boolean", "Shows ongoing work.", "false")
  ], [], []),
  component("AoiProgressBar", "feedback", "Lightweight CSS progress bar with intent colors.", "Use for compact status meters.", [
    row("value", "number", "Current value."),
    row("max", "number", "Maximum value.", "100"),
    row("intent", "AoiIntent", "Semantic fill color.", "\"primary\""),
    row("size", "\"sm\" | \"md\"", "Bar height.", "\"sm\""),
    row("label", "string", "Accessible label.")
  ], [], []),
  component("AoiStatusMessage", "feedback", "Inline semantic status block.", "Use for success, warning, danger, and info copy.", [
    row("intent", "\"danger\" | \"info\" | \"success\" | \"warning\"", "Semantic message style.", "\"info\""),
    row("icon", "string", "Optional Lucide icon."),
    row("message", "string", "Fallback message text."),
    row("as", "string", "Rendered tag.", "\"p\"")
  ]),
  component("AoiDialog", "overlays", "Material dialog wrapper registered in the Aoi layer stack.", "Use for modal confirmation and focused workflows.", [
    row("open", "boolean", "Dialog open state.", "false"),
    row("dismissible", "boolean", "Allows cancel/escape dismissal.", "true")
  ], [
    row("update:open", "boolean", "Emitted when dialog closes."),
    row("cancel", "Event", "Cancel event; can be prevented when not dismissible."),
    row("closed", "void", "Dialog has closed.")
  ], [
    row("headline", "slot", "Dialog title."),
    row("default", "slot", "Dialog content."),
    row("actions", "slot", "Dialog actions.")
  ]),
  component("AoiLightboxGallery", "overlays", "Client overlay gallery for image and video previews.", "Use for media lightboxes with keyboard and touch navigation.", [
    row("open", "boolean", "Gallery open state."),
    row("activeIndex", "number", "Current media index."),
    row("items", "AoiLightboxItem[]", "Image/video media entries."),
    row("loop", "boolean", "Wraps navigation.", "false")
  ], [
    row("update:open", "boolean", "Open state changes."),
    row("update:activeIndex", "number", "Active media changes.")
  ], [], ["Client-only overlay; keep media sources lazy where practical."]),
  component("AoiMenu", "overlays", "Material menu wrapper with Aoi layer positioning.", "Use for anchored command lists.", [
    row("open", "boolean", "Menu open state."),
    row("items", "AoiMenuItem[]", "Menu entries."),
    row("anchor", "string", "Anchor element id."),
    row("positioning", "\"absolute\" | \"fixed\" | \"document\" | \"popover\"", "Overlay positioning.", "\"popover\"")
  ], [
    row("update:open", "boolean", "Open state changes."),
    row("select", "string", "Selected item value.")
  ]),
  component("AoiPlayerContextMenu", "overlays", "Context menu specialized for player actions.", "Use inside player surfaces for grouped playback commands.", [
    row("open", "boolean", "Context menu open state."),
    row("items", "AoiPlayerContextMenuItem[]", "Grouped menu entries."),
    row("x / y", "number", "Menu coordinates.")
  ], [
    row("update:open", "boolean", "Open state changes."),
    row("select", "string", "Selected action value.")
  ]),
  component("AoiVideoControls", "media-player", "Composite player controls row with timeline, playback, volume, rate, theater, and fullscreen actions.", "Use inside Aoi video player shells.", [
    row("currentTime / duration", "number", "Playback time values."),
    row("isPlaying / muted / theaterMode / fullscreen", "boolean", "Playback state flags."),
    row("volumePercent", "number", "Volume percentage."),
    row("playbackRate", "PlayerPlaybackRate", "Current playback rate.")
  ], [
    row("seek", "number", "Timeline seek target."),
    row("toggle-play / toggle-muted / toggle-theater / toggle-fullscreen", "void", "Player toggle commands."),
    row("update:volume-percent / update:playback-rate", "number", "Control changes.")
  ], []),
  component("AoiVideoPlayer", "media-player", "Full video player wrapper with HLS/DASH source handling and optional danmaku integration.", "Use for watch experiences after sources are in viewport.", [
    row("title", "string", "Accessible player title."),
    row("sources", "VideoSource[]", "Playable media sources."),
    row("poster", "string", "Poster image."),
    row("surface", "\"solid\" | \"translucent\"", "Player chrome surface.", "\"solid\""),
    row("preload", "string", "Video preload behavior.")
  ], [
    row("play / pause / error", "Event", "Playback lifecycle events."),
    row("update:*", "payload", "Playback state sync events.")
  ], [row("default", "slot", "Optional surrounding player content.")], ["Media sources should stay lazy and viewport-aware."], "app/components/aoi/AoiVideoPlayer.vue", "media-heavy"),
  component("AoiVideoQueueList", "media-player", "Queue and related-video list component.", "Use beside watch surfaces.", [
    row("items", "VideoSummary[]", "Queue entries."),
    row("activeId", "string", "Current video id."),
    row("label", "string", "Accessible list label.")
  ], [row("select", "VideoSummary", "Selected queue item.")]),
  component("AoiVideoTimeline", "media-player", "Player timeline slider with seek behavior.", "Use when composing custom player controls.", [
    row("currentTime / duration", "number", "Timeline values."),
    row("disabled", "boolean", "Disables seeking.", "false"),
    row("label", "string", "Accessible label.")
  ], [row("seek", "number", "Requested seek time.")], []),
  component("AoiVideoToolbar", "media-player", "Playback toolbar with play, mute, rate, theater, and fullscreen controls.", "Use with AoiVideoTimeline or custom players.", [
    row("isPlaying / muted / theaterMode / fullscreen", "boolean", "Control state flags."),
    row("volumePercent", "number", "Volume percentage."),
    row("playbackRate", "PlayerPlaybackRate", "Current rate.")
  ], [
    row("toggle-play / toggle-muted / toggle-theater / toggle-fullscreen", "void", "Toolbar commands."),
    row("update:volume-percent / update:playback-rate", "number", "Control changes.")
  ], []),
  component("AoiWatchLayout", "media-player", "Responsive watch-page layout primitive.", "Use to align player, details, comments, and related content.", [
    row("theater", "boolean", "Expands player layout.", "false")
  ], [], [
    row("player", "slot", "Player surface."),
    row("details", "slot", "Video metadata."),
    row("aside", "slot", "Related content."),
    row("default", "slot", "Main below-player content.")
  ]),
  component("AoiDanmakuComposer", "danmaku-motion-rich-text", "Input composer for timed danmaku messages.", "Use beside danmaku-enabled player surfaces.", [
    row("count", "number", "Visible danmaku count.", "0"),
    row("compact / overlay", "boolean", "Density and overlay presentation modes.", "false"),
    row("disabled / enabled / playing", "boolean", "Composer and playback state flags.")
  ], [
    row("submit", "{ body, color, mode }", "Emitted when a danmaku is submitted."),
    row("toggle-enabled", "void", "Requests danmaku visibility toggle.")
  ]),
  component("AoiDanmakuLayer", "danmaku-motion-rich-text", "Visual danmaku overlay layer.", "Use over media surfaces for moving and fixed comments.", [
    row("items", "AoiDanmakuItem[]", "Danmaku entries."),
    row("currentTime", "number", "Playback time."),
    row("playing", "boolean", "Animation playback state.", "true"),
    row("settings", "Partial<AoiDanmakuRuntimeSettings>", "Runtime visual settings.")
  ], [], []),
  component("AoiDanmakuPanel", "danmaku-motion-rich-text", "Side panel/list for danmaku entries.", "Use for browsing and seeking timed comments.", [
    row("items", "AoiDanmakuItem[]", "Danmaku entries."),
    row("currentTime", "number", "Playback time."),
    row("emptyText", "string", "Empty state text.")
  ], [row("seek", "number", "Requested time from a danmaku item.")]),
  component("AoiDanmakuVideoPlayer", "danmaku-motion-rich-text", "Video player composition with mapped danmaku support.", "Use when backend or mock danmaku items need mapping into the player layer.", [
    row("items", "TItem[]", "Raw danmaku entries."),
    row("mapper", "AoiDanmakuMapper<TItem>", "Maps raw entries to layer items."),
    row("sources", "VideoSource[]", "Playable video sources."),
    row("enabled", "boolean", "Danmaku visibility.", "true")
  ], [
    row("send-danmaku", "payload", "New danmaku payload."),
    row("error", "payload", "Player or source error.")
  ], [row("controls / panel", "slot", "Optional control and panel customization.")], ["Use ClientOnly when media APIs are required."], "app/components/aoi/AoiDanmakuVideoPlayer.vue", "media-heavy")
]

export function getAoiComponentDocs(category?: AoiComponentDocCategory) {
  return category ? aoiComponentDocs.filter((item) => item.category === category) : aoiComponentDocs
}

export function getAoiComponentCategoryMeta(category: AoiComponentDocCategory) {
  return aoiComponentCategories.find((item) => item.id === category)
}

type AoiComponentDocApiKind = "events" | "props" | "slots"
type LocalizedCategoryCopy = Pick<AoiComponentDocCategoryMeta, "description" | "title">

const categoryCopyByLocale: Record<AoiDocsLocale, Record<AoiComponentDocCategory, LocalizedCategoryCopy>> = {
  "zh-CN": {
    actions: {
      title: "动作组件",
      description: "按钮、链接、动作条和命令式导航控件。"
    },
    forms: {
      title: "表单组件",
      description: "输入、选择、上传、颜色、日期、图片和编辑器控件。"
    },
    "layout-content": {
      title: "布局与内容",
      description: "Surface、网格、Section、元信息、懒加载、骨架屏和滚动基础组件。"
    },
    feedback: {
      title: "反馈组件",
      description: "状态、进度、加载和页面反馈组件。"
    },
    overlays: {
      title: "浮层组件",
      description: "Dialog、Menu、Gallery 和上下文菜单浮层 wrapper。"
    },
    "media-player": {
      title: "媒体播放器",
      description: "视频播放器布局、控制条、时间轴、工具栏和队列组件。"
    },
    "danmaku-motion-rich-text": {
      title: "弹幕、动效与富文本",
      description: "弹幕编写、动效编排和 Markdown 富文本编辑界面。"
    }
  },
  en: {
    actions: {
      title: "Actions",
      description: "Buttons, links, action bars, and command-style navigation."
    },
    forms: {
      title: "Forms",
      description: "Input, selection, upload, color, date, image, and editor controls."
    },
    "layout-content": {
      title: "Layout And Content",
      description: "Surface, grid, section, metadata, lazy loading, skeleton, and scroll primitives."
    },
    feedback: {
      title: "Feedback",
      description: "Status, progress, loading, and state feedback components."
    },
    overlays: {
      title: "Overlays",
      description: "Dialog, menu, gallery, and context-menu layer wrappers."
    },
    "media-player": {
      title: "Media Player",
      description: "Video player layout, controls, timeline, toolbar, and queue pieces."
    },
    "danmaku-motion-rich-text": {
      title: "Danmaku, Motion, Rich Text",
      description: "Danmaku composition, motion staging, and Markdown-rich editing surfaces."
    }
  },
  ja: {
    actions: {
      title: "アクション",
      description: "ボタン、リンク、アクションバー、コマンド型ナビゲーション。"
    },
    forms: {
      title: "フォーム",
      description: "入力、選択、アップロード、色、日付、画像、エディタの操作部品。"
    },
    "layout-content": {
      title: "レイアウトとコンテンツ",
      description: "Surface、グリッド、Section、メタ情報、遅延読み込み、スケルトン、スクロール基盤。"
    },
    feedback: {
      title: "フィードバック",
      description: "状態、進捗、読み込み、ページ状態を伝えるコンポーネント。"
    },
    overlays: {
      title: "オーバーレイ",
      description: "Dialog、Menu、Gallery、コンテキストメニューのレイヤー wrapper。"
    },
    "media-player": {
      title: "メディアプレイヤー",
      description: "動画プレイヤーのレイアウト、操作部、タイムライン、ツールバー、キュー。"
    },
    "danmaku-motion-rich-text": {
      title: "弾幕・モーション・リッチテキスト",
      description: "弾幕作成、モーション演出、Markdown リッチ編集の画面部品。"
    }
  }
}

const categoryPurposeByLocale: Record<Exclude<AoiDocsLocale, "en">, Record<AoiComponentDocCategory, string>> = {
  "zh-CN": {
    actions: "组织命令、导航和按钮式操作",
    forms: "收集、校验和编辑用户输入",
    "layout-content": "搭建稳定的页面结构、内容密度和滚动区域",
    feedback: "展示加载、进度、成功、警告和错误状态",
    overlays: "承载临时菜单、确认流程和媒体预览",
    "media-player": "组合视频播放、时间轴、控制条和队列体验",
    "danmaku-motion-rich-text": "呈现弹幕、动效和富文本编辑体验"
  },
  ja: {
    actions: "コマンド、ナビゲーション、ボタン型操作をまとめること",
    forms: "ユーザー入力の収集、検証、編集を扱うこと",
    "layout-content": "安定したページ構造、情報密度、スクロール領域を作ること",
    feedback: "読み込み、進捗、成功、警告、エラー状態を示すこと",
    overlays: "一時的なメニュー、確認フロー、メディアプレビューを扱うこと",
    "media-player": "動画再生、タイムライン、操作部、キュー体験を組み立てること",
    "danmaku-motion-rich-text": "弾幕、モーション、リッチテキスト編集体験を表現すること"
  }
}

export function getLocalizedAoiComponentCategories(locale: AoiDocsLocale): AoiComponentDocCategoryMeta[] {
  return aoiComponentCategories.map((item) => ({
    ...item,
    ...categoryCopyByLocale[locale][item.id]
  }))
}

export function getLocalizedAoiComponentCategoryMeta(category: AoiComponentDocCategory, locale: AoiDocsLocale) {
  return getLocalizedAoiComponentCategories(locale).find((item) => item.id === category)
}

export function getLocalizedAoiComponentDocs(locale: AoiDocsLocale, category?: AoiComponentDocCategory) {
  const docs = getAoiComponentDocs(category)

  return locale === "en" ? docs : docs.map((doc) => localizeComponentDoc(doc, locale))
}

function localizeComponentDoc(doc: AoiComponentDoc, locale: Exclude<AoiDocsLocale, "en">): AoiComponentDoc {
  const categoryTitle = categoryCopyByLocale[locale][doc.category].title
  const purpose = categoryPurposeByLocale[locale][doc.category]

  return {
    ...doc,
    description: locale === "zh-CN"
      ? `${doc.name} 是 ${categoryTitle} 分类下的 Aoi wrapper，用于${purpose}。`
      : `${doc.name} は ${categoryTitle} カテゴリの Aoi wrapper で、${purpose}を担います。`,
    events: localizeApiRows(doc.events, "events", locale),
    notes: doc.notes.map((note) => localizeNote(note, locale)),
    props: localizeApiRows(doc.props, "props", locale),
    slots: localizeApiRows(doc.slots, "slots", locale),
    usage: locale === "zh-CN"
      ? `在需要${purpose}时使用，业务代码继续通过 Aoi wrapper 组合界面。`
      : `${purpose}が必要な場面で使い、業務コードは Aoi wrapper 経由で画面を組み立てます。`
  }
}

function localizeApiRows(rows: AoiComponentDocApiRow[], kind: AoiComponentDocApiKind, locale: Exclude<AoiDocsLocale, "en">) {
  return rows.map((item) => ({
    ...item,
    description: localizeApiRowDescription(item, kind, locale)
  }))
}

function localizeApiRowDescription(row: AoiComponentDocApiRow, kind: AoiComponentDocApiKind, locale: Exclude<AoiDocsLocale, "en">) {
  if (kind === "events") {
    return localizeEventDescription(row.name, locale)
  }

  if (kind === "slots") {
    return localizeSlotDescription(row.name, locale)
  }

  return localizePropDescription(row.name, locale)
}

function localizePropDescription(name: string, locale: Exclude<AoiDocsLocale, "en">) {
  const key = name.toLowerCase()

  if (locale === "zh-CN") {
    if (key.includes("modelvalue")) return "v-model 绑定值。"
    if (key.includes("items") || key.includes("options") || key.includes("sources")) return "传入组件的数据列表。"
    if (key.includes("label") || key.includes("title") || key.includes("description") || key.includes("placeholder") || key.includes("supportingtext") || key.includes("emptytext")) return "显示文案或无障碍说明文本。"
    if (key.includes("icon")) return "通过 AoiIcon 渲染的 Lucide 图标名称。"
    if (key.includes("to") || key.includes("href") || key.includes("target") || key.includes("rel")) return "导航目标或外链关系设置。"
    if (key.includes("appearance") || key.includes("surface") || key.includes("intent") || key.includes("type") || key.includes("layout") || key.includes("variant") || key.includes("mode") || key.includes("positioning") || key.includes("shape")) return "控制组件的视觉样式或行为模式。"
    if (key.includes("size") || key.includes("density") || key.includes("compact") || key.includes("gap") || key.includes("padding") || key.includes("columns") || key.includes("width") || key.includes("height") || key.includes("lines") || key.includes("count")) return "控制尺寸、密度或布局数量。"
    if (key.includes("min") || key.includes("max") || key.includes("step") || key.includes("value") || key.includes("time") || key.includes("duration") || key.includes("rate") || key.includes("volume") || key.includes("index") || key.includes("x / y")) return "数值范围、当前位置或当前状态值。"
    if (key.includes("disabled") || key.includes("loading") || key.includes("enabled") || key.includes("playing") || key.includes("open") || key.includes("dismissible") || key.includes("loop") || key.includes("snap") || key.includes("rubberband") || key.includes("visible") || key.includes("animated") || key.includes("interactive") || key.includes("selected") || key.includes("active")) return "控制开启、禁用、动画、选中或交互状态。"
    if (key.includes("src") || key.includes("poster") || key.includes("placeholder") || key.includes("alt")) return "媒体资源、占位内容或替代文本。"
    if (key.includes("document") || key.includes("mapper") || key.includes("settings")) return "传入高级对象配置或数据映射规则。"
    if (key.includes("accept") || key.includes("multiple")) return "控制文件选择限制和选择数量。"
    if (key.includes("as")) return "覆盖组件根节点渲染标签。"

    return `${name} 配置项。`
  }

  if (key.includes("modelvalue")) return "v-model で同期する値です。"
  if (key.includes("items") || key.includes("options") || key.includes("sources")) return "コンポーネントへ渡すデータ一覧です。"
  if (key.includes("label") || key.includes("title") || key.includes("description") || key.includes("placeholder") || key.includes("supportingtext") || key.includes("emptytext")) return "表示文言またはアクセシビリティ用の説明です。"
  if (key.includes("icon")) return "AoiIcon で描画する Lucide アイコン名です。"
  if (key.includes("to") || key.includes("href") || key.includes("target") || key.includes("rel")) return "遷移先または外部リンクの関係設定です。"
  if (key.includes("appearance") || key.includes("surface") || key.includes("intent") || key.includes("type") || key.includes("layout") || key.includes("variant") || key.includes("mode") || key.includes("positioning") || key.includes("shape")) return "見た目または動作モードを制御します。"
  if (key.includes("size") || key.includes("density") || key.includes("compact") || key.includes("gap") || key.includes("padding") || key.includes("columns") || key.includes("width") || key.includes("height") || key.includes("lines") || key.includes("count")) return "サイズ、密度、レイアウト数を制御します。"
  if (key.includes("min") || key.includes("max") || key.includes("step") || key.includes("value") || key.includes("time") || key.includes("duration") || key.includes("rate") || key.includes("volume") || key.includes("index") || key.includes("x / y")) return "数値範囲、現在位置、現在状態を表します。"
  if (key.includes("disabled") || key.includes("loading") || key.includes("enabled") || key.includes("playing") || key.includes("open") || key.includes("dismissible") || key.includes("loop") || key.includes("snap") || key.includes("rubberband") || key.includes("visible") || key.includes("animated") || key.includes("interactive") || key.includes("selected") || key.includes("active")) return "有効、無効、アニメーション、選択、操作状態を制御します。"
  if (key.includes("src") || key.includes("poster") || key.includes("placeholder") || key.includes("alt")) return "メディアリソース、プレースホルダー、代替テキストです。"
  if (key.includes("document") || key.includes("mapper") || key.includes("settings")) return "高度なオブジェクト設定またはデータ変換ルールです。"
  if (key.includes("accept") || key.includes("multiple")) return "ファイル選択の制約と選択数を制御します。"
  if (key.includes("as")) return "コンポーネントのルート要素タグを上書きします。"

  return `${name} の設定項目です。`
}

function localizeEventDescription(name: string, locale: Exclude<AoiDocsLocale, "en">) {
  const key = name.toLowerCase()

  if (locale === "zh-CN") {
    if (key.startsWith("update:")) return "对应状态变化时触发。"
    if (key.includes("toggle")) return "请求切换对应状态时触发。"
    if (key.includes("select")) return "用户选择条目时触发。"
    if (key.includes("seek")) return "请求跳转到指定播放时间时触发。"
    if (key.includes("submit") || key.includes("send")) return "提交新内容时触发。"
    if (key.includes("change")) return "值或内容发生变化时触发。"
    if (key.includes("error")) return "出现错误时触发。"
    if (key.includes("load") || key.includes("visible")) return "内容加载或进入可见区域时触发。"
    if (key.includes("click")) return "用户激活命令时触发。"
    if (key.includes("cancel") || key.includes("closed")) return "浮层取消或关闭时触发。"

    return "组件事件触发时携带该 payload。"
  }

  if (key.startsWith("update:")) return "対応する状態が変わったときに発火します。"
  if (key.includes("toggle")) return "対応する状態の切り替えを要求するときに発火します。"
  if (key.includes("select")) return "ユーザーが項目を選択したときに発火します。"
  if (key.includes("seek")) return "指定した再生時刻へ移動するときに発火します。"
  if (key.includes("submit") || key.includes("send")) return "新しい内容を送信するときに発火します。"
  if (key.includes("change")) return "値または内容が変わったときに発火します。"
  if (key.includes("error")) return "エラーが発生したときに発火します。"
  if (key.includes("load") || key.includes("visible")) return "内容の読み込みまたは表示領域への進入時に発火します。"
  if (key.includes("click")) return "ユーザーがコマンドを実行したときに発火します。"
  if (key.includes("cancel") || key.includes("closed")) return "レイヤーがキャンセルまたは閉じられたときに発火します。"

  return "コンポーネントイベントの payload です。"
}

function localizeSlotDescription(name: string, locale: Exclude<AoiDocsLocale, "en">) {
  const key = name.toLowerCase()

  if (locale === "zh-CN") {
    if (key.includes("actions")) return "自定义动作区域。"
    if (key.includes("media")) return "媒体或图标区域。"
    if (key.includes("title") || key.includes("headline")) return "自定义标题区域。"
    if (key.includes("controls") || key.includes("panel")) return "自定义控制或面板区域。"
    if (key.includes("details") || key.includes("aside")) return "播放器详情或侧栏区域。"
    if (key.includes("default")) return "默认内容插槽。"

    return `${name} 插槽内容。`
  }

  if (key.includes("actions")) return "カスタム操作領域です。"
  if (key.includes("media")) return "メディアまたはアイコン領域です。"
  if (key.includes("title") || key.includes("headline")) return "カスタム見出し領域です。"
  if (key.includes("controls") || key.includes("panel")) return "カスタム操作部またはパネル領域です。"
  if (key.includes("details") || key.includes("aside")) return "プレイヤー詳細またはサイド領域です。"
  if (key.includes("default")) return "既定のコンテンツスロットです。"

  return `${name} スロットの内容です。`
}

function localizeNote(note: string, locale: Exclude<AoiDocsLocale, "en">) {
  const key = note.toLowerCase()

  if (locale === "zh-CN") {
    if (key.includes("client-only")) return "该组件依赖浏览器能力，文档和业务页面应通过 ClientOnly 边界使用。"
    if (key.includes("media sources")) return "媒体资源应尽量保持懒加载和视口感知。"
    if (key.includes("unsafe links")) return "富文本中的链接和图片仍由组件负责校验。"
    if (key.includes("media apis")) return "需要媒体 API 时应放在 ClientOnly 中。"

    return "注意该组件的运行边界和加载成本。"
  }

  if (key.includes("client-only")) return "このコンポーネントはブラウザー機能に依存するため、ドキュメントと業務画面では ClientOnly 境界で使います。"
  if (key.includes("media sources")) return "メディアリソースは可能な限り遅延読み込みと表示領域検知を保ちます。"
  if (key.includes("unsafe links")) return "リッチテキスト内のリンクと画像は引き続きコンポーネント側で検証します。"
  if (key.includes("media apis")) return "メディア API が必要な場合は ClientOnly 内で使います。"

  return "このコンポーネントの実行境界と読み込みコストに注意します。"
}


