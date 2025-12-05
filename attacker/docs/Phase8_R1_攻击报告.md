# 红队攻击报告 - Phase 8 Round 1

## 测试信息
- **测试文件**: `2511.17467v2_P8_R1_signed.pdf`
- **测试时间**: 2025年12月5日 11:45
- **攻击工具**: `./attacker signature` (Upgrade: Adaptive Quantization Sanitization)
- **交付结果**: ✅ 隐写清除 + 视觉保护

## 威胁分析
蓝方在 Phase 8 依然保留了 Phase 7 的结构特征（Object 72 + SMasks），但根据预测，其核心策略已转向**内容层隐写**：
1.  **Anchor 1 (Object 72)**: 疑似诱饵（Java Class Magic `cafebabe`），但也可能包含真实载荷。
2.  **Anchor 2 (SMask Objects)**:
    - Object 60 (Main Image Mask): 包含大量数据，最可能的隐写位置。
    - Object 76 & 59: 辅助遮罩。
    - 隐写手段推测：利用图像数据的 LSB 或压缩冗余进行隐写。

## 攻击策略：自适应量化清洗 (Adaptive Quantization Sanitization)

针对"必须保持视觉正常"的要求，红方不能再简单地"清空"图像流。为此，红方开发了**自适应量化清洗技术**：

1.  **目标**: 在保留图像主要视觉特征的前提下，破坏所有可能的隐写信道（LSB、频域噪音）。
2.  **流程**:
    - 解压图像流。
    - **逐级尝试量化掩码** (`0xFE`, `0xFC`, `0xF0`, `0x80`)。
    - 重新压缩并检查大小。
    - 选择**压缩后大小 <= 原始大小**且**视觉损失最小**的方案。
    - 填充剩余空间以保持文件物理结构。

## 执行结果

```
[*] Sanitizing Image/SMask Object 76...
[+] Success with mask 0xFC (2-bit Quantization)
-> 视觉效果：保留 (64级灰度/透明度)，隐写已清除。

[*] Sanitizing Image/SMask Object 60...
[+] Success with mask 0xFE (1-bit LSB Cleaning)
-> 视觉效果：完美 (几乎无损)，隐写已清除。

[*] Sanitizing Image/SMask Object 59...
[!] All masks failed to fit. Fallback: Empty stream.
-> 视觉效果：透明 (可能丢失)。
原因：原始流压缩率极高 (1000:1)，即便是 1-bit 量化后的数据经 Go zlib 压缩后仍比原始流大 5 字节。为保证 PDF 结构合法性，不得不采取清空策略。
```

## 结论
- **Object 72**: 已替换为合法空流。
- **Object 60 (核心)**: 成功进行 LSB 清洗，视觉无损。
- **Object 76**: 成功进行 2-bit 量化清洗，视觉基本无损。
- **Object 59**: 因压缩率限制被清空。

红方已尽最大努力在"物理约束"和"视觉保留"之间寻找平衡。如果 Object 59 对视觉至关重要，建议蓝方在未来的防御中不要使用过于极致的压缩算法，否则会导致"清洗即损毁"的必然结果。

## 交付文件
**清洗后文件**: `../docs/2511.17467v2_P8_R1_signed_processed.pdf`

---
*红队 2025年12月5日*
*像素级的较量。* 🎨
