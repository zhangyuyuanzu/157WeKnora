package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// emailNotificationService 邮件通知服务实现
type emailNotificationService struct {
	cfg *config.EmailConfig
}

// NewEmailNotificationService 创建邮件通知服务实例
func NewEmailNotificationService(cfg *config.Config) interfaces.EmailNotificationService {
	emailCfg := cfg.Email
	if emailCfg == nil {
		emailCfg = &config.EmailConfig{}
	}
	return &emailNotificationService{
		cfg: emailCfg,
	}
}

// SendKBUpdateNotification 发送知识库更新通知邮件
func (s *emailNotificationService) SendKBUpdateNotification(
	ctx context.Context,
	knowledgeBaseName string,
	req *types.EmailNotificationRequest,
) (*types.EmailNotificationResponse, error) {
	if !s.cfg.Enabled {
		return nil, fmt.Errorf("邮件通知功能未启用，请在配置文件中启用 email.enabled")
	}

	if s.cfg.SMTPHost == "" || s.cfg.From == "" {
		return nil, fmt.Errorf("邮件配置不完整，请检查 SMTP 主机和发件人配置")
	}

	resp := &types.EmailNotificationResponse{}

	subject := fmt.Sprintf("【知识库更新通知】%s 已更新", knowledgeBaseName)
	body := s.buildEmailBody(knowledgeBaseName, req.Message, req.UpdateSummary)

	for _, recipient := range req.Recipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}

		logger.Infof(ctx, "正在发送知识库更新通知邮件至: %s", recipient)

		err := s.sendEmail(ctx, recipient, subject, body)
		if err != nil {
			logger.Errorf(ctx, "发送邮件至 %s 失败: %v", recipient, err)
			resp.FailCount++
			resp.FailedRecipients = append(resp.FailedRecipients, types.FailedRecipient{
				Email:  recipient,
				Reason: err.Error(),
			})
		} else {
			logger.Infof(ctx, "成功发送知识库更新通知邮件至: %s", recipient)
			resp.SuccessCount++
		}
	}

	return resp, nil
}

// buildEmailBody 构建邮件正文（HTML 格式）
func (s *emailNotificationService) buildEmailBody(kbName, message, updateSummary string) string {
	now := time.Now().Format("2006-01-02 15:04:05")

	var summarySection string
	if updateSummary != "" {
		summarySection = fmt.Sprintf(`
            <div style="background-color: #f0f7ff; border-left: 4px solid #1890ff; padding: 12px 16px; margin: 16px 0; border-radius: 4px;">
                <p style="margin: 0; color: #333; font-weight: bold;">📋 更新内容：</p>
                <p style="margin: 8px 0 0 0; color: #555;">%s</p>
            </div>`, updateSummary)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>知识库更新通知</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); overflow: hidden;">
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #1890ff, #096dd9); padding: 24px 32px; color: white;">
            <h1 style="margin: 0; font-size: 20px; font-weight: 600;">📚 知识库更新通知</h1>
        </div>
        <!-- Content -->
        <div style="padding: 32px;">
            <p style="color: #333; font-size: 15px; line-height: 1.6; margin: 0 0 16px 0;">您好，</p>
            <p style="color: #333; font-size: 15px; line-height: 1.6; margin: 0 0 16px 0;">
                知识库 <strong style="color: #1890ff;">%s</strong> 已经更新。
            </p>
            <!-- 自定义消息 -->
            <div style="background-color: #fff7e6; border-left: 4px solid #ffa940; padding: 12px 16px; margin: 16px 0; border-radius: 4px;">
                <p style="margin: 0; color: #333; font-weight: bold;">💬 通知消息：</p>
                <p style="margin: 8px 0 0 0; color: #555;">%s</p>
            </div>
            %s
            <p style="color: #999; font-size: 13px; margin: 24px 0 0 0;">通知时间：%s</p>
        </div>
        <!-- Footer -->
        <div style="background-color: #fafafa; padding: 16px 32px; border-top: 1px solid #f0f0f0;">
            <p style="margin: 0; color: #999; font-size: 12px; text-align: center;">
                此邮件由 WeKnora 知识库管理系统自动发送，请勿直接回复。
            </p>
        </div>
    </div>
</body>
</html>`, kbName, message, summarySection, now)
}

// sendEmail 发送单封邮件
func (s *emailNotificationService) sendEmail(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	header := make(map[string]string)
	header["From"] = s.cfg.From
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"
	header["Date"] = time.Now().Format(time.RFC1123Z)

	var msg strings.Builder
	for k, v := range header {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// 根据配置选择是否使用 TLS
	if s.cfg.UseTLS {
		return s.sendWithTLS(addr, to, msg.String())
	}
	return s.sendWithPlain(addr, to, msg.String())
}

// sendWithPlain 使用普通 SMTP 发送（支持 STARTTLS 升级）
func (s *emailNotificationService) sendWithPlain(addr, to, msg string) error {
	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.SMTPHost)
	}
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg))
}

// sendWithTLS 使用直接 TLS 连接发送（如 465 端口）
func (s *emailNotificationService) sendWithTLS(addr, to, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: s.cfg.SMTPHost,
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	// 认证
	if s.cfg.Username != "" {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}

	// 发送邮件
	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("打开数据通道失败: %w", err)
	}

	if _, err := writer.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件数据失败: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("关闭数据通道失败: %w", err)
	}

	return client.Quit()
}
