# S14 分配量对比

实施中使用来源 `BenchmarkStripOpenAIResponsesInputNamespaces`，命令：

```text
go test ./internal/service -run '^$' -bench '^BenchmarkStripOpenAIResponsesInputNamespaces/(tools_only|nested_only|keep_calls)' -benchtime=3x -benchmem
```

同机实测摘录（Linux amd64 / i9-13900H，3 次采样；不以短采样时延宣称固定加速倍数）：

| 无修改输入 | 改前 B/op | 改后 B/op | 改前/改后 allocs/op |
|---|---:|---:|---:|
| tools_only 1 MiB | 3,170,304 | 1,056,773 | 3 / 1 |
| tools_only 8 MiB | 25,190,400 | 8,396,800 | 3 / 1 |
| nested_only 8 MiB | 25,190,448 | 8,396,800 | 4 / 1 |
| keep_calls 8 MiB | 25,190,464 | 8,396,800 | 4 / 1 |

约减少三分之二的分配字节。仍有 gjson 读取正文的分配，不宣称零分配。存在 namespace 修改时才分配重建缓冲；字节保真与错误回退由 TestStripOpenAIResponsesInputNamespaces 系列覆盖。
