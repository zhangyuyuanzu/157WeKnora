package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
)

// EmailNotificationService 邮件通知服务接口
// 提供知识库更新后向指定人员发送邮件通知的功能
type EmailNotificationService interface {
	// SendKBUpdateNotification 发送知识库更新通知邮件
	// Parameters:
	//   - ctx: 上下文信息，携带请求追踪、用户身份等
	//   - knowledgeBaseName: 知识库名称
	//   - req: 邮件通知请求，包含收件人列表、自定义消息等
	// Returns:
	//   - 邮件通知响应，包含成功/失败的发送结果
	//   - 可能的错误
	SendKBUpdateNotification(ctx context.Context, knowledgeBaseName string, req *types.EmailNotificationRequest) (*types.EmailNotificationResponse, error)
}
