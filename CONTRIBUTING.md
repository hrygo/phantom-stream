# 贡献指南

感谢您对 PhantomStream 项目的关注！我们欢迎任何形式的贡献。

## 贡献方式

### 1. 报告问题 (Issues)

如果您发现了 Bug 或有功能建议，请：

- 使用 GitHub Issues 提交
- 使用适当的 Issue 模板
- 提供详细的复现步骤或使用场景
- 附上环境信息（操作系统、Go 版本等）

### 2. 提交代码 (Pull Requests)

#### 开发流程

1. **Fork 本仓库**
   ```bash
   git clone https://github.com/YOUR_USERNAME/phantom-stream.git
   cd phantom-stream
   ```

2. **创建功能分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **进行开发**
   - 遵循项目代码规范
   - 编写必要的测试
   - 更新相关文档

4. **提交代码**
   ```bash
   git add .
   git commit -m "feat: 添加新功能描述"
   ```

5. **推送到您的仓库**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **创建 Pull Request**
   - 详细描述您的修改
   - 引用相关的 Issue
   - 等待代码审查

#### Commit 信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

- `feat:` 新功能
- `fix:` 修复 Bug
- `docs:` 文档更新
- `style:` 代码格式调整（不影响功能）
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建/工具配置

示例：
```
feat: 添加 Visual 水印中文支持
fix: 修复 SMask 锚点提取失败问题
docs: 更新用户手册中的示例
```

### 3. 代码规范

#### Go 代码

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指南
- 为导出的函数和类型添加文档注释
- 保持函数简洁（建议不超过 50 行）

#### 测试要求

- 新功能必须包含单元测试
- 测试覆盖率应保持在 70% 以上
- 使用有意义的测试用例名称

```bash
cd defender
go test -v ./...
go test -cover ./...
```

### 4. 文档贡献

文档同样重要！您可以：

- 修正错别字或不清晰的表述
- 添加使用示例
- 翻译文档到其他语言
- 改进 README 和用户手册

## 开发环境配置

### 系统要求

- Go 1.24 或更高版本
- Make（可选，用于构建脚本）
- Git

### 构建项目

```bash
cd defender
make build
```

### 运行测试

```bash
cd defender
make test
```

## 问题排查

如果遇到开发问题：

1. 查看 [用户手册](defender/USER_MANUAL.md) 的常见问题部分
2. 搜索现有的 Issues
3. 在 Issue 中提问

## 行为准则

请阅读并遵守我们的 [行为准则](CODE_OF_CONDUCT.md)。

## 许可协议

提交代码即表示您同意将您的贡献以 MIT 协议授权。

---

再次感谢您的贡献！🎉
