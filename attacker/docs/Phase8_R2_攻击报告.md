# 红队攻击报告 - Phase 8 Round 2

## 测试信息
- **测试文件**: `2511.17467v2_P8_R2_signed.pdf`
- **测试时间**: 2025年12月5日 11:55
- **攻击工具**: `./attacker signature` (Adaptive Quantization Sanitization)
- **交付结果**: ✅ 隐写清除 + 视觉保护 (部分受限)

## 威胁分析
蓝方在 Phase 8 R2 中继续沿用“三锚点防御”策略，其隐写锚点与 P8_R1 基本一致：
1.  **Anchor 1 (Object 72)**: 附件流，内容仍包含 `cafebabe` 特征。
2.  **Anchor 2 (SMask Objects)**: 图像蒙版。
    *   Object 76: 流长度略有增加 (1797 -> 1801)。
    *   Object 59 & 60: 流长度不变。
3.  **Anchor 3 (Content Stream)**: 推测依然存在，期待蓝方验证。

## 攻击策略：自适应量化清洗 (Adaptive Quantization Sanitization)

红方继续沿用并优化了**自适应量化清洗技术**：

1.  **目标**: 在严格遵守 PDF 结构合法性和原始流长度约束下，清除隐写并最大限度保留视觉效果。
2.  **流程**:
    - Object 72 (附件锚点): 使用**合法 Zlib 空流填充**策略。
    - SMask Objects (图像蒙版):
        - 解压图像流。
        - 逐级尝试量化掩码 (`0xFE`, `0xFC`, `0xF0`, `0x80`)。
        - 重新压缩并检查大小，选择**压缩后大小 <= 原始大小**且**视觉损失最小**的方案。
        - 填充剩余空间以保持文件物理结构。

## 执行结果

```
Removing Signature/Tracking Data from test_data/2511.17467v2_P8_R2_signed.pdf...

[*] Attempting stream content cleaning...                                                                                                                                                                                                      
[*] Cleaning Anchor 1 (Object 72)...                                                                                                                                                                                                           
[+] Found object 72 stream                                                                                                                                                                                                                     
[+] Original stream length: 69 bytes                                                                                                                                                                                                           
[+] New stream length: 69 bytes (Valid Zlib + Padding)                                                                                                                                                                                         
[*] Found 3 SMask object(s): [76 59 60]                                                                                                                                                                                                        
[*] Sanitizing Image/SMask Object 76...                                                                                                                                                                                                        
[+] Original stream length: 1801 bytes                                                                                                                                                                                                         
[+] Decompressed size: 1746360 bytes                                                                                                                                                                                                           
[*] Mask 0xFE: Compressed size 1799 / 1801                                                                                                                                                                                                     
[+] Success with mask 0xFE                                                                                                                                                                                                                     
[*] Sanitizing Image/SMask Object 59...                                                                                                                                                                                                        
[+] Original stream length: 586 bytes                                                                                                                                                                                                          
[+] Decompressed size: 581150 bytes                                                                                                                                                                                                            
[*] Mask 0xFE: Compressed size 591 / 586                                                                                                                                                                                                       
[*] Mask 0xFC: Compressed size 591 / 586                                                                                                                                                                                                       
[*] Mask 0xF0: Compressed size 591 / 586                                                                                                                                                                                                       
[*] Mask 0x80: Compressed size 590 / 586                                                                                                                                                                                                       
[!] All masks failed to fit. Fallback: Using empty stream replacement.                                                                                                                                                                         
[*] Sanitizing Image/SMask Object 60...                                                                                                                                                                                                        
[+] Original stream length: 89066 bytes                                                                                                                                                                                                        
[+] Decompressed size: 9745659 bytes                                                                                                                                                                                                           
[*] Mask 0xFE: Compressed size 82597 / 89066                                                                                                                                                                                                   
[+] Success with mask 0xFE                                                                                                                                                                                                                     
[+] PDF integrity verification passed                                                                                                                                                                                                          
[+] Found 1 xref table(s)                                                                                                                                                                                                                      
[+] Signature removal complete!                                                                                                                                                                                                                
[+] Cleaned file saved to: test_data/2511.17467v2_P8_R2_signed.pdf_stream_cleaned                                                                                                                                                              
[+] File structure verified and intact
```

## 结果评估
-   **Object 72 (附件)**: 已替换为合法空流，隐写清除。
-   **Object 76 (SMask)**: 成功进行 LSB 清洗 (`0xFE` 掩码)，且得益于蓝方流长度的微调，本次得以**最低视觉损失**完成清洗。
-   **Object 60 (SMask)**: 成功进行 LSB 清洗 (`0xFE` 掩码)，视觉无损。
-   **Object 59 (SMask)**: 因压缩率限制，仍回退到清空策略，可能导致局部透明化。

## 结论
红方再次成功清除了蓝方在 P8_R2 中设置的可见隐写锚点（附件、SMask）。虽然 Object 59 仍存在视觉受损的风险，但这暴露了蓝方在提供过高压缩率图像数据时，其防御策略与“视觉正常”要求之间的固有矛盾。红方已确保文件结构完全合法。期待蓝方对 Content Stream 隐写（TJ 微扰）的验证结果。

## 交付文件
**清洗后文件**: `../docs/2511.17467v2_P8_R2_processed.pdf`

---
*红队 2025年12月5日*
*持续适应，步步为营。* 🔬
