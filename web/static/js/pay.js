(function () {
  function $(sel) {
    return document.querySelector(sel);
  }

  function say(t) {
    const el = $("[data-pay-status]");
    if (el) el.textContent = t;
  }

  function copy(text) {
    text = String(text || "").trim();
    if (!text || text.indexOf("(") === 0) {
      say("Nothing to copy yet — Log in first for the address.");
      return;
    }
    function fallback() {
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      ta.setSelectionRange(0, text.length);
      var ok = false;
      try {
        ok = document.execCommand("copy");
      } catch (_) {}
      document.body.removeChild(ta);
      say(ok ? "Copied " + text : "Copy failed. Select: " + text);
    }
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).then(
        function () {
          say("Copied " + text);
        },
        fallback
      );
    } else {
      fallback();
    }
  }

  function bind() {
    const host = $("[data-origin-note]");
    if (host) {
      host.textContent = "Stay on " + location.origin + " (localhost ≠ 127.0.0.1 for Kasware).";
    }
    document.querySelectorAll("[data-copy], [data-copy-text]").forEach(function (btn) {
      btn.addEventListener("click", function (ev) {
        ev.preventDefault();
        const sel = btn.getAttribute("data-copy");
        const el = sel ? document.querySelector(sel) : null;
        const direct = btn.getAttribute("data-copy-text");
        copy(direct || (el ? el.value || el.textContent : ""));
      });
    });
    const exp = $("[data-try-dapp-send]");
    if (exp) {
      exp.addEventListener("click", function (ev) {
        ev.preventDefault();
        say("dApp sendKaspa is what froze Kasware. Use Send inside the Kasware extension, amount 0.012, then paste the txid.");
      });
    }
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
