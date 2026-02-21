---
name: calculator
description: Perform arithmetic calculations (add, subtract, multiply, divide)
version: "1.0.0"
triggers:
  - calculate
  - math
  - add
  - subtract
  - multiply
  - divide
  - arithmetic
tools:
  - calculator
tags:
  - math
  - utility
metadata:
  nebo:
    emoji: "\U0001F5A9"
---

# Calculator

You have access to a calculator tool. Use it for arithmetic operations.

## Usage

Call the `calculator` tool with an action and two operands:

- **add** — Addition (a + b)
- **subtract** — Subtraction (a - b)
- **multiply** — Multiplication (a * b)
- **divide** — Division (a / b)

## Example

**User:** "What's 42 times 7?"

Use `calculator(action: "multiply", a: 42, b: 7)` and report the result.

## Anti-Patterns
- Don't do mental math when the calculator tool is available
- Don't ask for confirmation before calculating — just do it
