# 复现命令

环境为Go1.27.4（模块自动选择）、Node24.16.0/pnpm，检查使用现有依赖。

后端目录执行全量unit与四包race，测试子进程清除代理变量，CI=true/TZ=UTC。完整输出保留于对应jsonl.gz。

```bash
rtk proxy env -u HTTP_PROXY -u HTTPS_PROXY -u ALL_PROXY -u http_proxy -u https_proxy -u all_proxy CI=true TZ=UTC go test -tags=unit -json ./... -count=1
rtk proxy env -u HTTP_PROXY -u HTTPS_PROXY -u ALL_PROXY -u http_proxy -u https_proxy -u all_proxy CI=true TZ=UTC go test -tags=unit -race -json ./internal/pkg/openai ./internal/pkg/apicompat ./internal/service ./internal/handler -run 'TestGPT62|TestSolUltrafastHTTPAndWSSettlementSQLite|TestGPT61|TestTier2AccountStatsLongContextGate|TestOpenAI.*PricingAt|TestOpenAI.*TurnPricing|TestFastTier' -count=1
rtk proxy env -u HTTP_PROXY -u HTTPS_PROXY -u ALL_PROXY -u http_proxy -u https_proxy -u all_proxy CI=true TZ=UTC go test -tags=unit -race -json ./internal/service -run 'TestGPT62|TestCalculateCost_OpenAIGPT54LongContextMarkerRequiresActualCostIncrease|TestTryModelFilePricing|TestResolveAccountStatsCost|TestTier2AccountStatsLongContextGate|TestOpenAIUsagePricingAt' -count=1
rtk proxy /tmp/sub2api-023-tier2-lint-em_ayj5u/golangci-lint run ./...
rtk proxy env CGO_ENABLED=0 go build -tags embed -o /tmp/sub2api-gpt62-embed ./cmd/server
rtk proxy /tmp/sub2api-gpt62-embed -version
```

前端目录执行已安装pnpm，PATH增补现有Node工具目录。最后一次build在embed构建前完成。

```bash
rtk proxy env PATH=/home/liang/.nvm/versions/node/v24.16.0/bin:$PATH pnpm exec vitest run src/composables/__tests__/useModelWhitelist.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts src/components/admin/channel/__tests__/PricingEntryCard.modelDefaultPrice.spec.ts
rtk proxy env PATH=/home/liang/.nvm/versions/node/v24.16.0/bin:$PATH pnpm run test:run
rtk proxy env PATH=/home/liang/.nvm/versions/node/v24.16.0/bin:$PATH pnpm run typecheck
rtk proxy env PATH=/home/liang/.nvm/versions/node/v24.16.0/bin:$PATH pnpm run lint:check
rtk proxy env PATH=/home/liang/.nvm/versions/node/v24.16.0/bin:$PATH pnpm run build
```

仓库根目录使用已缓存OpenSpec CLI严格校验。

```bash
rtk proxy env PATH=/home/liang/.nvm/versions/node/v24.16.0/bin:$PATH node /home/liang/.npm/_npx/abab5bd700860149/node_modules/@fission-ai/openspec/bin/openspec.js validate port-upstream-0.2.23-gpt62 --strict --json
rtk proxy git diff --check
```
