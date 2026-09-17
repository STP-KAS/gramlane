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
      say("Copied");
    }
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).then(function () {
        say("Copied");
      }, fallback);
    } else fallback();
  }
  function withdrawn(ev) {
    ev.preventDefault();
    ev.stopPropagation();
    say("Wallet inject withdrawn 17 Sep 2026. Copy the kaspa: URI, scan the QR, or paste a txid. This desk does not ship wallet integrations.");
  }
  function bind() {
    document.querySelectorAll("[data-pay-l1], [data-kachat-send]").forEach(function (btn) {
      btn.addEventListener("click", withdrawn);
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
