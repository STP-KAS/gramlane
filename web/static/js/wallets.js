(function () {
  const KEY_ADDR = "kaspaAddress";
  const KEY_WALLET = "kaspaWallet";

  function shortAddr(a) {
    if (!a) return "";
    if (a.length <= 18) return a;
    return a.slice(0, 10) + "…" + a.slice(-8);
  }

  function sleep(ms) {
    return new Promise(function (r) {
      setTimeout(r, ms);
    });
  }

  function withTimeout(p, ms, label) {
    return Promise.race([
      p,
      sleep(ms).then(function () {
        throw new Error(label || "wallet timed out");
      }),
    ]);
  }

  function detected() {
    const d = [];
    if (typeof window.kasware !== "undefined") d.push("kasware");
    if (typeof window.kastle !== "undefined") d.push("kastle");
    return d;
  }

  function status(msg) {
    let el = document.querySelector("[data-wallet-status]");
    if (!el) {
      const bar = document.querySelector("header.nav");
      if (!bar) return;
      el = document.createElement("span");
      el.className = "tiny";
      el.setAttribute("data-wallet-status", "");
      bar.appendChild(el);
    }
    el.textContent = msg || "";
  }

  function persist(id, address) {
    try {
      sessionStorage.setItem(KEY_ADDR, address);
      sessionStorage.setItem(KEY_WALLET, id);
    } catch (_) {}
    window.dispatchEvent(new CustomEvent("kaspa-wallet", { detail: { id: id, address: address } }));
  }

  function clearSession() {
    try {
      sessionStorage.removeItem(KEY_ADDR);
      sessionStorage.removeItem(KEY_WALLET);
    } catch (_) {}
    window.dispatchEvent(new CustomEvent("kaspa-wallet", { detail: { id: "", address: "" } }));
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
    document.querySelectorAll("[data-wallet-connect]").forEach(function (btn) {
      if (c.address) {
        btn.textContent = "Logged in";
        btn.title = c.address;
        btn.dataset.connected = c.id;
      } else {
        btn.textContent = btn.getAttribute("data-idle-label") || "Log in";
        btn.removeAttribute("title");
        delete btn.dataset.connected;
      }
    });
    document.querySelectorAll("[data-wallet-addr]").forEach(function (el) {
      if (el.tagName === "INPUT" || el.tagName === "TEXTAREA") el.value = c.address || "";
      else el.textContent = c.address || "";
    });
    document.querySelectorAll("[data-wallet-logout]").forEach(function (btn) {
      btn.hidden = !c.address;
    });
    const addrInput = document.querySelector('input[name="address"]');
    if (c.address && addrInput && !addrInput.value) addrInput.value = c.address;
    const payer = document.querySelector('input[name="payer"]');
    if (payer) payer.value = c.address || "";
    const wallet = document.querySelector('input[name="wallet"]');
    if (wallet) wallet.value = c.id || "";
    if (c.address) status("");
  }

  async function kaswareAccountsQuiet() {
    const w = window.kasware;
    if (!w || typeof w.getAccounts !== "function") return [];
    try {
      const acc = await withTimeout(w.getAccounts(), 2500, "getAccounts timeout");
      return acc && acc.length ? acc : [];
    } catch (_) {
      return [];
    }
  }

  async function connectKasware() {
    closeModal();
    let w = window.kasware;
    if (!w) {
      for (let i = 0; i < 8 && !w; i++) {
        await sleep(120 * (i + 1));
        w = window.kasware;
      }
    }
    if (!w) {
      window.open("https://www.kasware.xyz", "_blank");
      throw new Error("Kasware is not in this tab. Install it, unlock it, stay on http://127.0.0.1:8081");
    }
    const quiet = await kaswareAccountsQuiet();
    if (quiet[0]) return { id: "kasware", address: quiet[0] };
    status("Kasware: approve Log in. If the window is black, close it, click the Kasware icon, unlock, try again.");
    const acc = await withTimeout(w.requestAccounts(), 45000, "Kasware Log in timed out (black window?). Close it and retry.");
    if (!acc || !acc[0]) throw new Error("Kasware returned no account.");
    return { id: "kasware", address: acc[0] };
  }

  async function connectKastle() {
    closeModal();
    const w = window.kastle;
    if (!w) {
      window.open("https://kastle.cc", "_blank");
      throw new Error("Kastle is not installed.");
    }
    const ok = await withTimeout(w.connect(), 45000, "Kastle connect timed out.");
    if (!ok) throw new Error("Kastle connect was declined.");
    const acc = await w.getAccount();
    const address = acc && (acc.address || acc);
    if (!address) throw new Error("Kastle returned no account.");
    return { id: "kastle", address: String(address) };
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
    status("");
  }

  async function connect(id) {
    let r;
    if (id === "kasware") r = await connectKasware();
    else if (id === "kastle") r = await connectKastle();
    else throw new Error("This wallet has no in-page provider.");
    persist(r.id, r.address);
    paintButtons();
    status("Logged in");
    return r;
  }

  function preferredWallet() {
    const d = detected();
    if (d.indexOf("kasware") !== -1) return "kasware";
    if (d.indexOf("kastle") !== -1) return "kastle";
    return "";
  }

  function closeModal() {
    const m = document.getElementById("walletModal");
    if (m) m.hidden = true;
  }

  function openModal() {
    const m = document.getElementById("walletModal");
    if (!m) return;
    const det = detected();
    m.querySelectorAll("[data-detect]").forEach(function (el) {
      const id = el.getAttribute("data-detect");
      el.hidden = det.indexOf(id) === -1;
    });
    m.querySelectorAll("[data-missing]").forEach(function (el) {
      const id = el.getAttribute("data-missing");
      el.hidden = det.indexOf(id) !== -1;
    });
    m.hidden = false;
  }

  async function clickLogin(btn) {
    if (current().address) {
      status("Logged in · " + shortAddr(current().address));
      return;
    }
    const id = preferredWallet();
    if (!id) {
      openModal();
      status("No wallet in this tab. Install Kasware, then Log in.");
      return;
    }
    btn.disabled = true;
    try {
      await connect(id);
    } catch (err) {
      status(err && err.message ? err.message : String(err));
    } finally {
      btn.disabled = false;
    }
  }

  async function resume() {
    const quiet = await kaswareAccountsQuiet();
    if (quiet[0]) {
      persist("kasware", quiet[0]);
      paintButtons();
      return;
    }
    paintButtons();
  }

  function bind() {
    if (!document.getElementById("walletModal")) {
      const wrap = document.createElement("div");
      wrap.innerHTML =
        '<div id="walletModal" class="wmodal" hidden><div class="wmodal-card">' +
        '<div class="wmodal-head"><strong>Log in with a Kaspa wallet</strong>' +
        '<button type="button" class="btn ghost" data-wallet-close>Close</button></div>' +
        '<p class="tiny">Kasware or Kastle. This site never asks for a seed.</p>' +
        '<div class="row" style="margin-top:12px">' +
        '<button type="button" class="btn mint" data-wallet-id="kasware">Kasware</button>' +
        '<button type="button" class="btn mint" data-wallet-id="kastle">Kastle</button>' +
        '<a class="btn ghost" href="/wallets">Catalog</a></div>' +
        "</div></div>";
      document.body.appendChild(wrap.firstElementChild);
    }
    document.querySelectorAll("[data-wallet-connect]").forEach(function (btn) {
      btn.addEventListener("click", function (e) {
        e.preventDefault();
        clickLogin(btn);
      });
    });
    document.addEventListener("click", function (e) {
      if (e.target.closest("[data-wallet-close]")) closeModal();
      const out = e.target.closest("[data-wallet-logout]");
      if (out) {
        logout().catch(function (err) {
          status(err.message || String(err));
        });
        return;
      }
      const pick = e.target.closest("[data-wallet-id]");
      if (pick) {
        connect(pick.getAttribute("data-wallet-id")).then(closeModal).catch(function (err) {
          status(err.message || String(err));
        });
      }
      if (e.target.id === "walletModal") closeModal();
    });
    resume();
  }

  window.KaspaWallets = { connect: connect, logout: logout, current: current, detected: detected, paintButtons: paintButtons };

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", bind);
  else bind();
})();
