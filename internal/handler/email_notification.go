package handler

import (
	"net/http"
	"strconv"

	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	secutils "github.com/Tencent/WeKnora/internal/utils"
	"github.com/gin-gonic/gin"
)

// EmailNotificationHandler 邮件通知处理器
type EmailNotificationHandler struct {
	emailService interfaces.EmailNotificationService
	kbService    interfaces.KnowledgeBaseService
	userService  interfaces.UserService
}

// NewEmailNotificationHandler 创建邮件通知处理器实例
func NewEmailNotificationHandler(
	emailService interfaces.EmailNotificationService,
	kbService interfaces.KnowledgeBaseService,
	userService interfaces.UserService,
) *EmailNotificationHandler {
	return &EmailNotificationHandler{
		emailService: emailService,
		kbService:    kbService,
		userService:  userService,
	}
}

// SendKBUpdateNotification godoc
// @Summary      发送知识库更新通知
// @Description  知识库更新后，向指定人员发送邮件通知
// @Tags         邮件通知
// @Accept       json
// @Produce      json
// @Param        request  body      types.EmailNotificationRequest  true  "邮件通知请求"
// @Success      200      {object}  map[string]interface{}          "发送结果"
// @Failure      400      {object}  errors.AppError                 "请求参数错误"
// @Failure      404      {object}  errors.AppError                 "知识库不存在"
// @Failure      500      {object}  errors.AppError                 "内部服务错误"
// @Security     Bearer
// @Security     ApiKeyAuth
// @Router       /email-notifications/kb-update [post]
func (h *EmailNotificationHandler) SendKBUpdateNotification(c *gin.Context) {
	ctx := c.Request.Context()
	logger.Info(ctx, "开始发送知识库更新邮件通知")

	// 解析请求体
	var req types.EmailNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(ctx, "解析邮件通知请求参数失败", err)
		c.Error(apperrors.NewBadRequestError("请求参数错误").WithDetails(err.Error()))
		return
	}

	logger.Infof(ctx, "发送知识库更新邮件通知，知识库ID: %s, 收件人数量: %d",
		secutils.SanitizeForLog(req.KnowledgeBaseID), len(req.Recipients))

	// 查询知识库信息以获取知识库名称
	kb, err := h.kbService.GetKnowledgeBaseByID(ctx, req.KnowledgeBaseID)
	if err != nil {
		logger.Error(ctx, "获取知识库信息失败", err)
		c.Error(apperrors.NewNotFoundError("知识库不存在或无权访问"))
		return
	}

	// 发送邮件通知
	resp, err := h.emailService.SendKBUpdateNotification(ctx, kb.Name, &req)
	if err != nil {
		logger.Error(ctx, "发送邮件通知失败", err)
		c.Error(apperrors.NewInternalServerError(err.Error()))
		return
	}

	logger.Infof(ctx, "知识库更新邮件通知发送完成，成功: %d, 失败: %d",
		resp.SuccessCount, resp.FailCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// SearchPlatformUsers godoc
// @Summary      搜索平台用户
// @Description  根据用户名或邮箱搜索平台用户，用于知识推送时选择收件人
// @Tags         邮件通知
// @Produce      json
// @Param        q      query  string  false  "搜索关键词（用户名或邮箱）"
// @Param        limit  query  int     false  "返回数量限制" default(20)
// @Success      200    {object}  map[string]interface{}
// @Security     Bearer
// @Router       /email-notifications/search-users [get]
func (h *EmailNotificationHandler) SearchPlatformUsers(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Query("q")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}

	users, err := h.userService.SearchUsers(ctx, query, limit)
	if err != nil {
		logger.Errorf(ctx, "搜索平台用户失败: %v", err)
		c.Error(apperrors.NewInternalServerError("搜索用户失败"))
		return
	}

	result := make([]gin.H, 0, len(users))
	for _, u := range users {
		result = append(result, gin.H{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"avatar":   u.Avatar,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
