# How everything works

## 1. A name, today

Someone already inscribed `example.kas` with the official KNS envelope:

```json
{"op":"create","p":"domain","v":"example"}
```

Reveal output 0 paid the KNS fee address. The indexer (`api.knsdomains.org/mainnet`) maps that inscription to an owner `kaspa:` address and optional profile texts (avatar, website, x, redirectUrl, …).

`kns.exe` calls that API. It does not have its own name database for live names. Uniqueness is **first-come, first-served at the indexer**. Consensus does not know the string `example`.

Resolve:

```
GET /api/v1/resolve?q=example.kas
GET /name/example.kas
```

You get owner, pay URI, records, evidence tag `indexer`.

## 2. A name, after Toccata (not deployed)

Toccata made covenants real (KIP-16/17/20/21). A Name *can* be a UTXO whose script is `KasName.sil`:

- `owner`, `labelHash`, `kasPkh`, `vaultCommit`
- entries: `setKas`, `bindVault`, `transfer`

That UTXO is unique because its **outpoint** is unique (KIP-20 `covenant_id` = BLAKE2b of that outpoint). It is **not** unique because the label is `alice`. Anyone can genesis another covenant and write the same label. Nodes accept both.

So a real naming product still needs one of:

1. A **root registrar** that is the only minter of global `.kas` children, or
2. **Parent-issued subnames** (`pay.shop.kas` only by `shop.kas`), or
3. A **based ZK set** on a KIP-21 lane with a proof the name was free.

This app compiles `KasName.sil`. It does not submit the registrar.

## 3. Web 4.0 loop (agent)

```
resolve name  →  read card  →  call MCP  →  pay 402
```

- **Readable:** `/site/kns.kas` HTML for humans; `Accept: application/json` or `?format=json` for machines. `/llms.txt`.
- **Discoverable:** `/agent/kns.kas.json` is an ERC-8004 *registration file shape*. There is no Kaspa ERC-8004 contract. The file is synthesized from indexer records.
- **Callable:** `POST /mcp` with tools listed at `GET /mcp`.
- **Payable:** `GET /api/v1/call/kns.kas` returns HTTP 402. Body asks for KAS to the owner, and (now) also advertises `kaspa-work-credit` grams. Header `X-Kaspa-Payment: <txid>` is accepted at HTTP layer only — **not** verified on-chain.

This is Cloudflare’s 2026 agentic-internet list, not a Kaspa hard fork.

## 4. KaChat, Kassword, KasRanks, KCC

These are **other products**. The superapp’s job is a shared name and honest pointers.

| Product | Real thing | What `:8080` does |
| --- | --- | --- |
| KaChat | Client encrypts `ciph_msg:1:` / `kchat:1:` | Resolve name → owner address as contact. Envelope *shapes*. No keys. |
| Kassword | Browser vault, optional DAG backup | Page + `vaultCommit` record convention. No ciphertext. |
| KasRanks | NFT collection + site | Ocean rank from `api.kaspa.org` balance. Not token id lookup. |
| KCC | `kaspanet/kccs` drafts | List KCC-0/1/2/20. Local ideas are not submitted drafts. |

KaPosts on `/kaposts` is an in-process board so the UI has a feed. It is not KaChat’s indexer.

## 5. Fees on L1 (why grams)

A Kaspa tx has **mass** in compute / storage / transient grams. Relayers currently want at least **100 sompi per gram** (policy). KIP-21 adds **lanes**: 50 per block, 1e9 gas each. A sequenced dApp’s bill is some number of grams on some lane.

That bill, in KAS, is:

```
sompi = grams × sompiPerGram
```

At policy 100, **1,000,000 grams = 1 KAS**.

USD(bill) = KAS(bill) × USD(KAS). The second factor is the volatility people feel. The first factor is already a stable *work* unit.

## 6. Work Credits (the stable-shaped object)

See [WORK-CREDITS.md](WORK-CREDITS.md). Short version:

1. Sequencer/dApp operator **sells grams** for KAS now.
2. A `WorkCredit` UTXO records `issuer`, `holder`, `credits`, `lane`.
3. Later, user work is paid by `consume(holderSig, issuerSig, used)` — credits fall.
4. dApp invoices always say `N grams`. Dollars are a different dApp (Kaspa Till), not this one, and not an L2.

The covenant enforces the **ledger**. The issuer is the **counterparty** (they must still pay miners in KAS). No Circle, no DAI vault, no rebase.

Quote without chain:

```
GET /api/v1/credits/quote?grams=50000
GET /credits?grams=50000
MCP tool quote_work
```

## 7. Two dApps, no L2

- **Gramlane** (`:8081`, this folder) — jobs in grams. The alternative that is possible today.
- **Kaspa Till** (`:8082`, `Documents\kaspa\superappstablesalternative`) — shelf in reserved `kUSD`. Vision of a native L1 stable. Settle in KAS now.

L2 bridged dollars exist in the wild. This stack does not use them.

## 8. What is in memory vs on chain

| Data | Where |
| --- | --- |
| Live name owner/profile | KNS indexer |
| Live KAS balance / rank | api.kaspa.org |
| Agent cards, 402, MCP | synthesized in `kns.exe` |
| Local name-set `/sim` | process memory, toy root |
| Local KaPosts | process memory |
| Covenant/vault JSON under `kns/data` | local files, not mainnet |
| Compiled `.json` artifacts | embedded in the binary |

Restarting the process wipes `/sim` and `/kaposts`. It does not wipe KNS names.

## 9. Compiler path

```powershell
C:\Users\Remco\tools\silverc\silverc.exe `
  C:\Users\Remco\kns\contracts\v1\WorkCredit.sil `
  --constructor-args C:\Users\Remco\kns\contracts\v1\ctor-WorkCredit.json `
  -o C:\Users\Remco\kns\contracts\v1\WorkCredit.json
```

Or `C:\Users\Remco\kns\compile-sil.ps1`. Artifacts are served from `kns/contracts` embed.
