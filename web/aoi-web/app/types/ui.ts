import type { AoiRevealDirectiveValue } from "~/utils/aoiReveal"

export type AoiLayoutMode = "stack" | "grid" | "inline" | "split"
export type AoiIntent = "primary" | "secondary" | "neutral" | "success" | "warning" | "danger" | "info"
export type AoiActionAppearance = "solid" | "soft" | "outline" | "plain" | "elevated"
export type AoiFieldAppearance = "filled" | "outlined"
export type AoiSurfaceKind = "plain" | "panel" | "card" | "state" | "code" | "toolbar"
export type AoiSurfacePadding = "none" | "sm" | "md" | "lg"
export type AoiContentGridGap = "normal" | "compact" | "video"
export type AoiInfoCardDensity = "default" | "compact"
export type AoiInfoCardLayout = "inline" | "stack"

export interface AoiStatItem {
  description?: string
  icon?: string
  intent?: AoiIntent
  label: string
  value: number | string
}

export interface AoiTagItem {
  external?: boolean
  href?: string
  icon?: string
  label: string
  target?: string
  to?: string
  value?: string
}

export type AoiRevealProp = AoiRevealDirectiveValue


