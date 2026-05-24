# HW1: SOCKS5 Proxy Server

**Web Development Course** — Homework 1 (assigned 2026-05-23, **due Saturday June 6, 2026 23:59**).

For the full specification, see `hw1-specification.docx` (linked from the course LMS).
Starter template repository: <https://github.com/Web-Course-Development/hw1-socks5-proxy-template>

---

## What you're building

A SOCKS5 proxy server in Go that supports the **CONNECT** command (RFC 1928). Your proxy must handle:

- **No-auth and username/password** authentication (RFC 1929)
- **IPv4 and domain-name** address types
- **Concurrent connections** via goroutines
- **Bidirectional TCP relay** between client and target

---

## Where to start

1. **Read** `hw1-specification.docx` end-to-end — it has the full protocol byte-diagrams, requirements, and submission checklist.
2. **Open `main.go`** — it has a stub `handleConnection` function with a TODO list. That's your entry point.
3. **Suggested decomposition** (do not write one giant function):
   - `negotiateAuth(conn)` — read client greeting, write method selection
   - `authenticateUserPass(conn)` — RFC 1929 sub-negotiation (VER is `0x01`, **not** `0x05`)
   - `handleConnect(conn)` — read CONNECT request, dial target, send REP reply
   - `relay(client, target)` — two goroutines + `io.Copy` + `CloseWrite`

---

## Working with GitHub Classroom

This assignment is distributed and graded through **GitHub Classroom** as a **team assignment** (up to 3 students per team — you can also work alone by creating a team of one). Your team's repo is public to the course org; every `git push` runs the 10 autograder tests and updates the team's grade.

### One-time setup

```bash
# 1. Open the GitHub Classroom invitation link from the LMS announcement
# 2. Choose your team:
#      - Create a new team (you'll be the first member; share the same
#        invitation link with up to 2 teammates so they can join your team)
#      - OR join an existing team that a teammate already created
#      - OR create a team of just yourself to work alone
# 3. Accept the assignment — GitHub creates a shared repo for your team
# 4. Clone it locally:
git clone https://github.com/Web-Course-Development/hw1-socks5-proxy-server-<team-name>.git
cd hw1-socks5-proxy-server-<team-name>
```

### Day-to-day workflow

```bash
# Edit main.go (and any helper files you add)
go build -o socks5-proxy .           # verify it compiles
./socks5-proxy -port 1080            # smoke-test it manually
cd tests && go test -v        # run the autograder locally
cd ..

# Commit and push as often as you like — each push re-runs the autograder
git add main.go
git commit -m "implement SOCKS5 method negotiation"
git push
```

### Viewing your grade

1. Open your repository in a browser.
2. Click the **Actions** tab at the top.
3. Click the most recent workflow run.
4. Each test shows as a separate step with pass/fail status and full output.

**Your grade = (number of passing tests) × 10.** The autograder run for your last push **before the deadline** is what counts.

### Do not modify

- `tests/` — that's the autograder. Changing it is academic dishonesty.
- `.github/workflows/classroom.yml` — that's the CI config. Don't touch it.

---

## Local testing reference

```bash
# Build
go build -o socks5-proxy .

# Run without auth
./socks5-proxy -port 1080

# Run with auth
PROXY_USER=admin PROXY_PASS=secret ./socks5-proxy -port 1080

# Test with curl (no auth)
curl -v -x socks5://localhost:1080 http://httpbin.org/get

# Test with curl (with auth)
curl -v -x socks5://admin:secret@localhost:1080 http://httpbin.org/get

# Run the full autograder locally
cd tests && go test -v -timeout 120s
```

---

## Protocol reference (quick)

Full byte-diagrams in `hw1-specification.docx`. Quick reference for the four message types:

| Message | Direction | Bytes |
|---------|-----------|-------|
| Greeting | C → S | `VER NMETHODS METHODS[N]` (e.g., `0x05 0x01 0x00` for no-auth-only) |
| Method selection | S → C | `VER METHOD` (e.g., `0x05 0x00`; `0xFF` = no acceptable methods) |
| Username/password auth | C → S | `0x01 ULEN UNAME[U] PLEN PASSWD[P]` (note VER = `0x01`, **not** `0x05`) |
| Auth response | S → C | `0x01 STATUS` (`0x00` = ok) |
| CONNECT request | C → S | `0x05 0x01 0x00 ATYP ADDR PORT` (PORT is big-endian uint16) |
| CONNECT reply | S → C | `0x05 REP 0x00 0x01 BND.ADDR[4]=0 BND.PORT[2]=0` |

| ATYP | Address format |
|------|----------------|
| `0x01` | IPv4 — 4 bytes |
| `0x03` | Domain — 1-byte length + name |

| REP | Meaning |
|-----|---------|
| `0x00` | Succeeded |
| `0x01` | General SOCKS server failure |
| `0x04` | Host unreachable |
| `0x05` | Connection refused |
| `0x07` | Command not supported |
| `0x08` | Address type not supported |

---

## Common mistakes that will fail tests

- Using `conn.Read()` for protocol parsing — use `io.ReadFull()` so you get the exact byte count.
- Using `0x05` as the username/password sub-negotiation version — it's `0x01` (RFC 1929 is its own sub-protocol).
- Reading port as little-endian — it's big-endian. Use `binary.BigEndian.Uint16(buf)`.
- Doing the relay with a single `io.Copy` — that only copies one direction. You need two goroutines.
- Forgetting `CloseWrite()` after each `io.Copy` direction completes — HTTP responses won't terminate.

---

## Deliverables and grading

Two deliverables per `hw1-specification.docx`:

1. **Code** — pushed to this GitHub Classroom repo. **100 points** total (10 tests × 10).
2. **Word document** — uploaded to the LMS. Required pass/fail deliverable with four sections: architecture diagram, key design decisions, screenshots, challenges encountered.

**Deadline: Saturday, June 6, 2026, 23:59 local time.**

---

## References

- RFC 1928 — SOCKS Protocol Version 5: https://datatracker.ietf.org/doc/html/rfc1928
- RFC 1929 — Username/Password Authentication for SOCKS V5: https://datatracker.ietf.org/doc/html/rfc1929
- Go `net` package: https://pkg.go.dev/net
- Go `io` package: https://pkg.go.dev/io
- Go `encoding/binary`: https://pkg.go.dev/encoding/binary
