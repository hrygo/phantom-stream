# PhantomGuard 用户手册

**版本**: 1.0.0  
**最后更新**: 2025-12-05

---

## ⚠️ 使用前必读

**重要法律声明**：

- ✅ 本工具**仅限于**授权的安全研究、教育培训和合法的文档保护场景
- ❌ **严格禁止**未经授权对他人文档进行追踪或监控
- ⚖️ 使用者必须遵守所在地区的所有法律法规（包括但不限于隐私保护法、数据保护法）
- 📋 使用本工具即表示您已阅读并同意 [LICENSE](../LICENSE) 和 [README](../README.md) 中的免责声明
- 🔒 处理个人信息时，请确保符合 GDPR、CCPA 等数据保护法规要求

**使用者对使用本工具的一切行为及后果承担全部法律责任。**

如有疑问，请在使用前咨询法律顾问。

---

## 目录

1. [概述](#概述)
2. [安装与构建](#安装与构建)
3. [快速开始](#快速开始)
4. [命令参考](#命令参考)
5. [使用场景](#使用场景)
6. [常见问题](#常见问题)
7. [合规使用建议](#合规使用建议)

---

## 概述

PhantomGuard 是一款 PDF 文件水印嵌入与验证工具，支持多锚点签名策略，**仅用于授权的文档追踪与权利追究场景**。

### 核心特性

- **多锚点防护**: 支持 Attachment、SMask、Content、Visual 四种锚点组合
- **加密签名**: 使用 AES-256-GCM 加密，确保签名不可伪造
- **灵活验证**: Auto 模式（快速验证）与 All 模式（完整诊断）
- **交互友好**: 支持命令行与交互式两种使用方式
- **一键验证**: Lookup 功能快速验证上次签名文件

### 系统要求

- 操作系统: macOS, Linux, Windows
- Go 版本: 1.24.0+（仅构建时需要）
- 依赖: 无额外依赖（静态编译二进制）

---

## 安装与构建

### 方式一：使用 Makefile（推荐）

```bash
cd defender
make build
```

生成的二进制文件位于 `defender/bin/phantom-guard`。

### 方式二：手动编译

```bash
cd defender
go build -o bin/phantom-guard .
```

### 安装到系统路径（可选）

```bash
cd defender
make install
```

将二进制安装到 `/usr/local/bin/`，之后可直接运行 `phantom-guard`。

---

## 快速开始

### 1. 签名 PDF 文件

```bash
./bin/phantom-guard sign \
  -f input.pdf \
  -m "UserID:Alice-2025" \
  -k "12345678901234567890123456789012"
```

**参数说明**:
- `-f`: 输入 PDF 文件路径
- `-m`: 要嵌入的消息（追踪信息）
- `-k`: 32 字节加密密钥

**输出**: 
- 签名文件: `input_signed.pdf`（与原文件同目录）
- 默认使用 Maximum 保护级别（所有锚点）

### 2. 验证签名文件

```bash
./bin/phantom-guard verify \
  -f input_signed.pdf \
  -k "12345678901234567890123456789012"
```

**输出**: 
- 验证成功: 显示提取的消息
- 验证失败: 显示错误原因

### 3. 交互式模式

直接运行不带参数启动交互界面:

```bash
./bin/phantom-guard
```

提供三个功能菜单：
1. **Protect PDF**: 签名保护流程
2. **Verify PDF**: 验证流程
3. **Lookup**: 一键验证上次签名文件（All 模式）

---

## 命令参考

### `sign` - 签名命令

**用法**:
```bash
phantom-guard sign -f <file> -m <message> -k <key>
```

**必选参数**:
- `-f, --file`: PDF 文件路径
- `-m, --msg`: 嵌入消息（如 "UserID:Bob"）
- `-k, --key`: 32 字节加密密钥

**示例**:

```bash
# 基础签名
phantom-guard sign -f doc.pdf -m "Confidential-2025" -k "your-32-byte-key-here-12345678"
# 生成: doc_signed.pdf

# 嵌入用户追踪信息
phantom-guard sign -f doc.pdf -m "UserID:Alice" -k "your-key-32bytes-long-string!!"
# 生成: doc_signed.pdf
```

**签名策略**:
- 默认使用 Maximum 级别（Attachment + SMask + Content + Visual）
- 如 PDF 无图像，SMask 自动降级
- Visual 水印为明文震慑，不参与验证

---

### `verify` - 验证命令

**用法**:
```bash
phantom-guard verify -f <file> -k <key> [--mode <mode>]
```

**必选参数**:
- `-f, --file`: 签名 PDF 文件路径
- `-k, --key`: 解密密钥（需与签名时一致）

**可选参数**:
- `--mode`: 验证模式，可选 `auto` 或 `all`（默认 `auto`）

**验证模式**:

#### Auto 模式（默认）
- 尝试所有锚点，遇到第一个成功即停止
- 快速验证，适合日常检查

```bash
phantom-guard verify -f signed.pdf -k "your-key"
```

#### All 模式
- 逐一尝试所有锚点，显示每一步结果
- 完整诊断，适合调试与分析

```bash
phantom-guard verify -f signed.pdf -k "your-key" --mode=all
```

**输出示例（All 模式）**:
```
🔍 Defender Verify Operation
   File: signed.pdf

 - Trying Attachment... OK
   Message(Attachment): UserID:Alice
 - Trying SMask... extract failed
 - Trying Content... OK
   Message(Content): UserID:Alice
✅ Verification finished (mode=all).
```

---

### 交互模式

**启动**:
```bash
phantom-guard
```

**功能菜单**:

#### 1. Protect PDF
引导式签名流程：
- Step 1: 选择 PDF 文件
- Step 2: 输入消息
- Step 3: 输入密钥（可回车自动生成）
- Step 4: 选择保护级别
  - 1: Standard（仅 Attachment）
  - 2: Stealth（Attachment + SMask）
  - 3: Maximum（全部锚点）- **默认**
  - 4: Custom（自定义组合）

**提示**: 回车采用默认 Maximum 保护级别。

#### 2. Verify PDF
引导式验证流程：
- Step 1: 选择签名 PDF 文件
- Step 2: 输入密钥
- Step 3: 选择验证模式
  - 1: Auto（默认）
  - 2: All（逐项展开）

#### 3. Lookup
一键验证上次 Protect 的文件：
- 自动使用上次签名的文件路径与密钥
- 以 All 模式逐项验证所有锚点
- 无需任何输入，快速诊断

**注意**: 首次运行时 Lookup 不可见，需先执行一次 Protect。

---

## 使用场景

### 场景 1: 文档发布追踪

为对外发布的 PDF 嵌入追踪信息，用于溯源泄露责任：

```bash
# 签名
phantom-guard sign -f report.pdf -m "Recipient:Customer-A-20251205" -k "your-secret-key-32bytes-long!!"

# 发布 report_signed.pdf

# 发现泄露后验证
phantom-guard verify -f leaked_copy.pdf -k "your-secret-key-32bytes-long!!"
# 输出: Recipient:Customer-A-20251205
```

### 场景 2: 多版本文档管理

在不同版本文档中嵌入版本信息：

```bash
# 版本 1.0
phantom-guard sign -f contract.pdf -m "Version:1.0-20251205" -k "contract-key-32bytes-long-str!!"
# 生成: contract_signed.pdf

# 版本 2.0（手动重命名输出文件以区分版本）
phantom-guard sign -f contract_v2.pdf -m "Version:2.0-20251210" -k "contract-key-32bytes-long-str!!"
# 生成: contract_v2_signed.pdf

# 验证某版本
phantom-guard verify -f contract_signed.pdf -k "contract-key-32bytes-long-str!!" --mode=all
```

### 场景 3: 交互式快速操作

适合非技术人员或不熟悉命令行场景：

```bash
# 启动交互模式
phantom-guard

# 选择 1: Protect PDF
# 拖拽文件路径，输入消息，回车使用自动生成密钥
# 回车采用默认 Maximum 保护

# 选择 3: Lookup 快速验证
# 无需输入，自动诊断刚签名的文件
```

### 场景 4: 批量验证（脚本集成）

```bash
#!/bin/bash
KEY="shared-secret-key-32bytes-long"

for file in signed_batch/*.pdf; do
  echo "Verifying: $file"
  phantom-guard verify -f "$file" -k "$KEY" || echo "FAILED: $file"
done
```

---

## 常见问题

### Q1: 密钥长度要求？

**A**: 必须是 32 字节（32 个字符）。过短或过长会报错。

示例有效密钥:
- `12345678901234567890123456789012`
- `MySecretKey32BytesLongString!!`

### Q2: 忘记密钥怎么办？

**A**: 无法恢复。签名采用 AES-256-GCM 加密，无密钥无法解密。建议：
- 使用密钥管理工具（如 1Password、KeePass）
- 为项目设置统一密钥并妥善保存

### Q3: Visual 水印可以验证吗？

**A**: 不可以。Visual 为明文震慑水印，不参与自动验证。验证依赖加密锚点（Attachment/SMask/Content）。

### Q4: 签名后文件大小变化？

**A**: 通常略微减小或几乎不变（得益于 pdfcpu 优化）。典型变化：-0.5% ~ +0.1%。

### Q5: All 模式与 Auto 模式区别？

**A**: 
- **Auto**: 找到第一个有效锚点即停止，快速验证
- **All**: 尝试所有锚点并逐项显示结果，用于诊断与调试

### Q6: 纯文本 PDF（无图像）可以签名吗？

**A**: 可以。SMask 锚点会自动降级，使用 Attachment + Content 锚点。

### Q7: 如何查看已签名文件的锚点状态？

**A**: 使用 `--mode=all` 验证模式：

```bash
phantom-guard verify -f signed.pdf -k "your-key" --mode=all
```

会显示每个锚点的提取与解密结果。

### Q8: 交互模式中的 Lookup 看不到？

**A**: Lookup 仅在执行过一次 Protect 后出现。首次运行时菜单只有：
- 1: Protect PDF
- 2: Verify PDF
- 3: Exit

完成一次签名后，Lookup 会自动出现在菜单。

### Q9: 输出文件名规则？

**A**: 
签名后的文件统一使用 `原文件名_signed.pdf` 格式。

示例:
- `report.pdf` → `report_signed.pdf`
- `contract.pdf` → `contract_signed.pdf`

### Q10: 可以修改默认保护级别吗？

**A**: 命令行模式默认使用 Maximum。交互模式可在 Step 4 选择或回车默认 Maximum。

--

## 合规使用建议

为确保合法、合规地使用 PhantomGuard，请遵循以下建议：

### 1. 使用前的准备

#### ✅ 获得授权
- **企业内部使用**：获得组织管理层的书面授权
- **客户文档处理**：在合同中明确约定文档保护机制
- **红蓝对抗演练**：确保所有参与方已签署授权协议
- **个人文档**：仅处理您拥有合法权利的文档

#### 📋 合规性评估
- 评估使用场景是否符合当地法律法规
- 确定是否涉及个人信息处理（需遵守数据保护法）
- 咨询法律顾问，特别是跨境数据传输场景

### 2. 使用中的注意事项

#### 🔒 数据保护
- **加密密钥管理**：
  - 使用强随机密钥（32 字节）
  - 安全存储密钥（密钥管理系统、环境变量等）
  - 定期更换密钥
  - 不在代码或日志中硬编码密钥

- **个人信息处理**：
  - 嵌入消息中避免包含敏感个人信息（如身份证号、电话号码）
  - 遵守最小化原则，仅嵌入必要的追踪标识
  - 保留数据处理记录，满足合规审计要求

#### 📢 透明度要求
- **Visual 水印**：
  - 明文水印已起到告知作用，但仍建议额外说明
  - 在文档分发前，告知接收者追踪机制的存在
  - 在隐私政策或使用条款中披露文档保护措施

- **隐藏锚点**：
  - 虽然锚点隐藏，但法律上仍需满足透明度要求
  - 建议在文档首页或元数据中说明"本文档包含防护追踪技术"

### 3. 合法使用场景示例

#### ✅ 推荐场景

1. **企业内部文档管理**
   ```bash
   # 场景：企业内部敏感文档分发追踪
   # 前提：已获得企业授权，员工知晓文档保护政策
   phantom-guard sign -f confidential_report.pdf -m "EmployeeID:E12345" -k "${COMPANY_KEY}"
   ```

2. **安全培训与演练**
   ```bash
   # 场景：红蓝对抗演练中的文档追踪测试
   # 前提：所有参与方已签署保密协议
   phantom-guard sign -f redteam_doc.pdf -m "Exercise:2025-R3" -k "${DRILL_KEY}"
   ```

3. **学术研究**
   ```bash
   # 场景：研究 PDF 隐写技术
   # 前提：仅用于实验环境，不涉及他人隐私
   phantom-guard sign -f research_sample.pdf -m "Experiment:001" -k "${RESEARCH_KEY}"
   ```

4. **版权保护**
   ```bash
   # 场景：原创文档的版权声明嵌入
   # 前提：对文档拥有合法版权
   phantom-guard sign -f my_ebook.pdf -m "Copyright:Author2025" -k "${COPYRIGHT_KEY}"
   ```

#### ❌ 禁止场景

1. **未经授权的监控**
   - ❌ 在他人不知情的情况下追踪文档流转
   - ❌ 未获授权处理客户或用户的文档

2. **违反隐私法规**
   - ❌ 处理个人信息但未履行告知义务
   - ❌ 跨境传输数据但未遵守当地法规

3. **恶意用途**
   - ❌ 用于敲诈勒索或其他犯罪活动
   - ❌ 用于侵犯他人合法权益

### 4. 合规检查清单

使用前请确认：

- [ ] 我有权处理这些 PDF 文档
- [ ] 我已获得组织/客户的明确授权
- [ ] 如涉及个人信息，我已履行告知义务
- [ ] 我了解所在地区的数据保护法规要求
- [ ] 我已安全管理加密密钥
- [ ] 我已在隐私政策/使用条款中披露追踪机制
- [ ] 我已咨询法律顾问（如有必要）

### 5. 数据保护法规参考

#### 欧盟 GDPR（通用数据保护条例）
- 适用范围：处理欧盟居民的个人数据
- 关键要求：合法基础、透明度、最小化、安全保障
- 参考：https://gdpr.eu/

#### 美国 CCPA（加州消费者隐私法）
- 适用范围：处理加州居民的个人信息
- 关键要求：告知、选择退出、数据删除权
- 参考：https://oag.ca.gov/privacy/ccpa

#### 中国《个人信息保护法》
- 适用范围：在中国境内处理个人信息
- 关键要求：知情同意、最小必要、安全保护
- 参考：http://www.npc.gov.cn/

### 6. 遇到法律问题时

如果您在使用过程中遇到法律相关问题：

1. **立即停止使用**相关功能
2. **咨询专业法律顾问**
3. **保留相关记录**（授权文件、日志等）
4. **通知相关方**（如有数据泄露风险）

### 7. 免责声明重申

⚠️ **重要**：
- 本建议仅供参考，不构成法律意见
- 使用者需自行承担使用本工具的一切法律责任
- 作者和贡献者不对任何法律后果负责
- 建议在实际使用前咨询当地专业法律顾问

---

## 技术支持

- **项目文档**: `defender/docs/`
- **技术附录**: `defender/docs/Defender_Official_Report_Attr.md`
- **测试用例**: `defender/injector/*_test.go`

---

**版权所有 © 2025 PhantomStream Defense Team**
