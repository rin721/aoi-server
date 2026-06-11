# 维护指南

维护工作应保持代码、测试、文档和 AI 运行态一致。

## 常规变更流程

1. 阅读 `AGENTS.md` 和相关 `docs/ai` 运行态。
2. 识别本次变更真实影响的代码边界。
3. 更新代码和测试。
4. 更新 `docs` 下的结构化人类文档。
5. 如果运行态变化，更新或修复 `docs/ai` artifact。
6. 根据影响范围运行定向测试和更广测试。
7. 将剩余风险记录到 `docs/backlog/known-gaps.md` 或运行态证据中。

## 文档卫生

人类文档应解释当前代码，而不是把未来想法写成既成事实。未来工作或缺失能力应放入 backlog/known gaps，除非明确标记为计划变更。

优先在结构化目录中增加文档，避免继续增加顶层零散文件。如果旧链接可能被外部引用，可以保留短兼容入口。

## 运行态卫生

`docs/ai` 是运行态系统。当某个 task 或 slice 成为当前工作，它应能通过 current status、task tree、requirement ledger、evidence index
和 handoff 被发现。如果某个 artifact 缺失或太薄，应修复物理 artifact，而不是依赖聊天历史。

## Review 清单

- 变更是否保持目录边界？
- 配置示例和 env 文档是否同步？
- 启动、reload、shutdown 影响是否记录？
- 测试是否靠近它保护的行为？
- 生产风险是否清楚标记？
- AI 运行态 artifact 是否与当前工作状态一致？
