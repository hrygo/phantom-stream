# PhantomStream 攻防演习 - Defender 技术附录 (Branch/Attributes)

**关联主文档**: `defender/docs/Defender_Official_Report_Main.md`  
**内容说明**: 本文档收录演习过程中的详细技术指标、日志片段、配置参数及具体验证数据。

---

## 1. 关键技术参数 (Technical Specifications)

### 1.1 Payload 结构 (Phase 5+)
隐写载荷采用统一的二进制结构，嵌入到不同载体（附件、SMask）中：
*   **加密算法**: AES-256-GCM
*   **结构**:
    ```
    [Magic Header 4B] + [Nonce 12B] + [Cipher Text N bytes] + [Auth Tag 16B (implicit in GCM)]
    ```
*   **Magic Header**: `0xCA 0xFE 0xBA 0xBE` (用于快速识别)
*   **特性**: 任何对密文的单位元组修改都会导致 GCM 认证失败 (`cipher: message authentication failed`)。

### 1.2 锚点类型定义
| 锚点代号 | 技术名称 | 注入位置 | 隐蔽性 | 鲁棒性 |
| :--- | :--- | :--- | :--- | :--- |
| **Anchor 1** | Attachment | Root/Names/EmbeddedFiles | ★★ | Low (易被替换) |
| **Anchor 2** | SMask | XObject (Image) -> /SMask | ★★★★ | Medium (依赖图像) |
| **Anchor 3** | Content | Page Content Stream (TJ op) | ★★★★★ | High (清洗即损毁 - 理论值) |
| **Anchor 4** | Visual | XObject (Form) | ★ | Low (可见即易删) |

---

## 2. 攻防详细数据记录 (Detailed Logs)

### 📅 Phase 6: 流内容清洗突破

**红队攻击手法**:
*   **对象定位**: 精确锁定 Object 72 (Embedded File Stream)。
*   **替换策略**: 保持 Stream 长度不变 (72 bytes)，替换内容。
*   **结果数据**:
    *   原始文件大小: 888,699 bytes
    *   清洗后文件: 888,711 bytes
    *   **差异**: +12 bytes (+0.00%)
    *   **验证报错**: `zlib: invalid header` (因红队破坏了压缩头)

### 📅 Phase 7: SMask 锚点调试与修复

**故障现象 (Phase 7.1)**:
*   签名显示成功，验证时报错 `SMask payload not found`。
*   **根本原因**: `pdfcpu` 库中 StreamDict 为值传递，修改未持久化到文件。

**修复日志**:
*   **缺陷 1**: 图像查找逻辑错误。
    *   *Fix*: 从 `pageDict.DictEntry("Resources")` 改为全量 `XRefTable` 扫描。
*   **缺陷 2**: 对象持久化失败。
    *   *Fix*: 重建 `types.StreamDict` 并赋值回 `ctx.Find(objNr).Object`。
*   **缺陷 3**: 解码数据源错误。
    *   *Fix*: 读取 `stream.Raw` (压缩数据) 而非 `stream.Content`。

### 📅 Phase 7 (Late): 双轨验证失效

**红队攻击手法 (Round 4 Re-Attack)**:
*   **技术**: Null Byte Overwrite (全零填充)。
*   **验证输出**:
    ```text
    [DEBUG] Attempting Anchor 1: Attachment...
    [DEBUG] Anchor 1: Extraction failed: zlib: invalid header
    [DEBUG] Attempting Anchor 2: SMask...
    [DEBUG] Anchor 2: Extraction failed: zlib: invalid header
    ```

**红队攻击手法 (Round 5)**:
*   **技术**: Valid Zlib Padding (合法空流填充)。
*   **效果**: 彻底规避文件格式检查，文件结构完美。

### 📅 Phase 8: 三锚点失效

**红队攻击手法 (Round 1-2)**:
*   **技术**: Adaptive Quantization (自适应量化)。
*   **验证输出**:
    *   Attachment: `0 bytes extracted`
    *   SMask: `cipher: message authentication failed` (LSB 被清洗)
    *   Content: `corrupt string literal` (内容流结构被破坏)

---

## 3. 架构重构 (Refactoring Stats)

**Phase 7.2 重构成果**:
*   **模块结构**:
    *   `injector/watermark.go`: 主入口 (API)
    *   `injector/anchor.go`: 接口定义 (Registry)
    *   `injector/crypto.go`: 加密隔离
*   **代码统计**:
    *   主入口行数: 355 → 180行 (**-49%**)
    *   循环复杂度: 15 → 8 (**-47%**)
    *   测试覆盖率: 100% (17/17 tests passed)

---

## 4. 测试框架 (Testing Framework)

**新增测试套件**: `defender/injector/comprehensive_test.go`

**测试场景覆盖**:
1.  **Full Defense**: Attachment + SMask + Content (All valid).
2.  **Resilience A**: Attachment removed -> Fallback to SMask.
3.  **Resilience B**: Attachment + SMask removed -> Fallback to Content.
4.  **Explicit Lookup**: 强制提取指定层级锚点。

**执行命令**:
```bash
go test -v -tags=integration ./defender/injector/...
```

---

## 5. 权利追究能力评估 (Attribution Capability)

| 场景 | 原始文件对比 | 盲测 (无原件) | 结论 |
| :--- | :--- | :--- | :--- |
| **Phase 5 (Attachment)** | 可检测 (+bytes) | 不可检测 | 弱 |
| **Phase 6 (Stream Clean)** | 可检测 (+12B) | **不可检测** | 丧失 |
| **Phase 7 (Valid Zlib)** | **不可检测** | **不可检测** | 完全丧失 |
| **Phase 9 (Visual)** | 明显缺失 | 需人工确认 | 无技术强制力 |

**最终结论**: 在红队掌握“合法 Zlib 填充”技术后，蓝队在盲测环境下的权利追究能力（即证明文件被篡改的能力）已降为零。
