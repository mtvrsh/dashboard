# AGENTS.md

## Regenerating the screenshot

The README image comes from the standalone demo binary (`cmd/demo`) and uses
mocked data, so it stays stable even when the host machine changes. After any
UI change (`ui/index.template` or `ui/style.css`), regenerate it:

```bash
go run ./cmd/demo -screenshot examples/screenshot.webp -dark
go run ./cmd/demo -demo examples/demo.html
```

Requires Firefox and `cwebp` in `PATH`. The mocked values live in `cmd/demo/main.go`.
The demo is a separate binary and shares the embedded template/CSS with the server
through the `ui` package.

Note: headless Firefox renders text about 2px lower than the interactive browser,
so the generated WebP differs slightly from the live page (letter spacing) even
though the HTML/CSS are identical. This is a headless rendering quirk, not a
styling bug — don't compensate for it in the template or CSS.
