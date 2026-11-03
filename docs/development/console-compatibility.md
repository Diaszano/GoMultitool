# Console compatibility

Run `multitool --console` for a line-oriented, ASCII-safe menu that does not require cursor control or color support. The Go console currently exposes read-only hardware inventory. Backup, restore, erase, and image burning remain available through the preserved legacy `dialog` recovery menu; they are not exposed through this Go console frontend.

Destructive confirmations must accept an explicit `y`; empty input and every other response leave the operation unapproved. Storage services still enforce their own validation and destination rechecks.
