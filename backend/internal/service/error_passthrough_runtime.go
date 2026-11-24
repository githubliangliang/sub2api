package service

import (
	"github.com/Wei-Shaw/sub2api/internal/model"
	"github.com/gin-gonic/gin"
)

const errorPassthroughServiceContextKey = "error_passthrough_service"

// BindErrorPassthroughService 将错误透传服务绑定到请求上下文，供 service 层在非 failover 场景下复用规则。
func BindErrorPassthroughService(c *gin.Context, svc *ErrorPassthroughService) {
	if c == nil || svc == nil {
		return
	}
	c.Set(errorPassthroughServiceContextKey, svc)
}

func getBoundErrorPassthroughService(c *gin.Context) *ErrorPassthroughService {
	if c == nil {
		return nil
	}
	v, ok := c.Get(errorPassthroughServiceContextKey)
	if !ok {
		return nil
	}
	svc, ok := v.(*ErrorPassthroughService)
	if !ok {
		return nil
	}
	return svc
}

// applyErrorPassthroughRule 按规则改写错误响应；未命中时返回默认响应参数。
func applyErrorPassthroughRule(
	c *gin.Context,
	platform string,
	upstreamStatus int,
	responseBody []byte,
	defaultStatus int,
	defaultErrType string,
	defaultErrMsg string,
) (status int, errType string, errMsg string, matched bool) {
	return applyErrorPassthroughRuleWithMessage(c, platform, upstreamStatus, responseBody, defaultStatus, defaultErrType, defaultErrMsg, ExtractUpstreamErrorMessage)
}

// applyErrorPassthroughRuleWithRule also returns the matched policy so callers
// can distinguish structured passthrough from an explicit administrator message.
func applyErrorPassthroughRuleWithRule(
	c *gin.Context, platform string, upstreamStatus int, responseBody []byte,
	defaultStatus int, defaultErrType, defaultErrMsg string,
) (status int, errType, errMsg string, rule *model.ErrorPassthroughRule, matched bool) {
	return applyErrorPassthroughRuleDetailed(c, platform, upstreamStatus, responseBody,
		defaultStatus, defaultErrType, defaultErrMsg, ExtractUpstreamErrorMessage)
}

// Grok media supplies its own extractor for xAI's flat error fields.
func applyErrorPassthroughRuleWithMessage(
	c *gin.Context,
	platform string,
	upstreamStatus int,
	responseBody []byte,
	defaultStatus int,
	defaultErrType string,
	defaultErrMsg string,
	extractMessage func([]byte) string,
) (status int, errType string, errMsg string, matched bool) {
	status, errType, errMsg, _, matched = applyErrorPassthroughRuleDetailed(c, platform, upstreamStatus,
		responseBody, defaultStatus, defaultErrType, defaultErrMsg, extractMessage)
	return
}

func applyErrorPassthroughRuleDetailed(
	c *gin.Context, platform string, upstreamStatus int, responseBody []byte,
	defaultStatus int, defaultErrType, defaultErrMsg string, extractMessage func([]byte) string,
) (status int, errType, errMsg string, rule *model.ErrorPassthroughRule, matched bool) {
	status = defaultStatus
	errType = defaultErrType
	errMsg = defaultErrMsg

	svc := getBoundErrorPassthroughService(c)
	if svc == nil {
		return status, errType, errMsg, nil, false
	}

	rule = svc.MatchRule(platform, upstreamStatus, responseBody)
	if rule == nil {
		return status, errType, errMsg, nil, false
	}

	status = upstreamStatus
	if !rule.PassthroughCode && rule.ResponseCode != nil {
		status = *rule.ResponseCode
	}

	errMsg = extractMessage(responseBody)
	if !rule.PassthroughBody && rule.CustomMessage != nil {
		errMsg = *rule.CustomMessage
	}

	// 命中 skip_monitoring 时在 context 中标记，供 ops_error_logger 跳过记录。
	if rule.SkipMonitoring {
		c.Set(OpsSkipPassthroughKey, true)
	}

	// 与现有 failover 场景保持一致：命中规则时统一返回 upstream_error。
	errType = "upstream_error"
	return status, errType, errMsg, rule, true
}
