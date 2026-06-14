package middleware

// 本文件定义 Gin 中间件能力，约束请求进入业务 handler 前后的链路上下文、副作用和错误输出。

import "github.com/rei0721/go-scaffold/internal/ports"

const LanguageHeader = "Accept-Language"

// I18n 中间件提取并存储用户的语言偏好
func I18n(i18nApp ports.I18n) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		// 从 Accept-Language 头部获取语言
		lang := c.GetHeader(LanguageHeader)

		// 如果语言不支持,使用默认语言
		if lang == "" || !i18nApp.IsSupported(lang) {
			lang = i18nApp.GetDefaultLanguage()
		}

		// 存储到上下文
		c.Set("lang", lang)
		c.Set("i18n", i18nApp)

		c.Next()
	}
}
