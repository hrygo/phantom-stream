# 🛡️ Defender - PDF 水印嵌入与验证工具

**执行摘要**: Defender 是 PhantomStream 攻防演习系统的防守方工具，旨在 PDF 文件中嵌入加密的追踪信息，并提供强大的验证机制。经过多轮攻防演习，Defender 已从最初的单锚点防御（Phase 5）升级至 **Phase 7.1 双轨验证（附件 + SMask）**，并通过 **Phase 7.2 架构重构**显著提升了系统的可扩展性和维护性。目前，Defender 能够有效抵御红队已知的各种攻击手段，提供高隐蔽性和强韧性的追踪信息保护。

## 📋 目录

- [执行摘要](#执行摘要)
- [特性](#特性)
- [技术原理：双锚点防御](#技术原理：双锚点防御)
- [安装](#安装)
- [快速开始](#快速开始)
- [使用指南](#使用指南)
- [架构设计](#架构设计)
  - [设计原则](#设计原则)
  - [模块结构](#模块结构)
  - [核心组件](#核心组件)
  - [工作流程](#工作流程)
  - [扩展性示例](#扩展性示例)
  - [性能指标](#性能指标)
  - [安全特性](#安全特性)
  - [测试覆盖](#测试覆盖)
  - [向后兼容](#向后兼容)
  - [未来规划](#未来规划)
- [安全性：深入解析](#安全性：深入解析)
- [攻防演习历史](#攻防演习历史)
- [FAQ](#faq)
- [许可证](#许可证)
- [贡献](#贡献)

## ✨ 特性

- **🔐 强加密保护**：使用 AES-256-GCM 加密算法，确保追踪信息安全
- **🔗 双轨验证**：引入附件和图像 SMask 双锚点，显著提升签名韧性
- **🕵️‍♂️ 极高隐蔽性**：SMask 锚点利用图像透明蒙版，不易被察觉
- **🛡️ 抗清洗攻击**：有效抵御红队精准流清洗等多种攻击手段
- **✅ 易于验证**：支持多锚点容错验证，任一锚点存活即可恢复信息
- **🔄 架构灵活**： Phase 7.2 架构重构支持轻松扩展新锚点类型
- **📦 零依赖部署**：编译为单个二进制文件，无需额外依赖

## 🔬 技术原理：双锚点防御

Defender 采用 **Phase 7.1 双轨验证方案**，通过同时嵌入两个独立的追踪锚点来提升防御韧性：

1.  **主锚点：附件 (Attachment)**  
    - 将加密追踪信息封装为 PDF 标准附件（`font_license.txt`），挂载在文档引用树上。
    - **特点**：合法性高，兼容性强，但易被检测。

2.  **隐蔽锚点：图像软蒙版 (SMask)**  
    - 将备份追踪信息嵌入到 PDF 图像的透明度蒙版（Soft Mask）数据中。
    - **特点**：极高隐蔽性，对视觉无影响，难以被常规工具检测和清除。

**双轨验证机制：**

-   签名时，两个锚点同时注入。  
-   验证时，任一锚点成功提取并验证即可恢复追踪信息（OR 逻辑）。

**核心优势：**
1.  **高韧性**：红队必须同时发现并清除两个独立锚点才能使签名完全失效。
2.  **高隐蔽性**：SMask 锚点利用 PDF 特殊结构，极难被发现。
3.  **符合标准**：所有锚点均利用 PDF ISO 32000 标准特性，不影响文件阅读。
4.  **抗破坏性**：任何尝试清除 SMask 锚点的行为都可能导致图像显示异常，从而提升红队的攻击成本。

## 🚀 安装

### 从源码编译

```bash
# 克隆项目
cd defender

# 编译
go build -o defender

# 验证安装
./defender --version
```

### 系统要求

- Go 1.24.0 或更高版本
- 支持 macOS、Linux、Windows

## 🎯 快速开始

### 1. 签名 PDF 文件 (双锚点)

为 PDF 文件嵌入追踪信息，默认会尝试嵌入附件和 SMask 双锚点：

```bash
./defender sign \
  -f /path/to/document.pdf \
  -m "UserID:12345" \
  -k "12345678901234567890123456789012"
```

**参数说明：**
- `-f, --file`: 源 PDF 文件路径
- `-m, --msg`: 要嵌入的追踪信息（如员工 ID、追踪码等）
- `-k, --key`: 32 字节加密密钥

**输出示例：**
```
✓ Anchor 1/2: Attachment embedded (54 bytes)
✓ Anchor 2/2: SMask embedded
✓ Signature mode: Dual-anchor (Attachment + SMask)
✅ Successfully signed PDF: document_signed.pdf
```
**注意**：如果 PDF 文件不包含图像，SMask 锚点将自动降级为单锚点模式（仅附件）。

### 2. 验证追踪信息 (多锚点容错)

从签名的 PDF 文件中提取并验证追踪信息。只要任一锚点有效，即可成功验证：

```bash
./defender verify \
  -f /path/to/document_signed.pdf \
  -k "12345678901234567890123456789012"
```

**输出示例：**
```
🔍 Defender Verify Operation
   File: document_signed.pdf

✓ Verified via Anchor 1: Attachment
✅ Verification successful!
📋 Extracted message: "UserID:12345"
```

**如果附件被清除，SMask 锚点仍可作为备份进行验证：**
```
🔍 Defender Verify Operation
   File: document_signed_noattach.pdf

✓ Verified via Anchor 2: SMask (backup anchor activated)
✅ Verification successful!
📋 Extracted message: "UserID:12345"
```

## 📖 使用指南

### 签名命令详解

```bash
defender sign [flags]

Flags:
  -f, --file string   源 PDF 文件路径 (必填)
  -m, --msg string    要嵌入的追踪信息 (必填)
  -k, --key string    32 字节加密密钥 (必填)
  -h, --help          显示帮助信息
```

**使用示例：**

```bash
# 基础用法
./defender sign -f report.pdf -m "Employee:Alice" -k "your-32-byte-secret-key-here!!"

# 嵌入复杂信息
./defender sign -f contract.pdf -m "TrackID:ABC-2024-001|Dept:Sales" -k "your-32-byte-secret-key-here!!"
```

### 验证命令详解

```bash
defender verify [flags]

Flags:
  -f, --file string   目标 PDF 文件路径 (必填)
  -k, --key string    32 字节解密密钥 (必填)
  -h, --help          显示帮助信息
```

**可能的错误：**

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| `file does not exist` | 文件不存在 | 检查文件路径 |
| `encryption key must be 32 bytes long` | 密钥长度错误 | 确保密钥正好 32 字节 |
| `attachment not found` | 文件未被签名 | 使用正确的签名文件 |
| `decryption failed` | 密钥错误或数据损坏 | 使用正确的密钥 |

### 密钥生成建议

**方法 1：使用随机字符串**
```bash
# 生成 32 字节随机密钥（Base64）
openssl rand -base64 32 | head -c 32
```

**方法 2：使用密码短语**
```bash
# 从密码短语派生密钥
echo -n "your-passphrase" | openssl dgst -sha256 -binary | base64 | head -c 32
```

**重要提示：**
- ⚠️ 密钥必须安全保存，丢失后无法恢复追踪信息
- ⚠️ 不同文件应使用不同密钥以提高安全性
- ⚠️ 密钥不应在代码或配置文件中明文存储

## 🏗️ 架构设计

### 设计原则

1. **分层架构**：职责清晰分离（加密层、锚点层、验证层）
2. **接口驱动**：使用 `Anchor` 接口支持多种隐写技术
3. **开放扩展**：通过 `AnchorRegistry` 轻松添加新的锚点类型
4. **向后兼容**：保留旧函数作为 Deprecated 包装器

### 模块结构

```
injector/
├── watermark.go            # 主入口 - Sign/Verify 公共 API (180 行)
├── crypto.go              # 加密/解密模块 (89 行)
├── validation.go          # 输入验证和路径处理 (66 行)
├── anchor.go              # 锚点接口定义和注册表 (57 行)
├── anchor_attachment.go   # 附件锚点实现 (85 行)
├── anchor_smask.go        # SMask 锚点实现 (410 行)
├── phase7_test.go         # Phase 7 集成测试 (347 行)
└── watermark_test.go      # 单元测试 (387 行)
```

**总代码量**: 1621 行 (vs 原来 ~768 行，重构后增加了模块化和可扩展性)

### 核心组件

#### 1. CryptoManager (crypto.go)

**职责**: 加密和解密操作

```go
type CryptoManager struct {
    key []byte
}

// 核心方法
func (c *CryptoManager) Encrypt(message string) ([]byte, error)
func (c *CryptoManager) Decrypt(payload []byte) (string, error)
```

**特性**:
- AES-256-GCM 认证加密
- Magic Header (0xCA 0xFE 0xBA 0xBE) 用于校验
- 12 字节随机 Nonce
- Payload 格式: `MagicHeader + Nonce + EncryptedData`

### 2. Anchor 接口 (anchor.go)

**职责**: 定义隐写锚点的统一接口

```go
type Anchor interface {
    Name() string
    Inject(inputPath, outputPath string, payload []byte) error
    Extract(filePath string) ([]byte, error)
    IsAvailable(ctx *model.Context) bool
}
```

**优势**:
- 统一不同隐写技术的操作
- 支持运行时检测锚点可用性
- 易于添加新的隐写方法

### 3. AnchorRegistry (anchor.go)

**职责**: 管理和注册锚点实现

```go
type AnchorRegistry struct {
    anchors []Anchor
}

func NewAnchorRegistry() *AnchorRegistry  // 默认注册 Attachment + SMask
func (r *AnchorRegistry) AddAnchor(anchor Anchor)  // 添加自定义锚点
```

### 4. AttachmentAnchor (anchor_attachment.go)

**技术**: PDF 附件隐写

**特点**:
- 标准 PDF 特性，兼容性强
- 易于检测和移除（主锚点）
- 始终可用（`IsAvailable` 返回 true）

**实现细节**:
- 附件名称: `font_license.txt`（伪装成字体许可证）
- 使用 pdfcpu 的 `AddAttachmentsFile` API

### 5. SMaskAnchor (anchor_smask.go)

**技术**: 图像软蒙版（Soft Mask）隐写

**特点**:
- 高隐蔽性（备份锚点）
- 需要 PDF 中至少有一张图像
- 数据嵌入在蒙版末尾，对视觉无影响

**实现细节**:
- 扫描 xRefTable 查找图像对象
- 创建透明蒙版（全 255 = 完全不透明）
- Payload 嵌入位置: 蒙版数据末尾
- 压缩: FlateDecode (zlib)
- 提取策略: 从末尾 500 bytes 扫描 Magic Header

**关键修复** (Phase 7.1):
1. 图像查找: xRefTable 全局扫描
2. 对象持久化: 重建 StreamDict（解决值类型问题）
3. Filter 声明: 显式添加 `FlateDecode`
4. 数据源: 使用 `Raw`（压缩）而非 `Content`（未压缩）
5. Payload 定位: Magic Header 扫描（取代固定大小）

### 6. Validation (validation.go)

**职责**: 输入验证和路径处理

**函数**:
- `validateInputs()`: 签名参数验证
- `validateVerifyInputs()`: 验证参数验证
- `generateOutputPath()`: 生成输出文件路径

## 🔄 工作流程

### Sign 签名流程

```
1. validateInputs()        → 验证输入参数
2. CryptoManager.Encrypt() → 加密消息
3. generateOutputPath()    → 生成临时和最终路径
4. AttachmentAnchor.Inject() → 嵌入附件锚点（主锚点）
5. SMaskAnchor.Inject()    → 嵌入 SMask 锚点（备份锚点）
   - 成功: 双轨签名
   - 失败: 降级为单锚点（仅附件）
```

### Verify 验证流程

```
1. validateVerifyInputs()  → 验证输入参数
2. AnchorRegistry.GetAvailableAnchors() → 获取锚点列表
3. 遍历锚点:
   a. Anchor.Extract()      → 提取 Payload
   b. CryptoManager.Decrypt() → 解密和验证
   c. 成功 → 返回消息
4. 所有锚点失败 → 返回错误
```

**容错设计**: 任一锚点验证成功即可（OR 逻辑）

## 🎯 扩展性示例

### 添加新的锚点类型

```go
// 1. 实现 Anchor 接口
type MetadataAnchor struct{}

func (a *MetadataAnchor) Name() string { return "Metadata" }
func (a *MetadataAnchor) Inject(inputPath, outputPath string, payload []byte) error { ... }
func (a *MetadataAnchor) Extract(filePath string) ([]byte, error) { ... }
func (a *MetadataAnchor) IsAvailable(ctx *model.Context) bool { ... }

// 2. 注册到 Registry
registry := NewAnchorRegistry()
registry.AddAnchor(&MetadataAnchor{})

// 3. 自动参与 Sign/Verify 流程（无需修改主逻辑）
```

## 📊 性能指标

| 操作 | 平均耗时 | 文件大小影响 |
|-----|---------|------------|
| 签名 (Dual-Anchor) | ~40ms | -0.65% (轻微减小) |
| 验证 (Attachment) | <1ms | - |
| 验证 (SMask) | ~10ms | - |

**注**: 文件大小减小是因为 pdfcpu 优化了 PDF 结构

## 🔒 安全性：深入解析

### 加密算法与特性

- **算法**：AES-256-GCM (Galois/Counter Mode)，军事级加密强度。
- **密钥长度**：256 位 (32 字节)，确保高安全性。
- **认证加密 (AEAD)**：内置完整性验证，防止任何对追踪信息的篡改。
- **随机 Nonce**：每次签名生成唯一的随机数，增加加密的不可预测性。
- **Magic Header**：快速识别和验证 Payload 的完整性。

### Defender 的多层安全机制 (Phase 7.1)

✅ **机密性**：即使攻击者获取文件，没有密钥也无法读取追踪信息。
✅ **完整性**：GCM 模式自动验证数据完整性，防止篡改。
✅ **隐蔽性**：通过附件和 SMask 双锚点，将追踪信息隐蔽地嵌入到 PDF 文件中。
✅ **韧性**：双锚点防御显著提高了红队清除追踪信息的难度和成本。

### 安全最佳实践

1. **密钥管理**
   - 使用强随机密钥（至少 256 位熵）。
   - 安全存储密钥（使用密钥管理系统或环境变量），绝不硬编码。
   - 定期轮换密钥。

2. **文件保护**
   - 原始文件和签名文件分开存储。
   - 限制签名文件的访问权限。
   - 记录所有签名和验证操作，进行安全审计。

3. **操作安全**
   - 不在命令历史中暴露密钥（使用环境变量）。
   - 验证后立即清除临时文件。
   - 监控异常的验证失败，警惕潜在攻击尝试。


## 🎮 攻防演习历史

Defender 在 PhantomStream 攻防演习中经历了多个阶段的进化：

| 阶段 | 策略 | 红队攻击 | 结果 |
|------|------|---------|------|
| Phase 1-4 | 物理/结构层防御 | 截断 / 间隙覆盖 / 版本回滚 / 图谱修剪 | ❌ 失败 |
| Phase 5 | 嵌入式附件 (单锚点) | 语义分析 (成功检测，但无法无损清除) | ✅ 阶段性成功 |
| Phase 6 | 流内容清洗突破 | 精准流清洗 (保持字节长度替换内容) | ❌ 失败 (被突破) |
| **Phase 7.1** | **双轨验证 (附件 + SMask)** | 深度探测 / 清除附件 (SMask 仍有效) | ✅ **成功防御** |
| **Phase 7.2** | **架构重构** | (内部优化) | ✅ 提升可扩展性 |

**Phase 7.1 成功的关键：**
- 🎯 引入双锚点策略，显著提升攻击成本。
- 🌳 SMask 锚点利用图像结构，实现高隐蔽性。
- 🔄 架构重构支持持续扩展和维护。

详细演习报告：[/docs/TOTAL_REPORTv1.0.md](../docs/TOTAL_REPORTv1.0.md)

## ❓ FAQ

### Q1: 签名后的 PDF 文件会变大吗？

A: 不会显著变大。由于 pdfcpu 在添加附件时会优化文件结构，实际上文件可能反而略微减小。追踪信息本身通常只有几十字节。

### Q2: 签名后的 PDF 是否还能正常阅读？

A: 完全可以。所有主流 PDF 阅读器（Adobe Acrobat、Chrome、Preview 等）都能正常打开和阅读签名后的文件。

### Q3: 用户能看到附件吗？

A: 如果用户在 PDF 阅读器中打开"附件"面板，会看到一个名为 `font_license.txt` 的文件。SMask 锚点则在视觉上完全不可见。

### Q4: 如果忘记密钥怎么办？

A: 无法恢复。AES-256-GCM 是强加密算法，没有密钥就无法解密。因此密钥管理至关重要。

**建议：** 使用密钥管理系统（如 HashiCorp Vault）或安全地记录密钥。

### Q5: 能否对已签名的文件再次签名？

A: 可以，但会生成新的锚点。由于默认附件名相同，可能会覆盖原有追踪信息。SMask 锚点则会尝试寻找下一个可用图像进行注入。

**建议：** 使用不同的附件名或在消息中包含版本信息，或考虑禁用重复签名。

### Q6: 红队能否清除追踪信息？

A: 红队必须同时发现并清除所有锚点才能完全失效签名。Phase 7.1 引入的 SMask 锚点具有极高隐蔽性，显著提升了红队的清除难度和成本。任何尝试清除 SMask 锚点的行为都可能导致图像显示异常，从而提升红队的攻击成本。

### Q7: 密钥长度为什么必须是 32 字节？

A: AES-256 算法要求密钥长度为 256 位（32 字节）。这是算法规范的要求，确保加密强度。

### Q8: 能否批量处理多个文件？

A: 当前版本不支持。但可以通过 shell 脚本实现：

```bash
#!/bin/bash
KEY="your-32-byte-secret-key-here!!"

for file in *.pdf; do
  ./defender sign -f "$file" -m "Batch:$(date +%s)" -k "$KEY"
done
```

## 🔮 未来改进方向

### Phase 8 候选特性

1. **XMP Metadata Anchor**：利用 PDF XMP 元数据
2. **Page Annotation Anchor**：利用注释对象
3. **Transparency Group Anchor**：利用透明度组
4. **Font Subsetting Anchor**：利用字体子集化

### 架构改进

- [ ] 并行锚点注入（目前是顺序）
- [ ] 锚点优先级配置
- [ ] 自定义锚点选择策略
- [ ] 签名元数据（版本、时间戳）
- [ ] 锚点健康度监控（成功率统计）

## 📜 许可证

本项目是 PhantomStream 攻防演习系统的一部分，仅用于教育和研究目的。

## 🤝 贡献

如有问题或建议，请提交 Issue 或 Pull Request。

---

**Defender - 守护您的数字资产！** 🛡️✨