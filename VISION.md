# Vision

A name an agent can pay, and a fee a dApp can predict.

That is the whole product.

## What “superapp” means here

Not a single corporation. A **stack** that a human, a wallet, an agent, and a sequenced dApp can share:

1. **Name** — `.kas` as the public handle (KNS inscriptions today, covenant names later).
2. **Talk** — KaChat envelopes addressed to that name (encryption stays in KaChat).
3. **Secrets** — Kassword vault commitment hanging off the name (ciphertext stays in the browser).
4. **Rank** — live KAS depth as a local ocean rank (not the KASRANKS NFT).
5. **Call** — MCP + agent card so a machine can resolve, call, pay.
6. **Pay fees** — Work Credits: prepaid grams of sequenced work, not a fake dollar.

The running process is `kns.exe` at `:8080`. It is a resolver, a protocol surface, and a design. It is not the official KNS team, not KaChat, not a deployed registrar, and not a stablecoin issuer.

## Web4.0

Web4.0 here is the agentic internet: **readable, discoverable, callable, payable**. A `.kas` name already pays a human. The stack adds an agent card, MCP tools, and a Kaspa-native HTTP 402 so a machine can do the same.

## The money problem (why work credits exist)

dApps sequenced on Kaspa still have a cost problem:

- Users hate **USD volatility** in fees.
- Operators hate **KAS volatility** in their own opex if they priced in dollars.
- Classic stables (USDC, USDT, DAI) need **capital**: fiat reserves, crypto overcollateral, or a bank.
- Algorithmic “stables” (UST) mint a dollar from a story. That story dies.

Kaspa L1 cannot print dollars. A Toccata covenant cannot see Circle’s reserve. So the honest move is to stop pretending the missing object is a dollar, and ask what dApps actually consume.

They consume **sequenced work**: mass, gas, lane slots, inclusion.

Kaspa already prices that work in **sompi per gram** (policy today: 100 sompi/gram; KIP-21: 50 lanes/block, 1e9 gas/lane). That unit is already more stable *in KAS* than KAS is *in USD*.

**Work Credits** make that unit a prepaid voucher:

- 1 credit = 1 gram of a named lane.
- A covenant UTXO holds `issuer`, `holder`, `credits`, `lane`.
- dApps invoice in grams. Settlement is burn-with-sequencer-cosign, or KAS fallback.
- USD is a different dApp: Kaspa Till, reserved L1 kUSD. No L2. Not live. Still needs capital when it lands.

The user of a sequenced dApp sees a bill that does not move when KAS/USD moves. The operator ate the KAS inventory when they sold the voucher. That is a business, not a mint.

## Four-era identity (so names survive forks)

| Era | Object | Status |
| --- | --- | --- |
| 1 | Inscription + indexer | Live |
| 2 | Toccata Name UTXO / registrar | Consensus ready, app not deployed |
| 3 | Based ZK name set on a KIP-21 lane | Research + local `/sim` |
| 4 | vProgs composition + DAGKnight ordering | Socket, not a switch |

`kas://name.kas` and `did:kas:name.kas` are conventions of this repo.

## What success looks like

- An agent resolves `shop.kas`, reads the card, pays in KAS *or* burns grams.
- Today: Kasware `sendKaspa` of the quoted sompi (KAS fallback) is the L1 receipt. WorkCredit consume waits on a genesis UTXO.
- A sequenced dApp quotes `50_000 grams`. A shop quotes reserved kUSD and settles KAS. Neither uses an L2.
- A human still uses KaChat / Kassword / KasRanks as those products exist. This app points; it does not clone them.
- Nobody ships a UST on Kaspa because “covenants back it”.
