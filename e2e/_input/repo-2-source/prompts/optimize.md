---
name: optimize
description: Performance optimization recommendations
arguments:
  - name: target
    description: What to optimize (speed, memory, size, etc.)
    required: true
  - name: constraints
    description: Any constraints or limitations
    required: false
  - name: current_metrics
    description: Current performance metrics if available
    required: false
resources:
  - path: ../resources/performance-tips.json
    name: perf-tips
---

I'll help you optimize for **${target}** performance.

**Constraints:** ${constraints}
**Current Metrics:** ${current_metrics}

Based on our performance optimization guidelines, here are my recommendations: