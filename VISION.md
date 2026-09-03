# Gramlane vision

**Stable work price on L1, not a synthetic dollar.**

Kaspa stays issuer-free. Gramlane is not a dollar and does not replace USDT.

It is an L1 work-credit lane. dApps invoice jobs in **grams** (KIP-21 mass), not USD. The fee for a named job stays the same in work units: lookup, vault bump, message postage, data pull, agent call. KAS is only the fallback if the user has no prepaid credits.

Centralized stables can exist on L2. They do not sit in Gramlane’s path. No mint from Tether. No freeze from Circle. No quoting FX as if it were protocol cost.

## Use it where the scarce thing is work, not rent

| Work | What grams pay | What grams are not |
| --- | --- | --- |
| AI bots / HTTP 402 | The sequenced call | The model’s output, the dollar the bot might later spend |
| Messaging postage | Inclusion of the envelope | The message, the chat app, encryption |
| Data transfer | The pull / the write | The dataset’s market price |
| Vault operations | The bump, the pin, the spend path | The KAS locked in the vault |

Vaults still lock KAS. Grams pay the **action**.

## What this is not

- A unit of account for humans (nobody prices coffee in grams).
- A savings asset (credits are inventory of work, not a store of value).
- A claim that “stables are banned.” Liquidity will still touch USDT off to the side. Gramlane just refuses to make that the **meter**.

A Toccata covenant can conserve grams. It cannot peg `$1`. The operator who sold the voucher ate the KAS inventory. That is a business, not a mint.

## Live on this dApp

Prepaid grams from a 0.5 KAS L1 sale, with a funded WorkCredit P2SH. Jobs burn that inventory. `consume()` of the UTXO is a later 2-sig spend — operator accounting until then. USD is never quoted.

| Path | Work |
| --- | --- |
| `/vault` | Vault bump (not the lock) |
| `/postage` | Message postage |
| `/agent` | AI agent call (Grok if `XAI_API_KEY`, else local tools) |
| `/site` | .kas site builder from the live indexer |
