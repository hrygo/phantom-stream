# 获取帮助

感谢您使用 PhantomStream！如果您遇到问题或需要帮助，有多种途径可以寻求支持。

## 📚 文档资源

在提问之前，请先查阅以下文档：

### 核心文档
- **[用户手册](../defender/USER_MANUAL.md)** - 详细的安装、使用说明和常见问题解答
- **[技术文档](../defender/README.md)** - 技术实现细节和架构设计
- **[攻防演练报告](../docs/PHANTOM_STREAM_JOINT_REPORT.md)** - 红蓝对抗全过程记录

### 开发者文档
- **[贡献指南](../CONTRIBUTING.md)** - 如何为项目做出贡献
- **[行为准则](../CODE_OF_CONDUCT.md)** - 社区行为规范

## 🐛 报告问题

如果您发现了 Bug：

1. **搜索现有 Issue** - 检查是否已有人报告了相同的问题
2. **创建新 Issue** - 使用 [Bug Report 模板](https://github.com/YOUR_USERNAME/phantom-stream/issues/new?template=bug_report.yml)
3. **提供详细信息**：
   - 复现步骤
   - 期望行为 vs 实际行为
   - 环境信息（OS、Go 版本等）
   - 相关日志或错误信息

## 💡 功能建议

如果您有新功能想法：

1. **搜索现有 Issue** - 检查是否已有类似建议
2. **创建 Feature Request** - 使用 [Feature Request 模板](https://github.com/YOUR_USERNAME/phantom-stream/issues/new?template=feature_request.yml)
3. **描述清楚**：
   - 要解决的问题
   - 期望的解决方案
   - 使用场景

## ❓ 提出问题

如果您有使用上的问题：

1. **查阅常见问题** - [用户手册的常见问题部分](../defender/USER_MANUAL.md#常见问题)
2. **创建 Question Issue** - 使用 [Question 模板](https://github.com/YOUR_USERNAME/phantom-stream/issues/new?template=question.yml)
3. **提供上下文**：
   - 您想实现什么
   - 已经尝试了什么
   - 遇到了什么困难

## 🔐 安全问题

**请勿在公开 Issue 中报告安全漏洞！**

如果您发现了安全漏洞，请：

1. 阅读 [安全政策](../SECURITY.md)
2. 通过 [GitHub Security Advisories](https://github.com/YOUR_USERNAME/phantom-stream/security/advisories/new) 私密报告
3. 我们将在 48 小时内回复

## 📧 其他联系方式

### GitHub Discussions
（如果启用了 Discussions 功能）
- 适合：一般性讨论、想法交流、最佳实践分享
- 地址：https://github.com/YOUR_USERNAME/phantom-stream/discussions

### Issue Tracker
- 适合：Bug 报告、功能请求、技术问题
- 地址：https://github.com/YOUR_USERNAME/phantom-stream/issues

## ⏰ 响应时间

我们是开源项目，维护者会尽力及时回复，但请理解：

- **Bug 报告**：通常 1-3 个工作日内回复
- **功能请求**：通常 3-7 个工作日内评估
- **问题咨询**：通常 1-5 个工作日内回复
- **安全问题**：承诺 48 小时内确认收到

## 🤝 社区准则

在寻求帮助时，请：

- ✅ 保持礼貌和尊重
- ✅ 提供清晰的问题描述
- ✅ 附上必要的上下文信息
- ✅ 遵守 [行为准则](../CODE_OF_CONDUCT.md)
- ❌ 避免重复提问已解决的问题
- ❌ 不要在 Issue 中讨论无关话题

## 💪 自助资源

### 常见问题快速链接

1. **如何安装？** → [用户手册 - 安装](../defender/USER_MANUAL.md#安装)
2. **如何使用？** → [用户手册 - 快速开始](../defender/USER_MANUAL.md#快速开始)
3. **中文水印不显示？** → 确保使用 v1.0.0+ 版本（已内嵌 Unicode 字体）
4. **锚点提取失败？** → 检查加密密钥是否正确，使用 `--mode=all` 诊断
5. **如何贡献代码？** → [贡献指南](../CONTRIBUTING.md)

### 调试技巧

```bash
# 查看详细日志
./bin/phantom-guard verify -f document.pdf -k "key" --mode=all

# 检查版本
./bin/phantom-guard version

# 运行测试
cd defender && go test -v ./...
```

---

感谢您的耐心！我们重视每一位用户的反馈和问题。🙏
