package app

// Config 是进程级的全局配置单例。
// 通过 init 或 Bootstrap 填充；业务代码只应通过 GetConfig/GetXxxConfig 访问。
var Config *GlobalConfig

// GetConfig 返回全局配置。如果尚未初始化，返回 nil。
// 调用方应确保在 main 或 init 阶段完成 Bootstrap 后再使用。
//
// 已有的细分 Getter（GetAppConfig / GetServerConfig 等）被移除，
// 因为项目内 30 处调用全部走 GetConfig() 再取字段，细分 Getter 属于死代码。
func GetConfig() *GlobalConfig {
	return Config
}
