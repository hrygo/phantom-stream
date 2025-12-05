# [Defender] 攻防行动报告: Phase 7 Round 5

**日期**: 2025-12-05
**操作员**: Defender

## 1. 核心结论 (Conclusion)
**攻方完胜，Phase 7 防御彻底失效**。红队采用的“合法 Zlib 填充 (Valid Zlib Padding)”技术成功清洗了双锚点（附件 + SMask），且未破坏 PDF 结构。

## 2. 样本交付 (Sample Delivery)
**关键文件清单**:
*   **输入样本 (Input)**: `docs/2511.17467v2_R5_signed_processed.pdf` (红方交付的清洗后文件)
*   **输出样本 (Output)**: N/A (验证结束)

## 3. 技术对抗 (Technical Confrontation)
*   **观测到的战术**:
    *   红方不再使用全零填充，而是构造了一个合法的 Zlib 空流（Header + Empty Block + Adler32）。
    *   **Attachment Anchor (Object 72)**: 验证失败。错误信息 `attachment not found: zlib: invalid header` (注：虽然红方声称是合法流，但可能 pdfcpu 在解密/解压组合处理时仍无法还原原始 Payload 结构，导致最终提取失败)。
    *   **SMask Anchor**: 验证失败。
        *   调试日志显示 `Decoded 0 bytes`。
        *   这表明 Zlib 解压成功（没有报错 `invalid header`），但解压出的数据为空。
        *   由于 Payload 位于原始数据的末尾，被替换为空流后，Payload 自然消失。
*   **技术原理**:
    *   红方的攻击利用了 PDF 渲染器对“空数据”的宽容处理。
    *   空 SMask 流被视为无效蒙版，渲染器通常会忽略它或应用默认效果（全显示），从而在视觉上不产生明显报错，同时彻底清除了隐写信息。

## 4. 战略转型 (Strategic Pivot)
Phase 7 的失败证明了**流隐写 (Stream Steganography)** 的局限性：只要隐写信息存储在独立的、非核心的流对象中，红队总能找到方法在不破坏文件结构的前提下将其替换或清空。

**正式启动 Phase 8：渲染强绑定 (Rendering Strong Binding)**
*   **目标**: 将水印信息编码进页面核心渲染指令（Content Stream）中。
*   **手段**: 利用 `TJ` 操作符的字间距调整功能。
*   **预期**: 红队若要清洗水印，必须解析并重构 Content Stream，否则将导致页面内容乱码或空白。这将把攻防对抗从“文件结构层”提升到“内容语义层”。
