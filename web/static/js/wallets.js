(function () {
  function say(msg) {
    document.querySelectorAll("[data-wallet-status]").forEach(function (el) {
      el.textContent = msg || "";
    });
  }
  function paint() {
    document.querySelectorAll("[data-wallet-connect]").forEach(function (btn) {
      btn.textContent = btn.getAttribute("data-idle-label") || "Pay with QR / kaspa: URI";
    });
    document.querySelectorAll("[data-wallet-logout]").forEach(function (btn) {
      btn.hidden = true;
    });
  }
  window.KaspaWallets = {
    connect: function () {
      throw new Error("withdrawn: this desk does not ship wallet inject. Use QR, kaspa: URI, or paste a txid.");
    },
    logout: function () {},
    current: function () {
      return { id: "", address: "" };
    },
    detected: function () {
      return [];
    },
    paintButtons: paint,
  };
  function bind() {
    paint();
    document.querySelectorAll("[data-wallet-connect]").forEach(function (btn) {
      btn.addEventListener("click", function (e) {
        e.preventDefault();
        say("Wallet inject withdrawn 17 Sep 2026. Pay with QR, kaspa: URI, or paste a txid.");
      });
    });
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
