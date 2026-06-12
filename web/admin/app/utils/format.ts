export function formatDateTime(value?: string | null) {
  if (!value) {
    return "-"
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date)
}

export function formatStatus(value?: string | null) {
  const map: Record<string, string> = {
    active: "启用",
    disabled: "禁用",
    expired: "已过期",
    pending: "待处理",
    revoked: "已撤销",
    used: "已使用"
  }

  return value ? map[value] || value : "-"
}

export function errorMessage(error: unknown) {
  if (error && typeof error === "object" && "message" in error) {
    const message = String((error as { message?: unknown }).message || "")
    if (message === "invalid iam input") {
      return "输入不符合 IAM 要求，请检查必填项和密码规则。"
    }
    return message || "请求失败"
  }

  return "请求失败"
}



