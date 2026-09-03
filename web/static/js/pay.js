(function () {
  function $(sel) {
    return document.querySelector(sel);
  }

  function parseTxid(raw) {
    if (!raw) return "";
    if (typeof raw === "object") {
      return raw.id || raw.transactionId || raw.txid || "";
    }
    const s = String(raw).trim();
    try {
      const j = JSON.parse(s);
      return j.id || j.transactionId || j.txid || s;
    } catch (_) {
      return s;
    }
  }

  async function refreshBalance() {
    const el = $("[data-kas-balance]");
    if (!el || !window.kasware || !window.kasware.getBalance) return;
    try {
      const b = await window.kasware.getBalance();
      if (!b || b.total == null) {
        el.textContent = "balance: (unlock Kasware)";
        return;
      }
      const kas = Number(b.total) / 1e8;
      el.textContent = "Kasware balance: " + kas.toFixed(8) + " KAS";
    } catch (e) {
      el.textContent = "balance: " + (e.message || e);
    }
  }

  async function payL1(ev) {
    const btn = ev.currentTarget;
    const sompi = Number(btn.getAttribute("data-sompi") || "0");
    const job = btn.getAttribute("data-job") || "";
    const grams = btn.getAttribute("data-grams") || "";
    const status = $("[data-pay-status]");
    const pay = $('input[name="payment"]');
    const payer = $('input[name="payer"]');
    const wallet = $('input[name="wallet"]');
    const link = $("[data-explorer]");
    function say(t) {
      if (status) status.textContent = t;
    }
    if (!window.kasware) {
      say("Kasware is not in this page. Open this URL in Chrome with Kasware unlocked.");
      return;
    }
    if (!sompi || sompi < 1) {
      say("No quote sompi.");
      return;
    }
    btn.disabled = true;
    say("Kasware popup: approve the send. This is the KAS fallback, to your own address. WorkCredit consume is not this.");
    try {
      const acc = await window.kasware.requestAccounts();
      const to = acc && acc[0];
      if (!to) throw new Error("no account");
      if (payer) payer.value = to;
      if (wallet) wallet.value = "kasware";
      const payload = "gramlane:" + job + ":" + grams;
      const raw = await window.kasware.sendKaspa(to, sompi, { payload: payload });
      const txid = parseTxid(raw);
      if (!txid) throw new Error("Kasware returned no txid");
      if (pay) pay.value = txid;
      if (link) {
        link.href = "https://explorer.kaspa.org/txs/" + txid;
        link.hidden = false;
        link.textContent = "explorer.kaspa.org/txs/" + txid.slice(0, 12) + "…";
      }
      say("Broadcast. Txid " + txid + " — that is on L1. Click Run job.");
      const form = btn.closest("form") || document.querySelector("form[action='/run']");
      if (form) form.submit();
    } catch (e) {
      say(e && e.message ? e.message : String(e));
      btn.disabled = false;
    }
  }

  function bind() {
    const btn = $("[data-pay-l1]");
    if (btn) btn.addEventListener("click", payL1);
    refreshBalance();
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
