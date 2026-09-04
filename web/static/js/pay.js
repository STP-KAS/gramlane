(function () {
  function $(sel) {
    return document.querySelector(sel);
  }

  function say(t) {
    const el = $("[data-pay-status]");
    if (el) el.textContent = t;
  }

  function loginAddr() {
    var el = document.querySelector("[data-wallet-addr]");
    if (!el) return "";
    return String(el.value || el.textContent || "").trim();
  }

  function parseTxid(raw) {
    if (!raw) return "";
    if (typeof raw === "object") return raw.id || raw.transactionId || raw.txid || "";
    var s = String(raw).trim();
    try {
      var j = JSON.parse(s);
      return j.id || j.transactionId || j.txid || s;
    } catch (_) {
      return s;
    }
  }

  function copy(text) {
    text = String(text || "").trim();
    if (!text) return;
    function fallback() {
      var ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      try {
        document.execCommand("copy");
      } catch (_) {}
      document.body.removeChild(ta);
      say("Copied " + text);
    }
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).then(function () {
        say("Copied " + text);
      }, fallback);
    } else fallback();
  }

  async function accounts() {
    var w = window.kasware;
    if (!w) throw new Error("Kasware is not in this tab. Log in on http://127.0.0.1:8081 first.");
    if (typeof w.getAccounts === "function") {
      try {
        var have = await w.getAccounts();
        if (have && have[0]) return have;
      } catch (_) {}
    }
    return w.requestAccounts();
  }

  async function payL1(ev) {
    ev.preventDefault();
    ev.stopPropagation();
    var btn = ev.currentTarget;
    if (btn.dataset.busy === "1") return;
    var to = (btn.getAttribute("data-to") || "").trim() || (($("#pay-to") && $("#pay-to").value.trim()) || "");
    var sompi = Number(btn.getAttribute("data-sompi") || "0");
    var me = loginAddr();
    if (!to || to.indexOf("kaspa:") !== 0) {
      say("Paste a desk kaspa: address first (not your login).");
      return;
    }
    if (me && to === me) {
      say("Desk address is your login. Kasware will block it. Use a different kaspa:.");
      return;
    }
    if (!sompi) {
      say("Missing amount.");
      return;
    }
    var modal = document.getElementById("walletModal");
    if (modal) modal.hidden = true;
    btn.dataset.busy = "1";
    btn.disabled = true;
    say("Kasware confirm should appear over this page — stay here, do not open the extension Send tab.");
    try {
      var acc = await accounts();
      if (!acc || !acc[0]) throw new Error("Log in first.");
      var payer = $('input[name="payer"]');
      var wallet = $('input[name="wallet"]');
      if (payer) payer.value = acc[0];
      if (wallet) wallet.value = "kasware";
      var raw = await window.kasware.sendKaspa(to, sompi, { priorityFee: 10000 });
      var txid = parseTxid(raw);
      if (!txid) throw new Error("Kasware returned no txid.");
      var pay = $('input[name="payment"]');
      if (pay) pay.value = txid;
      var link = $("[data-explorer]");
      if (link) {
        link.href = "https://explorer.kaspa.org/txs/" + txid;
        link.hidden = false;
        link.textContent = "explorer.kaspa.org";
      }
      say("On L1. Txid " + txid);
      var buyId = btn.getAttribute("data-buy-id");
      if (buyId && acc[0]) {
        var buyBody = "act=buy&id=" + encodeURIComponent(buyId) + "&buyer=" + encodeURIComponent(acc[0]) + "&tx=" + encodeURIComponent(txid);
        fetch("/market", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" }, body: buyBody })
          .finally(function () {
            location.href = "/mine?address=" + encodeURIComponent(acc[0]);
          });
        return;
      }
      var nm = btn.getAttribute("data-name");
      if (nm && acc[0]) {
        var body = "act=held&name=" + encodeURIComponent(nm) + "&address=" + encodeURIComponent(acc[0]) + "&tx=" + encodeURIComponent(txid);
        fetch("/kasdomain", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" }, body: body })
          .finally(function () {
            location.href = "/kasdomain?q=" + encodeURIComponent(nm) + "&address=" + encodeURIComponent(acc[0]);
          });
        return;
      }
      var after = btn.getAttribute("data-after");
      if (after === "run") {
        say("On L1. Txid " + txid + " — running job.");
        var form = document.querySelector("form[action='/run']");
        if (form) form.submit();
      } else if (after) {
        var form = document.querySelector("form[action='" + after + "']") || document.querySelector(after);
        if (form) form.submit();
      }
    } catch (e) {
      say((e && e.message ? e.message : String(e)) + " — if Kasware is a black window, close it, Log in, click Pay once.");
    } finally {
      btn.dataset.busy = "0";
      btn.disabled = false;
    }
  }

  function bind() {
    var host = $("[data-origin-note]");
    if (host) host.textContent = "Stay on " + location.origin + ".";
    document.querySelectorAll("[data-pay-l1]").forEach(function (payBtn) {
      payBtn.addEventListener("click", payL1);
    });
    document.querySelectorAll("[data-copy], [data-copy-text]").forEach(function (btn) {
      btn.addEventListener("click", function (ev) {
        ev.preventDefault();
        var sel = btn.getAttribute("data-copy");
        var el = sel ? document.querySelector(sel) : null;
        var direct = btn.getAttribute("data-copy-text");
        copy(direct || (el ? el.value || el.textContent : ""));
      });
    });
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
