// Package i18n 提供运行时国际化能力。
//
// 设计思路：
//   - 翻译文件为 YAML 格式，存放在 lang/<locale>.yaml 中。
//   - 支持通过 Accept-Language 头或查询参数自动识别语言。
//   - 支持消息模板变量替换（{{ .Name }} 风格）。
//   - 线程安全，可在 Gin 中间件和 Handler 中直接使用。
//
// 使用方式：
//
//	bundle := i18n.NewBundle("lang", "zh-CN")
//	bundle.LoadAll()
//	r.Use(i18n.Middleware(bundle))
//	// 在 Handler 中：
//	msg := i18n.T(c, "welcome", map[string]string{"Name": "张三"})
package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const (
	ginContextKey    = "thinkgin:i18n:locale"
	ginBundleCtxKey  = "thinkgin:i18n:bundle"
)

// Bundle 存放所有语言的翻译映射。
type Bundle struct {
	mu          sync.RWMutex
	langDir     string
	fallback    string
	messages    map[string]map[string]string // locale → key → message
}

// NewBundle 创建翻译包。
// langDir 为存放翻译 YAML 的目录，fallback 为默认语言。
func NewBundle(langDir string, fallback string) *Bundle {
	return &Bundle{
		langDir:  langDir,
		fallback: fallback,
		messages: make(map[string]map[string]string),
	}
}

// LoadAll 扫描 langDir 加载所有 .yaml 文件。
// 文件名即 locale（如 zh-CN.yaml → "zh-CN"）。
func (b *Bundle) LoadAll() error {
	entries, err := os.ReadDir(b.langDir)
	if err != nil {
		return fmt.Errorf("i18n: read dir %s: %w", b.langDir, err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		locale := strings.TrimSuffix(e.Name(), ext)
		data, err := os.ReadFile(filepath.Join(b.langDir, e.Name()))
		if err != nil {
			return fmt.Errorf("i18n: read %s: %w", e.Name(), err)
		}

		var flat map[string]string
		if err := yaml.Unmarshal(data, &flat); err != nil {
			return fmt.Errorf("i18n: parse %s: %w", e.Name(), err)
		}
		b.messages[locale] = flat
	}
	return nil
}

// Translate 返回指定 locale 和 key 的翻译文本。
// params 中的键可在消息模板中以 {{.Key}} 形式引用。
// 找不到翻译时回退 fallback 语言，仍找不到返回 key 本身。
func (b *Bundle) Translate(locale, key string, params map[string]string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msg := b.lookup(locale, key)
	if msg == "" {
		msg = b.lookup(b.fallback, key)
	}
	if msg == "" {
		return key
	}

	// 简单模板替换
	for k, v := range params {
		msg = strings.ReplaceAll(msg, "{{."+k+"}}", v)
	}
	return msg
}

func (b *Bundle) lookup(locale, key string) string {
	m, ok := b.messages[locale]
	if !ok {
		return ""
	}
	return m[key]
}

// Middleware 注入语言检测中间件。
// 优先级：?lang= 查询参数 > Accept-Language 头 > fallback。
func Middleware(bundle *Bundle) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := c.Query("lang")
		if locale == "" {
			locale = parseAcceptLanguage(c.GetHeader("Accept-Language"))
		}
		if locale == "" {
			locale = bundle.fallback
		}
		c.Set(ginContextKey, locale)
		c.Set(ginBundleCtxKey, bundle)
		c.Next()
	}
}

// T 是 Handler 中的便捷翻译函数。
// 从 gin.Context 中获取当前 locale 和 bundle，调用 Translate。
func T(c *gin.Context, key string, params ...map[string]string) string {
	locale, _ := c.Get(ginContextKey)
	l, _ := locale.(string)

	bundleVal, _ := c.Get(ginBundleCtxKey)
	bundle, _ := bundleVal.(*Bundle)
	if bundle == nil {
		return key
	}

	var p map[string]string
	if len(params) > 0 {
		p = params[0]
	}
	return bundle.Translate(l, key, p)
}

// parseAcceptLanguage 从 Accept-Language 中提取首选语言标签。
// 例："zh-CN,zh;q=0.9,en;q=0.8" → "zh-CN"
func parseAcceptLanguage(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return ""
	}
	tag := strings.TrimSpace(parts[0])
	// 去掉 ;q=... 后缀
	if idx := strings.Index(tag, ";"); idx >= 0 {
		tag = tag[:idx]
	}
	return strings.TrimSpace(tag)
}
