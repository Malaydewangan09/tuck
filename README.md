# tuck

A project-aware scratchpad for developers. Stores notes, commands, todos, and warnings in the project directory, searchable across projects.

## Install

**macOS (Apple Silicon)**
```
curl -Lo tuck https://github.com/Malaydewangan09/tuck/releases/latest/download/tuck-darwin-arm64
chmod +x tuck && sudo mv tuck /usr/local/bin/
```

**macOS (Intel)**
```
curl -Lo tuck https://github.com/Malaydewangan09/tuck/releases/latest/download/tuck-darwin-x86_64
chmod +x tuck && sudo mv tuck /usr/local/bin/
```

**Linux (amd64)**
```
curl -Lo tuck https://github.com/Malaydewangan09/tuck/releases/latest/download/tuck-linux-x86_64
chmod +x tuck && sudo mv tuck /usr/local/bin/
```

**Build from source**
```
git clone https://github.com/Malaydewangan09/tuck
cd tuck && go build -o tuck . && sudo mv tuck /usr/local/bin/
```

## Usage

```
tuck note "postgres runs on 5433 here, not default"
tuck warn "dont touch auth.js, sarah is refactoring"
tuck cmd  "docker run -p 5432:5432 postgres:15"
tuck todo "add rate limiting to /api/login"
tuck snap

tuck ls
tuck run <id>
tuck done <id>
tuck rm <id>
tuck grep <term>
tuck summary
```

## Entry types

| type | description |
|------|-------------|
| note | general notes about the project |
| cmd  | runnable commands, executed with `tuck run <id>` |
| todo | tasks with done/undone toggle |
| warn | warnings, shown first in listing |
| snap | snapshot of current branch, ports, and runtime versions |

## Team mode

Commit `.tuck` to share notes with teammates.

```
tuck team on
git add .tuck && git commit -m "add project notes"

# after teammates push their notes
tuck team sync
```

`tuck team sync` merges by content hash. Duplicate entries are skipped.

## Shell hook

Add to `~/.zshrc` or `~/.bashrc` to see a summary on every directory change:

```
function cd() { builtin cd "$@" && tuck summary; }
```

## How it works

Each project stores a `.tuck` file in its root directory. A global index at `~/.tuck-index` enables cross-project search. No daemon, no server, no cloud.

## License

MIT
