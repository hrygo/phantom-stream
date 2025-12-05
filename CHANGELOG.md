# 变更日志

所有重要的项目变更都将记录在此文件中。

格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [1.0.0] - 2025-12-05

### ✨ 新增

- **多锚点防御体系**
  - Attachment 锚点：伪装为合法 PDF 附件
  - SMask 锚点：隐藏在图像软蒙版数据中
  - Content 锚点：嵌入页面内容流的不可见字符
  - Visual 锚点：可见水印，明文展示追踪信息

- **完整 Unicode 支持**
  - 内嵌 Go Noto Universal 字体（14MB）
  - 支持全球所有语言字符（CJK、阿拉伯文、西里尔文等）
  - 跨平台一致性，无需依赖系统字体

- **加密与验证**
  - AES-256-GCM 加密保护追踪信息
  - 支持自定义加密密钥（32 字节）
  - 自动验证与解密提取

- **灵活的验证模式**
  - Auto 模式：首个成功锚点即停止（快速验证）
  - All 模式：逐一验证所有锚点（完整诊断）

- **CLI 工具**
  - `phantom-guard sign` - 嵌入追踪信息
  - `phantom-guard verify` - 验证追踪信息
  - `phantom-guard lookup` - 交互式菜单

### 🐛 修复

- 修复 Visual 水印仅显示 "CONFIDENTIAL" 的问题（改用 `|` 分隔符）
- 修复中文字符无法显示的问题（嵌入 Unicode 字体）
- 修复 SMask 锚点提取失败问题（Magic Header 处理）
- 修复字体名称不匹配问题（`GoNotoCurrent-Regular-Regular`）

### 📝 文档

- 添加项目总览 README
- 添加攻防演练联合报告
- 添加 Defender 技术文档
- 添加用户使用手册
- 添加 Phase 1-9 技术演进文档

### 🔧 优化

- 简化 Unicode 检测逻辑（移除冗余判断）
- 统一使用嵌入字体（移除系统字体探测）
- 添加详细代码注释（字体名称由来说明）
- 清理冗余目录（embedded_fonts）

### 🎯 红蓝对抗验证

- ✅ 通过 9 轮红蓝对抗测试
- ✅ 抵御常规清洗工具（PDFtk、Ghostscript 等）
- ✅ Visual 水印震慑效果显著
- ✅ 多锚点冗余保证容错性

---

## [未发布]

### 计划中的功能

- [ ] 支持批量处理 PDF 文件
- [ ] 添加 Web UI 界面
- [ ] 支持自定义水印位置和样式
- [ ] 优化大文件处理性能
- [ ] 添加更多锚点类型

### 已知问题

- 无法抵御彻底重建 PDF 结构的深度清洗
- Visual 水印旋转角度固定为 45 度
- 部分非标准 PDF 可能导致锚点嵌入失败

---

## 版本说明

- **[1.0.0]** - 稳定版本，已通过红蓝对抗验证
- **[未发布]** - 开发中的功能和改进

---

**格式说明**：
- ✨ 新增：新功能
- 🐛 修复：Bug 修复
- 📝 文档：文档更新
- 🔧 优化：性能优化或代码改进
- 🎯 验证：测试和验证相关

---

[1.0.0]: https://github.com/YOUR_USERNAME/phantom-stream/releases/tag/v1.0.0
[未发布]: https://github.com/YOUR_USERNAME/phantom-stream/compare/v1.0.0...HEAD
