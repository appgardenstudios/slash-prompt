---
name: analyze
description: Analyze code for issues and improvements
arguments:
  - name: language
    description: Programming language
    required: true
  - name: focus
    description: Areas to focus on
    required: false
resources:
  - path: ../resources/context.template.md
    name: context-template
---

Please analyze the following ${language} code focusing on ${focus}.

Use the provided context template to structure your analysis.