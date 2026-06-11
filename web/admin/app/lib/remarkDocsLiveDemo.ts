type MarkdownNode = {
  attributes?: Record<string, unknown>
  children?: MarkdownNode[]
  lang?: string | null
  meta?: string | null
  name?: string
  type: string
  value?: string
}

type LiveMeta = Record<string, boolean | string>

export function remarkDocsLiveDemo() {
  return (tree: MarkdownNode) => {
    transformChildren(tree)
  }
}

export default remarkDocsLiveDemo

function transformChildren(node: MarkdownNode) {
  if (!Array.isArray(node.children)) {
    return
  }

  node.children = node.children.map((child) => {
    if (isLiveMdcCode(child)) {
      return toDocsLiveDemoNode(child)
    }

    transformChildren(child)
    return child
  })
}

function isLiveMdcCode(node: MarkdownNode) {
  if (node.type !== "code" || node.lang !== "mdc") {
    return false
  }

  return parseLiveMeta(node.meta || "").live === true
}

function toDocsLiveDemoNode(node: MarkdownNode): MarkdownNode {
  const meta = parseLiveMeta(node.meta || "")
  const attributes: Record<string, unknown> = {
    code: node.value || "",
    language: node.lang || "mdc",
    meta: node.meta || ""
  }

  for (const key of ["description", "title", "unwrap"]) {
    if (typeof meta[key] === "string") {
      attributes[key] = meta[key]
    }
  }

  if (meta.client === true || meta.client === "true") {
    attributes.client = true
  }

  return {
    attributes,
    children: [],
    name: "docs-live-demo",
    type: "containerComponent"
  }
}

function parseLiveMeta(meta: string): LiveMeta {
  const result: LiveMeta = {}
  const pattern = /(?:^|\s)([A-Za-z][\w-]*)(?:=(?:"([^"]*)"|'([^']*)'|([^\s]+)))?/g

  for (const match of meta.matchAll(pattern)) {
    const [, key, doubleQuoted, singleQuoted, bare] = match

    if (!key) {
      continue
    }

    result[key] = doubleQuoted ?? singleQuoted ?? bare ?? true
  }

  return result
}


