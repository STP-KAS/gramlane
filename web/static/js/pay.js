(function () {
  function $(sel) {
    return document.querySelector(sel);
  }

  function say(t) {
    const el = $("[data-pay-status]");
    if (el) el.textContent = t;
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

  async function accounts() {
    const w = window.kasware;
    if (!w) throw new Error("Kasware not injected. Chrome + unlocked Kasware on this exact URL.");
    if (typeof w.getAccounts === "function") {
      const have = await w.getAccounts();
      if (have && have[0]) return have;
    }
    if (typeof w.requestAccounts !== "function") throw new Error("Kasware has no requestAccounts");
    return w.requestAccounts();
  }

  async function payL1(ev) {
    ev.preventDefault();
    ev.stopPropagation();
    const btn = ev.currentTarget;
    const sompi = Number(btn.getAttribute("data-sompi") || "0");
    const pay = $('input[name="payment"]');
    const payer = $('input[name="payer"]');
    const wallet = $('input[name="wallet"]');
    const link = $("[data-explorer]");
    if (!sompi || sompi < 1) {
      say("No quote sompi.");
      return;
    }
    if (btn.dataset.busy === "1") return;
    btn.dataset.busy = "1";
    btn.disabled = true;
    say("One Kasware window only. If it is already black: close it, click the Kasware icon, unlock, then try this button again.");
    try {
      const acc = await accounts();
      const to = acc && acc[0];
      if (!to) throw new Error("Kasware returned no address. Unlock the wallet.");
      if (payer) payer.value = to;
      if (wallet) wallet.value = "kasware";
      // No payload, no second popup, no page navigation — those black-screen Kasware.
      const raw = await window.kasware.sendKaspa(to, sompi, { priorityFee: 10000 });
      const txid = parseTxid(raw);
      if (!txid) throw new Error("Kasware returned no txid");
      if (pay) pay.value = txid;
      if (link) {
        link.href = "https://explorer.kaspa.org/txs/" + txid;
        link.hidden = false;
        link.textContent = "Open on explorer.kaspa.org";
      }
      say("Broadcast. Txid " + txid + ". Now click Run job — do not reload.");
    } catch (e) {
      const msg = e && e.message ? e.message : String(e);
      say(msg + " — close the black Kasware window, unlock the extension, retry. Use http://127.0.0.1:8081 (not localhost).");
    } finally {
      btn.dataset.busy = "0";
      btn.disabled = false;
    }
  }

  function bind() {
    const btn = $("[data-pay-l1]");
    if (btn) btn.addEventListener("click", payL1);
    const host = $("[data-origin-note]");
    if (host) {
      host.textContent =
        "This tab origin is " + location.origin + ". Kasware treats localhost and 127.0.0.1 as different sites. Stay on this one.";
    }
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
