package types

// EmailNotificationRequest 知识库更新邮件通知请求
type EmailNotificationRequest struct {
	// 知识库ID
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required"`
	// 收件人邮箱列表
	Recipients []string `json:"recipients" binding:"required,min=1"`
	// 自定义通知消息（用户编辑的一句话）
	Message string `json:"message" binding:"required"`
	// 更新内容摘要（可选，描述知识库更新了什么）
	UpdateSummary string `json:"update_summary,omitempty"`
}

// EmailNotificationResponse 邮件通知响应
type EmailNotificationResponse struct {
	// 成功发送的邮箱数量
	SuccessCount int `json:"success_count"`
	// 失败的邮箱数量
	FailCount int `json:"fail_count"`
	// 发送失败的详情
	FailedRecipients []FailedRecipient `json:"failed_recipients,omitempty"`
}

// FailedRecipient 发送失败的收件人信息
type FailedRecipient struct {
	// 邮箱地址
	Email string `json:"email"`
	// 失败原因
	Reason string `json:"reason"`
}
