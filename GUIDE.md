# Guide

Human steps for **grams** live on the dApp: http://127.0.0.1:8081/guide

## 0. What you have

- App source: `C:\Users\Remco\kns`
- This map: `C:\Users\Remco\Documents\kaspa\superapp`
- Compiler: `C:\Users\Remco\tools\silverc\silverc.exe` (official v1-rc1 zip)
- Language clone: `C:\Users\Remco\silverscript` @ `c7d17a1`
- Live names: KNS indexer, not this disk

## 1. Run

```powershell
cd C:\Users\Remco\kns
go test ./...
go run ./cmd/kns
```

Open http://localhost:8080

Health should say product `kns-web4`.

## 2. Human path (name)

1. `/app` — type `kns.kas` or a `kaspa:` address.
2. `/name/kns.kas` — owner, pay URI, live ocean rank (JS fetch).
3. `/site/kns.kas` — generated site. Add `?format=json` for the machine view.
4. `/register` — prices a name and builds the envelope. **Does not broadcast.** Use app.knsdomains.org or Kasware to actually inscribe.
5. `/honest` — every claim with a status tag. Read it before repeating slogans.

Connect Kasware on `/me` if you want this browser to see your address. Optional.

## 3. Agent path (Web4.0)

```
GET  /agent/kns.kas.json
GET  /mcp
POST /mcp          {"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"resolve_kas","arguments":{"q":"kns.kas"}}}
GET  /api/v1/call/kns.kas     → 402
```

MCP tools: `resolve_kas`, `agent_card`, `pay_kas`, `kachat_contact`, `kas_rank`, `quote_work`.

402 is Kaspa-native. It is not Coinbase x402. Retry with header `X-Kaspa-Payment: <txid>` only tests the HTTP layer.

## 4. Ecosystem pointers (not clones)

| Open | Expect |
| --- | --- |
| `/kachat?q=kns.kas` | Contact = owner address. No E2E. |
| `/kassword` | Link to kassword.com + vaultCommit convention. |
| `/ranks?q=kns.kas` | Shrimp… from live balance. Not the NFT. |
| `/kcc` | KCC-0/1/2/20 drafts. |
| `/kaposts` | Local feed. Dies on restart. |

## 5. Two dApps

Work-credit alternative (this folder, `:8081`):

```powershell
cd C:\Users\Remco\Documents\kaspa\superapp
go run ./cmd/gramlane
```

Reserved L1 stable vision (no grams, no L2) `:8082`:

```powershell
cd C:\Users\Remco\Documents\kaspa\superappstablesalternative
go run ./cmd/kastill
```

## 5b. Work credits (predictable dApp fees)

Quote 1 KAS of work at policy rate (1e6 grams):

```
http://localhost:8080/credits?grams=1000000
http://localhost:8080/api/v1/credits/quote?grams=1000000
```

You should see `credits = 1000000`, `kas = 1`, `usd = not quoted`.

MCP:

```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"quote_work","arguments":{"grams":"50000","lane":"KNS1"}}}
```

This does **not** mint a UTXO. It is the invoice. Deploying `WorkCredit.sil` on mainnet needs a wallet that can genesis Toccata covenants (tooling still catching up). Treat the artifact as the spec you will genesis later.

Recompile:

```powershell
cd C:\Users\Remco\kns
.\compile-sil.ps1
```

Artifact URLs:

- `/api/v1/artifact/KasName`
- `/api/v1/artifact/KaChatPayTimeout`
- `/api/v1/artifact/WorkCredit`

## 6. When to use which money

| You are… | Use |
| --- | --- |
| Paying a person / name | KAS to the pay URI |
| Budgeting a sequenced dApp / agent loop | Work Credits (grams) |
| Selling coffee, payroll, DeFi dollars | Kaspa Till (`:8082`) — reserved L1 kUSD, KAS today. No L2. |
| Tempted to launch “kUSD” on L1 | Stop. Read WORK-CREDITS.md |

## 7. Claims you should not repeat

- “`.kas` is unique on L1” — indexer FCFS.
- “This app is KaChat / Kassword” — pointers only.
- “WorkCredit is a stablecoin” — it is a gram voucher.
- “Covenants back a dollar” — they back a ledger of work.
- “Silverscript v1 is tagged mainnet” — you have **v1-rc1**.
- “vProgs / DAGKnight / 100 BPS are live” — they are not.

## 8. Files to touch

| Want to change… | File |
| --- | --- |
| Live resolve | `internal/knsapi`, `internal/resolver` |
| Agent card / 402 | `internal/web4`, `internal/server/web4.go` |
| Gram quotes | `internal/workcredit` |
| Covenant source | `contracts/v1/*.sil` |
| Honesty table | `internal/protocol/status.go` |
| This guide | `Documents\kaspa\superapp\` |
