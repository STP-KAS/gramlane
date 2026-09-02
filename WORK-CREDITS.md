# Work Credits — a stable-shaped fee, not a dollar

## The ask

dApps sequenced on Kaspa need **no volatility in price, so in fees and costs**. Other chains solve that with stables. Stables need **capital**. Can covenants replace that?

## Short answer

**Yes, for fees. No, for dollars.**

A covenant can make a **prepaid unit of sequenced work** that dApps invoice against. That unit does not move when KAS/USD moves. It is the same *idea* as a stable (a boring unit of account for apps) with a different *backing* (a work ledger, not a reserve).

A covenant **cannot** make `$1`. That requires an issuer who holds dollars (USDC), crypto-overcollateral plus an oracle (DAI), or a lie (UST).

## What other chains actually did

| Object | Looks like | Backing | Lesson |
| --- | --- | --- | --- |
| USDC / USDT | $1 token | Fiat + bank / T-bills | Capital. Bridging it via an L2 is out of scope for this stack. |
| DAI | $1 token | Overcollateral crypto | Still capital + oracle. |
| UST / Luna | $1 token | A mint/burn story | Do not port this. |
| CHI / GST2 (Ethereum gas tokens) | Fee hedge | Prepaid gas | Closest ancestor. Died as a dollar, worked as a hedge. |
| Aztec Fee Juice, Fuel prepaid gas, Cosmos feegrant | Prepaid execution | Locked native / voucher | This is the pattern. |
| Lightning inbound capacity | Prepaid throughput | Channel capital | Capacity, not FX. |

Kaspa does not need a new USDC. It needs the **gas-token / fee-juice** object, covenant-enforced.

## What dApps consume

Not dollars. **Grams.**

Toccata / KIP-21:

- Mass in compute, storage, transient grams
- Min-relay **policy**: 100 sompi / gram (not consensus; can change)
- 50 lanes / block, 1e9 gas / lane

Identity:

```
1 Work Credit = 1 gram of a named lane
1_000_000 grams ≈ 1 KAS at 100 sompi/gram
```

A sequenced dApp’s operating bill is `N grams` on lane `L`. Quote KAS only as a **fallback** if the user has no voucher.

## How the covenant backs it (no extra capital pool)

`WorkCredit.sil` state:

```
issuer   — sequencer / dApp operator pubkey
holder   — user (or agent) pubkey
credits  — remaining grams
lane     — 32-byte lane id (convention, not consensus uniqueness)
```

Entries:

| Entry | Who signs | Effect |
| --- | --- | --- |
| `mint(added)` | issuer | Credits increase. KAS sale happens *off-script*. |
| `consume(used)` | holder **and** issuer | Credits decrease. Means “this work was sequenced”. |
| `transfer(next)` | holder | Whole voucher changes holder. |

What the script **does**:

- Holder cannot inflate their own grams.
- Issuer cannot steal the holder’s remaining grams without the holder’s key (they *can* refuse to co-sign consume — that is counterparty risk).
- Conservation on consume/transfer. Mint is privileged, like a KCC-20 minter branch.

What the script **does not**:

- Pay miners. Miners still want KAS.
- Force inclusion. Blockspace is not a covenant output.
- Peg to USD. No oracle.
- Make `lane` globally unique. Same KIP-20 outpoint rule as names.

**Backing = the voucher accounting + the issuer’s obligation to spend KAS on inclusion.** The issuer sold grams today, received KAS today, and later spends KAS to miners. That is inventory risk on the operator, not a USDC treasury.

This is how you “bypass stables that require capital”: you never created a dollar claim, so you never needed dollar capital. You created a **work claim**.

## Dual-unit invoice (what a dApp should print)

```
work:      50_000 grams   (the stable-shaped number)
kas:       0.005 KAS      (fallback at 100 sompi/gram)
usd:       not quoted
dollar:    not this dApp — see Kaspa Till (reserved L1 kUSD, no L2)
```

HTTP 402 `Accepts` two schemes:

1. `kaspa` — pay owner in KAS
2. `kaspa-work-credit` — burn grams with sequencer co-sign

Agents prefer (2) when they hold a voucher. Humans without a voucher pay (1).

## Trust model (read this before calling it “stable”)

| Risk | Who holds it | Honest name |
| --- | --- | --- |
| KAS/USD moves after sale | Issuer (they already took KAS) | Inventory, not a peg |
| Issuer refuses to sequence | Holder | Counterparty. Pick issuers with a `.kas` name and reputation |
| Policy sompi/gram changes | Both | Recast quotes. Credits are grams, not sompi |
| Issuer mints infinitely | Market | Treat issuer like a bar tab, not like Circle |
| User wants $1 forever | Nobody on L1 today | Kaspa Till reserves the slot. Still needs capital when it lands. No L2. |

If you need the issuer to post a **KAS bond** (their own capital, still not USD), that is a future `FeeGrant` / bonded-minter — extra safety, not required for the unit of account.

## Why this matches “stables for dApps” without being one

dApps on Base invoice in USDC so that:

- yesterday’s fee table still works today
- a user’s subscription is not a KAS-denominated lottery
- agents can budget

Grams do the same **inside Kaspa’s cost model**. The number `50_000` is boring. Boring is the feature.

What it will **not** do: make a coffee cost “one kas-dollar” on L1. That is Kaspa Till’s reserved unit, waiting on a real L1 issuer or vault — not an L2.

## What we refuse to design

- Rebase token targeting $1
- “Covenant-backed USD” with no oracle and no reserve
- Claiming miners accept GRAM as fees (would be a KIP, not an app)
- Listing WorkCredit on a DEX as a stable

## Code

- Script: `C:\Users\Remco\kns\contracts\v1\WorkCredit.sil`
- Quote: `C:\Users\Remco\kns\internal\workcredit`
- Convention: `C:\Users\Remco\kns\conventions\kcc-work-credits.md`
- UI: `http://localhost:8080/credits`
- Artifact: `http://localhost:8080/api/v1/artifact/WorkCredit`
