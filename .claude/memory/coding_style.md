---
name: coding_style
description: 编码风格偏好
type: feedback
tags: style, preference
---

**Why:** 用户希望代码简洁高效，避免不必要的复杂性

**How to apply:**
- 避免过度抽象，三行相似代码不要提取成函数
- 不写防御性代码处理不可能发生的情况
- 信任内部代码和框架保证
- 在系统边界（用户输入、外部 API）处验证
- 不要在代码中添加未请求的功能或改进
