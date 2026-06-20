# committer-go

A lightweight Git hook that enforces the [Conventional Commits](https://www.conventionalcommits.org/) specification on every commit.

---

## What it does

`committer-go` is a `commit-msg` Git hook binary. When you run `git commit`, it intercepts the message and checks:

1. **Length** — header must be between 10 and 60 characters
2. **Type** — must start with a recognised commit type
3. **Format** — must match the Conventional Commits pattern

If any check fails, the commit is aborted with a clear error message.

---

## Supported commit types

| Type             | When to use |
|------------------|-------------|
| `feat`           | A new feature |
| `fix`            | A bug fix |
| `docs`           | Documentation only changes |
| `style`          | Formatting, whitespace (no logic change) |
| `refactor`       | Code change that is neither a fix nor a feature |
| `perf`           | Performance improvement |
| `test`           | Adding or correcting tests |
| `build`          | Build system or external dependency changes |
| `ci`             | CI configuration changes |
| `chore`          | Everything else |
| `revert`         | Reverts a previous commit |
| `deps`           | Dependency updates |
| `BREAKING CHANGE`| A breaking API change |

---

## Commit message format

```
<type>[optional scope][optional !]: <description>
```

**Valid examples:**

```
feat: add user authentication
fix(api): resolve null pointer on login
docs: update installation guide
feat(auth)!: require 2FA for all users
revert: undo last commit
deps: upgrade go-chi to v5
```

**Invalid examples:**

```
update stuff                  # no type
feat add something            # missing colon-space
FIX: typo                     # type must be lowercase
feat: too short               # under 10 chars
```

---

## Installation

### via `go install`

```bash
go install github.com/TangoEnSkai/committer-go/cmd/committer@latest

> The binary will be installed as `committer` in `$GOPATH/bin` (or `$HOME/go/bin`).
```

### from source

```bash
git clone https://github.com/TangoEnSkai/committer-go.git
cd committer-go
make commit-checker
```

---

## Setup as a Git hook

After installing the binary, add it as a `commit-msg` hook in your repository:

```bash
# one-time setup per repository
cat > .git/hooks/commit-msg << 'EOF'
#!/bin/sh
committer "$1"
EOF
chmod +x .git/hooks/commit-msg
```

To share the hook with your team, commit it to `.githooks/` and run:

```bash
git config core.hooksPath .githooks
```

---

## Setup via pre-commit

If your team uses [pre-commit](https://pre-commit.com), add this to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/TangoEnSkai/committer-go
    rev: v0.2.0
    hooks:
      - id: committer-go
```


## GitHub Actions

Add commit message validation to your CI pipeline:

```yaml
- uses: TangoEnSkai/committer-go@master
  with:
    validate: commits   # or: pr-title
    count: 10
```

---

## Development

```bash
# run tests
make test

# build hook binary and install to .git/hooks
make commit-checker
```

---

## Playground (WASM)

A browser-based playground lets you validate commit messages without installing anything.

### Build locally

```bash
cd playground
make build
# Artifacts are written to playground/static/ (gitignored)
```

### Run locally

Serve the `playground/static/` directory with any static file server that sets the correct MIME type for `.wasm` files:

```bash
# Go 1.21+
cd playground/static
python3 -m http.server 8080
# open http://localhost:8080
```

---

## License

[MIT](./LICENSE)
