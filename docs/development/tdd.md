# Test-Driven Development

MemLore requires RED → GREEN → REFACTOR for behavioral production code.

1. Identify the smallest observable behavior.
2. Write a failing test.
3. Run it and confirm RED for the expected reason.
4. Implement the minimum production change.
5. Confirm GREEN.
6. Run the relevant broader suite.
7. Refactor while staying green.

Defect fixes start with a regression test. Tests should name business rules,
not incidental structure.

See constitution principle I.
