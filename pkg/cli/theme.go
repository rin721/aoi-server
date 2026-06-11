package cli

import "charm.land/lipgloss/v2"

// Theme 控制内置 TUI 首页的展示样式。
type Theme struct {
	// Title 是首页标题样式。
	Title lipgloss.Style
	// Subtitle 是首页描述文本样式。
	Subtitle lipgloss.Style
	// ListItem 是普通命令项样式。
	ListItem lipgloss.Style
	// Selected 是当前选中命令项样式。
	Selected lipgloss.Style
	// Help 是命令帮助面板样式。
	Help lipgloss.Style
	// Hint 是底部操作提示样式。
	Hint lipgloss.Style
	// Empty 是没有注册命令时的空状态样式。
	Empty lipgloss.Style
}

// DefaultTheme 返回交互式首页的默认终端主题。
func DefaultTheme() Theme {
	return Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginBottom(1),
		ListItem: lipgloss.NewStyle().
			PaddingLeft(2),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			PaddingLeft(2),
		Help: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1),
		Empty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			PaddingLeft(2),
	}
}
