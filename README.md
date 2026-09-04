# Gramlane — dApp 1 (work-credit alternative)

**project delusional** · [@StppStp](https://x.com/StppStp)

**Stable work price on L1, not a synthetic dollar.**

A **Kaspa L1** work-credit lane. Jobs are billed in **grams** (KIP-21 mass). Not a dollar. Does not replace USDT. **No L2 in this path.** Vaults still lock KAS; grams pay the action.

Sister dApp (when the invoice really is money):  
`C:\Users\Remco\Documents\kaspa\superappstablesalternative` — Kaspa Till on `:8082`.

```powershell
cd C:\Users\Remco\Documents\kaspa\superapp
go test ./...
go build -o gramlane.exe ./cmd/gramlane
.\gramlane.exe
```

http://localhost:8081

If the browser says “localhost refused to connect”, the process is dead. Start all three:

```powershell
powershell -File C:\Users\Remco\Documents\kaspa\start-local.ps1
```

Index of the stack: [STP-KAS/project-delusional](https://github.com/STP-KAS/project-delusional).

| Path | What |
| --- | --- |
| `/` | The jar — fund, pay, till |
| `/why` | Beyond the chain: household, plant, agent, counter |
| `/vault` | Vault bump — grams pay the action, KAS stays locked |
| `/postage` | Message postage board |
| `/agent` | AI agent (Grok or local tools) |
| `/site` | .kas web domain builder |
| `/why` | Bill the action, not the exchange rate |
| `/guide` | Step-by-step: what you do with grams |
| `/genesis` | WorkCredit instance: 0.5 KAS sale + funded P2SH voucher UTXO |
| `/stablegram` | Set aside % of KAS into the Stablegram jar |
| `/pay` | Pay hub — EUR/USD ticket, QR, counter |
| `/pay/{id}` | Pay a ticket from the jar |
| `/kasdomain` | Your names: one unique sign, living page |
| `/market` | Buy/sell Kasdomains for grams only |
| `/kachat` | KaChat: resolve a name, postage, pay that wallet |
| `/apps` | Apps: KaChat, vault, agent, site, jobs |
| `/convert` | KAS ↔ grams at policy 100 sompi/gram. Not a KCC-20 |
| `/desk` | Job catalog in GRAM |
| `/job/resolve` | Quote + burn prepaid grams |
| `/api/run?job=dag` | Burns prepaid grams; 402 only after inventory is spent |
| `/api/seq` | Remaining grams, voucher outpoint |
| `/honest` | Claims |
| `/234` | Worked #234: same 42 bytes, amount 1 → vault 264 |
| `/wallets` | All Kaspa wallets. Inject: Kasware + Kastle. Ledger: KasVault. Log out supported. |
| `/safety` | This site never DMs you. Never paste a seed. |
| `/feedback` | Stored on this PC under `Documents\kaspa\feedback\gramlane` |

## Map (this folder)

| File | What |
| --- | --- |
| [VISION.md](VISION.md) | Work-credit lane: grams meter work, not USD |
| [DIFF-MAP.md](DIFF-MAP.md) | Live / local / wrong |
| [HOW-IT-WORKS.md](HOW-IT-WORKS.md) | Stack |
| [WORK-CREDITS.md](WORK-CREDITS.md) | Why grams, not USD |
| [GUIDE.md](GUIDE.md) | How to operate |

## Honesty

Covenants lock work. They do not mint USD. This dApp never routes through an L2. USDT can sit off to the side; it is not the meter. A Kaspa L1 dollar is **not live**; that slot is the other dApp.
