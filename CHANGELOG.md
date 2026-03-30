# CHANGELOG - Lobster Orchestrator

所有重要变更将记录在此文件中。

---

## [V0.2.0] - 2026-03-30 18:00 UTC

### Added
- 日志系统 (infoLog/errorLog/warnLog)
- 实例 LastError 字段记录最后错误
- Stop() 5 秒超时等待逻辑

### Changed
- 使用标准 log 包替代 fmt.Printf
- 改进日志输出格式 (带 emoji)
- 优化错误处理流程

### Fixed
- 进程清理不彻底的问题
- 错误信息未记录的问题

---

## [V0.1.5] - 2026-03-30 17:45 UTC

### Added
- docs/BEST_PRACTICES.md - 最佳实践文档
- 不死龙虾哲学章节

---

## [V0.1.4] - 2026-03-30 17:30 UTC

### Added
- docs/TROUBLESHOOTING.md - 故障排查指南
- 8 个常见问题及解决方案
- 调试工具使用说明

---

## [V0.1.3] - 2026-03-30 17:15 UTC

### Added
- docs/CODE_REVIEW.md - 代码审查报告
- 识别 5 个需改进点
- 提出 3 个待添加功能

---

## [V0.1.2] - 2026-03-30 17:00 UTC

### Added
- docs/API.md - 完整 API 文档
- cURL/JavaScript/Python 示例
- 错误处理说明

---

## [V0.1.1] - 2026-03-30 16:45 UTC

### Added
- scripts/deploy-termux.sh - Termux 部署脚本
- scripts/stress-test.sh - 50 实例压力测试
- scripts/generate-config.sh - 配置生成器
- scripts/monitor.sh - 性能监控脚本
- docs/TEST_PLAN.md - 完整测试计划

---

## [V0.1.0] - 2026-03-30 16:30 UTC

### Added
- 核心功能实现
  - 实例管理 (启动/停止/重启)
  - Web Dashboard (实时监控)
  - RESTful API
  - 健康监控 (自动重启)
  - YAML 配置
- 基础文档
  - README.md
  - DESIGN.md

---

## 🦞 版本命名规则

- **V0.x.x**: 初始开发阶段
- **V1.0.0**: 首个稳定版本 (通过 50 实例测试)
- **V2.0.0**: 生产就绪版本 (通过 24h 稳定性测试)

---

**🦞 不死龙虾，持续进化！**
