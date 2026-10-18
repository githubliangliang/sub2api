# 验收测试清单

本轮修改或新增的测试均参与相应默认/unit/Vitest 检查；没有把 PostgreSQL 专用 integration 当作 SQLite 验收。S08 使用真实文件 SQLite/WAL 多连接，S04/S06 使用本地 WS 请求。

| 文件 | 构建标签 |
|---|---|
| `backend/internal/handler/gateway_models_test.go` | 默认 / Vitest |
| `backend/internal/handler/openai_gateway_handler_test.go` | 默认 / Vitest |
| `backend/internal/handler/openai_gateway_ws_group_pricing_test.go` | 默认 / Vitest |
| `backend/internal/handler/openai_ws_turn_pricing_test.go` | 默认 / Vitest |
| `backend/internal/handler/ops_error_logger_google_test.go` | 默认 / Vitest |
| `backend/internal/handler/ops_error_logger_test.go` | 默认 / Vitest |
| `backend/internal/pkg/apicompat/chatcompletions_to_responses_legacy_function_call_test.go` | 默认 / Vitest |
| `backend/internal/pkg/apicompat/responses_to_chatcompletions_refusal_test.go` | 默认 / Vitest |
| `backend/internal/repository/admin_bootstrap_test.go` | 默认 / Vitest |
| `backend/internal/repository/group_moderation_sqlite_test.go` | 默认 / Vitest |
| `backend/internal/repository/group_repo_delete_cascade_sqlite_test.go` | 默认 / Vitest |
| `backend/internal/service/account_stats_pricing_test.go` | //go:build unit |
| `backend/internal/service/composite_route_resolver_test.go` | 默认 / Vitest |
| `backend/internal/service/content_moderation_scope_test.go` | 默认 / Vitest |
| `backend/internal/service/gateway_forward_as_chat_completions_ping_test.go` | //go:build unit |
| `backend/internal/service/gateway_forward_as_chat_completions_test.go` | //go:build unit |
| `backend/internal/service/gateway_structured_outputs_beta_test.go` | //go:build unit |
| `backend/internal/service/gateway_tool_changes_beta_test.go` | //go:build unit |
| `backend/internal/service/gemini_cached_content_not_found_test.go` | //go:build unit |
| `backend/internal/service/gemini_error_policy_test.go` | //go:build unit |
| `backend/internal/service/gemini_flash_promotion_test.go` | 默认 / Vitest |
| `backend/internal/service/gemini_quota_test.go` | //go:build unit |
| `backend/internal/service/openai_alpha_search_billing_test.go` | //go:build unit |
| `backend/internal/service/openai_gateway_chat_completions_test.go` | 默认 / Vitest |
| `backend/internal/service/openai_gateway_responses_chat_fallback_test.go` | //go:build unit |
| `backend/internal/service/openai_gateway_search_surcharge_test.go` | //go:build unit |
| `backend/internal/service/openai_gateway_service_hotpath_test.go` | 默认 / Vitest |
| `backend/internal/service/openai_oauth_web_search_history_forward_test.go` | //go:build unit |
| `backend/internal/service/openai_oauth_web_search_history_test.go` | 默认 / Vitest |
| `backend/internal/service/openai_oauth_web_search_history_ws_test.go` | 默认 / Vitest |
| `backend/internal/service/openai_responses_namespace_benchmark_test.go` | 默认 / Vitest |
| `backend/internal/service/openai_responses_namespace_test.go` | 默认 / Vitest |
| `backend/internal/service/openai_subscription_long_context_test.go` | //go:build unit |
| `backend/internal/service/openai_ws_v2/passthrough_relay_image_usage_test.go` | 默认 / Vitest |
| `backend/internal/service/openai_ws_v2_passthrough_image_usage_test.go` | 默认 / Vitest |
| `backend/internal/service/pricing_service_test.go` | 默认 / Vitest |
| `backend/internal/setup/admin_credentials_test.go` | 默认 / Vitest |
| `backend/internal/setup/handler_test.go` | 默认 / Vitest |
| `backend/internal/util/logredact/redact_test.go` | 默认 / Vitest |
| `frontend/src/components/admin/account/__tests__/ScheduledTestsPanel.results.spec.ts` | 默认 / Vitest |
| `frontend/src/components/admin/group/__tests__/GroupRPMOverridesModal.edit.spec.ts` | 默认 / Vitest |
| `frontend/src/components/admin/group/__tests__/GroupRPMOverridesModal.requests.spec.ts` | 默认 / Vitest |
| `frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts` | 默认 / Vitest |
| `frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.requests.spec.ts` | 默认 / Vitest |
| `frontend/src/components/admin/monitor/__tests__/MonitorTemplateManagerDialog.requests.spec.ts` | 默认 / Vitest |
| `frontend/src/components/admin/usage/__tests__/UsageCleanupDialog.spec.ts` | 默认 / Vitest |
| `frontend/src/components/common/__tests__/AnnouncementBell.spec.ts` | 默认 / Vitest |
| `frontend/src/components/user/profile/__tests__/ProfileEditForm.draft.spec.ts` | 默认 / Vitest |
| `frontend/src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts` | 默认 / Vitest |
| `frontend/src/composables/__tests__/useStepUp.spec.ts` | 默认 / Vitest |
| `frontend/src/features/channel-monitor-v2/__tests__/MonitorSettingsPanel.spec.ts` | 默认 / Vitest |
| `frontend/src/stores/__tests__/adminCompliance.reset.spec.ts` | 默认 / Vitest |
| `frontend/src/stores/__tests__/announcements.ownership.spec.ts` | 默认 / Vitest |
| `frontend/src/views/admin/ops/components/__tests__/OpsAlertEventsCard.pagination.spec.ts` | 默认 / Vitest |
| `frontend/src/views/setup/__tests__/SetupWizardView.spec.ts` | 默认 / Vitest |
| `frontend/src/views/user/__tests__/CustomPageView.race.spec.ts` | 默认 / Vitest |
