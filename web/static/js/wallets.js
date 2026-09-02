(function () {
  const KEY_ADDR = "kaspaAddress";
  const KEY_WALLET = "kaspaWallet";

  function shortAddr(a) {
    if (!a) return "";
    if (a.length <= 18) return a;
    return a.slice(0, 10) + "…" + a.slice(-8);
  }

  function detected() {
    const d = [];
    if (typeof window.kasware !== "undefined") d.push("kasware");
    if (typeof window.kastle !== "undefined") d.push("kastle");
    return d;
  }

  async function connectKasware() {
    const w = window.kasware;
    if (!w) {
      window.open("https://www.kasware.xyz", "_blank");
      throw new Error("Kasware is not installed.");
    }
    const acc = await w.requestAccounts();
    if (!acc || !acc[0]) throw new Error("Kasware returned no account.");
    return { id: "kasware", address: acc[0] };
  }

  async function connectKastle() {
    const w = window.kastle;
    if (!w) {
      window.open("https://kastle.cc", "_blank");
      throw new Error("Kastle is not installed.");
    }
    const ok = await w.connect();
    if (!ok) throw new Error("Kastle connect was declined.");
    const acc = await w.getAccount();
    const address = acc && (acc.address || acc);
    if (!address) throw new Error("Kastle returned no account.");
    return { id: "kastle", address: String(address) };
  }

  function persist(id, address) {
    try {
      sessionStorage.setItem(KEY_ADDR, address);
      sessionStorage.setItem(KEY_WALLET, id);
    } catch (_) {}
    window.dispatchEvent(new CustomEvent("kaspa-wallet", { detail: { id, address } }));
  }

  function clearSession() {
    try {
      sessionStorage.removeItem(KEY_ADDR);
      sessionStorage.removeItem(KEY_WALLET);
    } catch (_) {}
    window.dispatchEvent(new CustomEvent("kaspa-wallet", { detail: { id: "", address: "" } }));
  }

  async function logout() {
    const c = current();
    try {
      if (c.id === "kasware" && window.kasware && window.kasware.disconnect) {
        await window.kasware.disconnect(location.origin);
      }
    } catch (_) {}
    try {
      if (c.id === "kastle" && window.kastle && window.kastle.disconnect) {
        await window.kastle.disconnect();
      }
    } catch (_) {}
    clearSession();
    paintButtons();
  }

  function current() {
    try {
      return {
        id: sessionStorage.getItem(KEY_WALLET) || "",
        address: sessionStorage.getItem(KEY_ADDR) || "",
      };
    } catch (_) {
      return { id: "", address: "" };
    }
  }

  function paintButtons() {
    const c = current();
    document.querySelectorAll("[data-wallet-connect]").forEach((btn) => {
      if (c.address) {
        btn.textContent = shortAddr(c.address);
        btn.dataset.connected = c.id;
      } else {
        btn.textContent = btn.getAttribute("data-idle-label") || "Connect wallet";
        delete btn.dataset.connected;
      }
    });
    document.querySelectorAll("[data-wallet-addr]").forEach((el) => {
      el.textContent = c.address || "";
    });
    document.querySelectorAll("[data-wallet-logout]").forEach((btn) => {
      btn.hidden = !c.address;
    });
    const addrInput = document.querySelector('input[name="address"]');
    if (c.address && addrInput && !addrInput.value) addrInput.value = c.address;
  }

  async function connect(id) {
    let r;
    if (id === "kasware") r = await connectKasware();
    else if (id === "kastle") r = await connectKastle();
    else throw new Error("This wallet has no in-page provider. Open or install it from the catalog.");
    persist(r.id, r.address);
    paintButtons();
    const onMe = location.pathname.startsWith("/me");
    const onWallets = location.pathname.startsWith("/wallets");
    if (!onMe && !onWallets && document.querySelector('a[href="/me"]')) {
      location.href = "/me?address=" + encodeURIComponent(r.address);
    }
    return r;
  }

  function closeModal() {
    const m = document.getElementById("walletModal");
    if (m) m.hidden = true;
  }

  function openModal() {
    const m = document.getElementById("walletModal");
    if (!m) return;
    const det = detected();
    m.querySelectorAll("[data-detect]").forEach((el) => {
      const id = el.getAttribute("data-detect");
      el.hidden = det.indexOf(id) === -1;
    });
    m.querySelectorAll("[data-missing]").forEach((el) => {
      const id = el.getAttribute("data-missing");
      el.hidden = det.indexOf(id) !== -1;
    });
    m.hidden = false;
    paintButtons();
  }

  function bind() {
    if (!document.getElementById("walletModal")) {
      const wrap = document.createElement("div");
      wrap.innerHTML = `
<div id="walletModal" class="wmodal" hidden>
  <div class="wmodal-card">
    <div class="wmodal-head">
      <strong>Connect a Kaspa wallet</strong>
      <button type="button" class="btn ghost" data-wallet-close>Close</button>
    </div>
    <p class="tiny">Only Kasware and Kastle inject a Kaspa L1 provider into this page. Ledger signs in KasVault (or via Kastle). Everything else is install/open — we do not fake a connect.</p>
    <p class="tiny"><strong>This site never DMs you.</strong> We never ask for a seed, key, or password. <a href="/safety">Safety</a></p>
    <div class="row" style="margin-top:12px">
      <button type="button" class="btn mint" data-wallet-id="kasware">Connect Kasware <span data-detect="kasware" hidden class="chip live">detected</span></button>
      <button type="button" class="btn mint" data-wallet-id="kastle">Connect Kastle <span data-detect="kastle" hidden class="chip live">detected</span></button>
      <a class="btn ghost" href="https://kasvault.io" target="_blank" rel="noopener">Ledger → KasVault</a>
      <button type="button" class="btn ghost" data-wallet-logout hidden>Log out</button>
    </div>
    <p class="tiny" data-missing="kasware">Kasware not detected. <a href="https://www.kasware.xyz" target="_blank" rel="noopener">Install</a></p>
    <p class="tiny" data-missing="kastle">Kastle not detected. <a href="https://kastle.cc" target="_blank" rel="noopener">Install</a></p>
    <p><a href="/wallets">Full catalog (hardware, mobile, multi-chain)</a></p>
  </div>
</div>`;
      document.body.appendChild(wrap.firstElementChild);
    }
    document.querySelectorAll("[data-wallet-connect]").forEach((btn) => {
      btn.addEventListener("click", () => {
        openModal();
      });
    });
    document.addEventListener("click", (e) => {
      const close = e.target.closest("[data-wallet-close]");
      if (close) closeModal();
      const out = e.target.closest("[data-wallet-logout]");
      if (out) {
        logout().then(() => closeModal()).catch((err) => alert(err.message || err));
        return;
      }
      const pick = e.target.closest("[data-wallet-id]");
      if (pick) {
        connect(pick.getAttribute("data-wallet-id")).then(() => closeModal()).catch((err) => alert(err.message || err));
      }
      if (e.target.id === "walletModal") closeModal();
    });
    paintButtons();
  }

  window.KaspaWallets = { connect, logout, current, detected, paintButtons, openModal };

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
