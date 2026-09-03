(function () {
  function $(sel) {
    return document.querySelector(sel);
  }

  function say(t) {
    const el = $("[data-pay-status]");
    if (el) el.textContent = t;
  }

  function copy(text) {
    if (!text) return;
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(
        function () {
          say("Copied.");
        },
        function () {
          say("Copy failed — select the field and Ctrl+C.");
        }
      );
    }
  }

  function bind() {
    const host = $("[data-origin-note]");
    if (host) {
      host.textContent = "Stay on " + location.origin + " (localhost ≠ 127.0.0.1 for Kasware).";
    }
    document.querySelectorAll("[data-copy]").forEach(function (btn) {
      btn.addEventListener("click", function () {
        const sel = btn.getAttribute("data-copy");
        const el = sel ? document.querySelector(sel) : null;
        const direct = btn.getAttribute("data-copy-text");
        copy(direct || (el ? el.textContent || el.value : ""));
      });
    });
    const exp = $("[data-try-dapp-send]");
    if (exp) {
      exp.addEventListener("click", function (ev) {
        ev.preventDefault();
        say("dApp sendKaspa is what froze Kasware. Use Send inside the Kasware extension instead, then paste the txid.");
      });
    }
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
