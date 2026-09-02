# Gramlane — dApp 1 (work-credit alternative)

**project delusional** · [@StppStp](https://x.com/StppStp)

A **Kaspa L1** sequenced work desk. Jobs are billed in **grams** (Work Credits). Not a dollar. **No L2.**

The stable *alternative is possible for fees only*. This dApp is that alternative.

Sister dApp (no grams; vision of a native Kaspa L1 stable):  
`C:\Users\Remco\Documents\kaspa\superappstablesalternative` — Kaspa Till on `:8082`.

```powershell
cd C:\Users\Remco\Documents\kaspa\superapp
go test ./...
go run ./cmd/gramlane
```

http://localhost:8081

| Path | What |
| --- | --- |
| `/desk` | Job catalog in GRAM |
| `/job/resolve` | Quote + run a name lookup |
| `/api/run?job=dag` | HTTP 402 until `X-Work-Credit` |
| `/honest` | Claims |
| `/wallets` | All Kaspa wallets. Inject: Kasware + Kastle. Ledger: KasVault. |

## Map (this folder)

| File | What |
| --- | --- |
| [VISION.md](VISION.md) | Why names + predictable work costs |
| [DIFF-MAP.md](DIFF-MAP.md) | Live / local / wrong |
| [HOW-IT-WORKS.md](HOW-IT-WORKS.md) | Stack |
| [WORK-CREDITS.md](WORK-CREDITS.md) | Why grams, not USD |
| [GUIDE.md](GUIDE.md) | How to operate |

## Honesty

Covenants lock work. They do not mint USD. This dApp never routes through an L2. A Kaspa L1 dollar is **not live**; that slot is the other dApp.
