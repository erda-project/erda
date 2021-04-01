package i18n

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func InitI18N() {
	message.SetString(language.SimplifiedChinese, "ImagePullFailed", "拉取镜像失败")
	message.SetString(language.SimplifiedChinese, "Unschedulable", "调度失败")
	message.SetString(language.SimplifiedChinese, "InsufficientResources", "资源不足")
	message.SetString(language.SimplifiedChinese, "ProbeFailed", "健康检查失败")
	message.SetString(language.SimplifiedChinese, "ContainerCannotRun", "容器无法启动")
}
