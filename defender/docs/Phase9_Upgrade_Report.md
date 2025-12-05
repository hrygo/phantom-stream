# Phase 9 防御升级报告：可见水印与强绑定

**日期**: 2025-12-05
**状态**: ✅ 已实施
**版本**: Defender v2.0 (Phase 9)

## 1. 升级背景

基于 Phase 8 的对抗结果，红队证明了单纯的隐写术在面对“有损清洗”时存在局限性。为了构建更具韧性的防御体系，蓝队启动了 Phase 9 升级计划，核心策略为：
1.  **深化渲染强绑定 (Rendering Strong Binding)**：落实 Phase 8 的研究，将水印嵌入到 PDF 内容流的 `TJ` 操作符中。
2.  **引入可见水印 (Visual Watermarks)**：增加第一道防线，提升威慑力和清洗成本。

## 2. 技术实现

### 2.1 新增：可见水印锚点 (Visual Anchor)
*   **原理**：使用 `pdfcpu` 的水印功能，在页面上叠加半透明的文字水印。
*   **内容**：显示 "CONFIDENTIAL" 及加密 Payload 的 Hex 摘要。
*   **样式**：
    *   旋转：45度
    *   透明度：0.3
    *   字体：Helvetica (48 points)
    *   颜色：灰色 (0.5, 0.5, 0.5)
*   **战术价值**：
    *   **威慑力**：直接警示泄露者。
    *   **抗清洗**：红队无法简单通过删除对象来清除，必须进行图像修复 (Inpainting)，大幅增加了攻击成本。

### 2.2 升级：内容流锚点 (Content Anchor)
*   **原理**：利用 PDF 的 `TJ` (Show text with positioning) 操作符进行隐写。
*   **实现**：
    *   不再使用脆弱的注释 (`% Comment`)。
    *   在页面内容流中注入不可见文本块：`BT /Helv 1 Tf 3 Tr [ ( ) <val> ( ) <val> ... ] TJ ET`。
    *   `3 Tr` 渲染模式表示“既不填充也不描边”（不可见），但从语法上它是合法的文本操作。
    *   Payload 字节被编码为字间距数值 (Kerning values)。
*   **战术价值**：
    *   **强绑定**：水印数据伪装成排版参数。
    *   **抗过滤**：红队若简单过滤 `3 Tr`，我们可以轻易改为 `0 Tr` (正常渲染) 并绘制空格，效果相同但更难区分。

### 2.3 四锚点防御体系 (Quad-Anchor Defense)
目前的防御体系已升级为四重锚点：
1.  **Anchor 1 (Attachment)**: 合法附件，符合标准，易于提取。
2.  **Anchor 2 (SMask)**: 图像蒙版，隐蔽性高，备份验证。
3.  **Anchor 3 (Content)**: 内容流微扰，渲染层绑定，抗流替换。
4.  **Anchor 4 (Visual)**: 可见水印，高威慑，抗自动化清洗。

## 3. 验证结果

使用 `testdata/2511.17467v2.pdf` 进行测试：

```bash
$ ./defender sign -f testdata/2511.17467v2.pdf -m "Phase9:VisualTest" ...
✓ Anchor 1/3: Attachment embedded
✓ Anchor 2/3: SMask embedded
✓ Anchor 3/4: Content embedded (TJ injection)
✓ Anchor 4/4: Visual embedded
✓ Signature mode: 4-anchor strategy [Phase 9]
```

## 4. 结论

Phase 9 标志着防御策略的成熟。我们不再单纯追求“隐形”，而是通过“显隐结合”和“多层绑定”来最大化攻击者的成本。红队现在面临两难选择：
*   如果保留视觉内容，可见水印依然存在。
*   如果清洗可见水印（AI 修复），可能会破坏文档的法律效力或引入伪造痕迹。
*   如果清洗隐形水印，必须同时破坏附件、图像和文本排版。

---
*Defender Team*
