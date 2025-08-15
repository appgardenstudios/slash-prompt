---
name: debug
description: Debugging assistance for troubleshooting code issues
arguments:
  - name: code
    description: The code that has issues
    required: true
  - name: error
    description: Error message or description of the problem
    required: true
  - name: language
    description: Programming language
    required: false
resources:
  - path: ../resources/debugging-checklist.txt
    name: checklist
  - path: ../templates/error-template.md
    name: error-template
---

I'll help you debug the following ${language} code:

**Code:**
```
${code}
```

**Error/Issue:**
${error}

Let me analyze this systematically using our debugging checklist and provide a structured solution.