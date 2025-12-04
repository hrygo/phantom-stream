# Defender Injector 架构文档

## 📋 概述

本模块实现了 PDF 隐写签名系统，采用双轨验证（Dual-Anchor）策略，通过多种隐写技术在 PDF 中嵌入和验证加密消息。

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

## 🔑 核心组件

### 1. CryptoManager (crypto.go)

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

## 🔒 安全特性

1. **AES-256-GCM**: 军事级加密强度
2. **AEAD**: 认证加密，防篡改
3. **随机 Nonce**: 每次签名唯一
4. **Magic Header**: 快速验证 Payload 完整性
5. **双轨验证**: 容错和冗余

## 🧪 测试覆盖

- ✅ 单元测试: 加密、解密、验证、路径生成
- ✅ 集成测试: 双锚点签名/验证、单锚点降级
- ✅ 容错测试: 附件移除后 SMask 验证
- ✅ 性能测试: 基准测试（Benchmark）

**测试通过率**: 100% (17 个测试，0 失败)

## 📝 向后兼容

旧代码仍可工作，已标记为 Deprecated：

```go
// Deprecated: Use CryptoManager.Encrypt instead
func createEncryptedPayload(message string, key []byte) ([]byte, error)

// Deprecated: Use AttachmentAnchor.Extract instead
func extractPayloadFromPDF(filePath string) ([]byte, error)

// Deprecated: Use CryptoManager.Decrypt instead
func decryptPayload(payload, key []byte) (string, error)
```

**迁移建议**: 新代码使用新 API，旧测试暂时保留包装器

## 🚀 未来规划

### Phase 8 候选特性

1. **XMP Metadata Anchor**: 利用 PDF XMP 元数据
2. **Page Annotation Anchor**: 利用注释对象
3. **Transparency Group Anchor**: 利用透明度组
4. **Font Subsetting Anchor**: 利用字体子集化

### 架构改进

- [ ] 并行锚点注入（目前是顺序）
- [ ] 锚点优先级配置
- [ ] 自定义锚点选择策略
- [ ] 签名元数据（版本、时间戳）

## 📚 参考文档

- PDF 规范: ISO 32000-2:2020
- pdfcpu 文档: https://pdfcpu.io/
- AES-GCM: NIST SP 800-38D

---

**最后更新**: 2025-12-05  
**版本**: Phase 7.1 (Dual-Anchor Refactored)
