# 产品需求文档 (PRD): PhantomStream (幻流)

| 项目名称     | PhantomStream (幻流)                |
| :----------- | :---------------------------------- |
| **版本**     | v1.0 (Locked / 已锁定)              |
| **产品类型** | CLI (命令行工具) / 网络安全攻防套件 |
| **开发语言** | Go (Golang)                         |
| **核心场景** | PDF 隐写追踪与反追踪攻防演习        |

## 1. 背景与目标 (Background & Goals)

### 1.1 背景

在数据防泄漏（DLP）与网络攻防演习中，文档追踪是一项核心技术。传统的元数据标记容易被清洗，而复杂的数字水印算法开发成本高。我们需要一个轻量级、基于 Go 语言的演示系统，展示“数据如何隐藏”以及“数据如何被剥离”的博弈过程。

### 1.2 目标

构建两个独立的二进制 CLI 工具：

1.  **Defender (蓝队/防守方)：** 能够在不破坏 PDF 阅读体验的前提下，将一段加密的身份信息（Payload）注入文件，并能随时提取验证。
2.  **Attacker (红队/攻击方)：** 能够识别 PDF 文件是否存在异常数据，并能对其进行“外科手术式”的清除，切断追踪链。

## 2. 系统架构 (System Architecture)

项目采用 **Monorepo** 结构，包含两个独立的 Go Module。

  * **技术原理：** 利用 PDF 文件格式特性。标准 PDF 阅读器只读取到 `%%EOF` (End Of File) 标记为止。蓝队将数据追加在 `%%EOF` 之后；红队通过扫描文件物理末尾来对抗。
  * **加密机制：** 使用 **AES-256-GCM** 对称加密，确保隐写内容即使被发现，没有密钥也无法解读。

### 2.1 工程目录结构 (Project Structure)

```text
phantom-stream/
├── go.work               # Go Workspace 文件 (方便同时管理两个模块)
├── docs/                 # 项目文档
│   ├── PRD.md            # [Locked] 产品需求文档
│   ├── defender/         # Defender 模块详细文档
│   └── attacker/         # Attacker 模块详细文档
├── defender/             # [蓝队] 防守方工具 (注入/验证)
│   ├── go.mod
│   ├── main.go           # CLI 入口
│   └── injector/         # 核心注入逻辑包
│       └── watermark.go
└── attacker/             # [红队] 攻击方工具 (检测/清洗)
    ├── go.mod
    ├── main.go           # CLI 入口
    └── hunter/           # 核心扫描逻辑包
        └── analyzer.go
```

## 3. 功能需求说明 (Functional Requirements)

### 3.1 模块 A：Defender (防守方工具)

**用户画像：** 安全管理员、取证专家。

| 功能 ID  | 功能名称              | 详细描述                                                           | 输入参数                                                   | 预期输出                                                                        |
| :------- | :-------------------- | :----------------------------------------------------------------- | :--------------------------------------------------------- | :------------------------------------------------------------------------------ |
| **D-01** | **注入签名 (Sign)**   | 读取源 PDF，将用户指定的字符串（如员工ID）加密后，隐写至文件末尾。 | `-file` (源路径)<br>`-msg` (追踪信息)<br>`-key` (加密密钥) | 生成新文件 `{filename}_signed.pdf`<br>控制台显示“注入成功”。                    |
| **D-02** | **提取验证 (Verify)** | 尝试从 PDF 中读取隐写数据，解密并验证其完整性。                    | `-file` (目标路径)<br>`-key` (解密密钥)                    | 控制台输出：<br>1. 发现签名<br>2. 解密内容: "EmployeeID:123"<br>3. 校验指纹匹配 |

  * **D-01 关键逻辑：**
      * 生成 Magic Header（如 `0xCA 0xFE 0xBA 0xBE`）以便快速识别。
      * Payload = Magic Header + AES\_Encrypt(msg, key)。
      * NewPDF = OriginalPDF Bytes + `\n` + Payload。

### 3.2 模块 B：Attacker (攻击方工具)

**用户画像：** 渗透测试人员、隐私保护者。

| 功能 ID  | 功能名称             | 详细描述                                                    | 输入参数           | 预期输出                                                                     |
| :------- | :------------------- | :---------------------------------------------------------- | :----------------- | :--------------------------------------------------------------------------- |
| **A-01** | **异常扫描 (Scan)**  | 扫描 PDF 文件结构，计算 `%%EOF` 标记后的冗余字节数。        | `-file` (目标路径) | 控制台输出：<br>1. 文件状态：可疑/干净<br>2. 冗余数据大小：128 Bytes         |
| **A-02** | **强力清洗 (Clean)** | 强制截断文件，移除所有位于最后一个 `%%EOF` 标记之后的数据。 | `-file` (目标路径) | 生成新文件 `{filename}_cleaned.pdf`<br>控制台显示“清洗完成，已移除 X 字节”。 |

  * **A-02 关键逻辑：**
      * 反向读取文件（Reverse Read）。
      * 定位最后一个 `%%EOF` 字符串的字节位置。
      * Buffer = Data[0 : EOF\_Location + 5]。
      * Write Buffer to new file。

## 4. 非功能需求 (Non-Functional Requirements)

1.  **隐蔽性：** 注入后的 PDF 必须能被主流阅读器（Chrome, Acrobat, Preview）正常打开，且无报错弹窗。
2.  **安全性：** 注入的签名必须经过加密，防止攻击者通过 `strings` 等命令直接读取明文。
3.  **健壮性：** 如果文件本身不是 PDF 或已损坏，CLI 应报错并优雅退出，不能 Panic。

## 5. 交互设计 (CLI Usage Design)

为了保持简洁，我们统一使用 Cobra 或标准 flag 库。

**Defender 示例：**

```bash
./defender sign -f contract.pdf -m "User:Alice" -k "MySecretPass"
./defender verify -f contract_signed.pdf -k "MySecretPass"
```

**Attacker 示例：**

```bash
./attacker scan -f contract_signed.pdf
./attacker clean -f contract_signed.pdf
```

## 6. 后续迭代规划 (Roadmap)

  * **Phase 1 (当前):** 基于 EOF 的追加式隐写（实现最快，原理最易懂）。
  * **Phase 2 (未来):** 基于 Stream 对象的隐写（将数据藏在 PDF 内部的图片空隙或字体表中，更难被 `clean` 功能清除）。
