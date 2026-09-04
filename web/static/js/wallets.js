(function () {
  const KEY_ADDR = "kaspaAddress";
  const KEY_WALLET = "kaspaWallet";
  const KEY_NAME = "kaspaName";

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

  function persist(id, address, name) {
    try {
      sessionStorage.setItem(KEY_ADDR, address);
      sessionStorage.setItem(KEY_WALLET, id);
      if (name) sessionStorage.setItem(KEY_NAME, name);
      else sessionStorage.removeItem(KEY_NAME);
    } catch (_) {}
    window.dispatchEvent(new CustomEvent("kaspa-wallet", { detail: { id: id, address: address, name: name || "" } }));
  }

  function clearSession() {
    try {
      sessionStorage.removeItem(KEY_ADDR);
      sessionStorage.removeItem(KEY_WALLET);
      sessionStorage.removeItem(KEY_NAME);
    } catch (_) {}
    window.dispatchEvent(new CustomEvent("kaspa-wallet", { detail: { id: "", address: "", name: "" } }));
  }

  function current() {
    try {
      return {
        id: sessionStorage.getItem(KEY_WALLET) || "",
        address: sessionStorage.getItem(KEY_ADDR) || "",
        name: sessionStorage.getItem(KEY_NAME) || "",
      };
    } catch (_) {
      return { id: "", address: "", name: "" };
    }
  }

  function display(c) {
    if (c && c.name) return c.name;
    if (c && c.address) return shortAddr(c.address);
    return "";
  }

  function paintButtons() {
    const c = current();
    document.querySelectorAll("[data-wallet-connect]").forEach(function (btn) {
      if (c.address) {
        btn.hidden = true;
        btn.textContent = display(c);
        btn.title = c.name ? c.name + " · " + c.address : c.address;
        btn.dataset.connected = c.id;
      } else {
        btn.hidden = false;
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
    var who = document.querySelector("[data-who]");
    if (who) {
      who.hidden = !c.address;
      var nm = who.querySelector("[data-copy-who='name']");
      var ad = who.querySelector("[data-copy-who='addr']");
      if (nm) {
        nm.textContent = c.name || "no kasname";
        nm.title = c.name ? "Copy " + c.name : "No kasname yet — Change name";
      }
      if (ad) {
        ad.textContent = shortAddr(c.address);
        ad.title = c.address ? "Copy " + c.address : "";
      }
    }
    const addrInput = document.querySelector('input[name="address"]');
    if (c.address && addrInput && !addrInput.value) addrInput.value = c.address;
    const payer = document.querySelector('input[name="payer"]');
    if (payer) payer.value = c.address || "";
    const wallet = document.querySelector('input[name="wallet"]');
    if (wallet) wallet.value = c.id || "";
    if (!c.address) status("");
    document.querySelectorAll("[data-wallet-status]").forEach(function (el) {
      if (c.address) el.setAttribute("data-wallet-change-name", "1");
      else el.removeAttribute("data-wallet-change-name");
    });
    refreshBalances();
  }

  function fmtKas1(sompi) {
    var n = Number(sompi);
    if (!isFinite(n) || n < 0) n = 0;
    return (n / 100000000).toFixed(1);
  }

  function sompiFromBalance(b) {
    if (b == null) return 0;
    if (typeof b === "number") return b;
    if (typeof b === "string") return Number(b) || 0;
    return Number(b.total || b.confirmed || b.balance || b.amount || 0) || 0;
  }

  async function refreshBalances() {
    var box = document.querySelector("[data-balances]");
    if (!box) return;
    var c = current();
    if (!c.address) {
      box.hidden = true;
      return;
    }
    box.hidden = false;
    var kasEl = box.querySelector("[data-kas-balance]");
    var gEl = box.querySelector("[data-gram-balance]");
    try {
      var seq = await fetch("/api/seq").then(function (r) { return r.json(); });
      var left = seq && seq.data && seq.data.remaining;
      if (gEl && left != null) gEl.textContent = "jar " + Number(left).toLocaleString("en-US") + " g";
    } catch (_) {}
    try {
      var sompi = 0;
      if (c.id === "kasware" && window.kasware && typeof window.kasware.getBalance === "function") {
        sompi = sompiFromBalance(await window.kasware.getBalance());
      } else if (window.kastle && typeof window.kastle.getBalance === "function") {
        sompi = sompiFromBalance(await window.kastle.getBalance());
      }
      if (kasEl) kasEl.textContent = fmtKas1(sompi) + " KAS";
    } catch (_) {
      if (kasEl) kasEl.textContent = "KAS —";
    }
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
    persist(r.id, r.address, "");
    paintButtons();
    status("Logged in");
    loadIdentity(r.address).then(function (id) {
      if (id && id.linked) status(id.linked);
    });
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
      if (current().name) {
        status(current().name);
        return;
      }
      location.href = "/mine?address=" + encodeURIComponent(current().address);
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
      persist("kasware", quiet[0], current().name);
      paintButtons();
      loadIdentity(quiet[0]);
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
      if (e.target.closest("[data-kns-close]")) {
        var km = document.getElementById("knsModal");
        if (km) km.hidden = true;
      }
      if (e.target.closest("[data-kns-clear]")) {
        pickName("");
        return;
      }
      if (e.target.closest("[data-kns-pin]")) {
        var pin = document.getElementById("knsPin");
        pickName(pin && pin.value ? pin.value : "kdao");
        return;
      }
      if (e.target.closest("[data-wallet-change-name]") || e.target.closest("[data-change-name]")) {
        loadIdentity(current().address).then(function (id) {
          openNameModal(id || { names: [] });
        });
        return;
      }
      var copyWho = e.target.closest("[data-copy-who]");
      if (copyWho) {
        e.preventDefault();
        var kind = copyWho.getAttribute("data-copy-who");
        var c = current();
        if (kind === "name" && !c.name) {
          loadIdentity(c.address).then(function (id) {
            openNameModal(id || { names: [] });
          });
          return;
        }
        var text = kind === "name" ? c.name : c.address;
        if (text) copyText(text, copyWho);
        return;
      }
      var kns = e.target.closest("[data-kns-pick]");
      if (kns) {
        pickName(kns.getAttribute("data-kns-pick"));
        return;
      }
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

  async function loadIdentity(address) {
    if (!address) return null;
    try {
      const res = await fetch("/api/id?address=" + encodeURIComponent(address));
      const j = await res.json();
      const id = j && j.id ? j.id : null;
      if (!id) return null;
      var name = id.linked || "";
      if (!name && id.names && id.names.length === 1) name = id.names[0];
      persist(current().id || "kasware", address, name);
      paintButtons();
      window._gramlaneId = id;
      return id;
    } catch (_) {
      return null;
    }
  }

  function copyText(text, btn) {
    text = String(text || "");
    if (!text) return;
    function ok() {
      status("Copied");
      if (!btn) return;
      var old = btn.textContent;
      btn.textContent = "Copied";
      setTimeout(function () {
        if (btn.textContent === "Copied") btn.textContent = old;
      }, 1200);
    }
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).then(ok, function () {});
    } else {
      var ta = document.createElement("textarea");
      ta.value = text;
      document.body.appendChild(ta);
      ta.select();
      try {
        document.execCommand("copy");
        ok();
      } catch (_) {}
      document.body.removeChild(ta);
    }
  }

  function openNameModal(id) {
    var m = document.getElementById("knsModal");
    if (!m) {
      var wrap = document.createElement("div");
      wrap.innerHTML =
        '<div id="knsModal" class="wmodal" hidden><div class="wmodal-card">' +
        '<div class="wmodal-head"><strong>Change name</strong>' +
        '<button type="button" class="btn ghost" data-kns-close>Close</button></div>' +
        '<p class="tiny">Same wallet. Pick which .kas is your face — that is the shop you are using.</p>' +
        '<input id="knsFilter" placeholder="filter" style="width:100%;margin:10px 0" />' +
        '<div id="knsList" class="kns-list"></div>' +
        '<p style="margin-top:12px"><a class="btn ghost" href="/kasdomain">Get a name</a> <button type="button" class="btn ghost" data-kns-clear>Show kaspa address only</button></p>' +
        "</div></div>";
      document.body.appendChild(wrap.firstElementChild);
      m = document.getElementById("knsModal");
    }
    var list = document.getElementById("knsList");
    var names = (id && id.names) || [];
    function draw(filter) {
      filter = (filter || "").toLowerCase();
      list.innerHTML = names
        .filter(function (n) {
          return !filter || n.indexOf(filter) !== -1;
        })
        .map(function (n) {
          var on = current().name === n ? " mint" : " ghost";
          return '<button type="button" class="btn' + on + '" data-kns-pick="' + n + '">' + n + "</button>";
        })
        .join(" ");
    }
    draw("");
    var f = document.getElementById("knsFilter");
    f.value = "";
    f.oninput = function () {
      draw(f.value);
    };
    m.hidden = false;
  }

  function pickName(name) {
    var c = current();
    if (!c.address) return;
    if (name) {
      var body = "act=face&address=" + encodeURIComponent(c.address) + "&name=" + encodeURIComponent(name);
      fetch("/kasdomain", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" }, body: body }).catch(function () {});
    }
    fetch("/api/id", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ address: c.address, name: name || "" }),
    })
      .then(function (r) {
        return r.json();
      })
      .then(function (j) {
        if (!j || !j.ok) {
          status((j && j.error) || "could not pin");
          var sug = (j && j.suggestions) || [];
          var list = document.getElementById("knsList");
          if (list && sug.length) {
            list.innerHTML = sug
              .map(function (n) {
                return '<button type="button" class="btn mint" data-kns-pick="' + n + '">' + n + "</button>";
              })
              .join(" ");
          }
          return;
        }
        var linked = name ? (j && j.id && j.id.linked ? j.id.linked : name) : "";
        persist(c.id, c.address, linked);
        paintButtons();
        var m = document.getElementById("knsModal");
        if (m) m.hidden = true;
        status(linked ? linked + " is your shop" : "Showing Kaspa address");
      })
      .catch(function (err) {
        status(err.message || String(err));
      });
  }

  window.KaspaWallets = { connect: connect, logout: logout, current: current, detected: detected, paintButtons: paintButtons };

  function bindEasy() {
    var main = document.querySelector("main");
    if (main && !main.id) main.id = "main";
    document.addEventListener("click", function (e) {
      var b = e.target.closest("[data-copy], [data-copy-text]");
      if (!b || b.closest("[data-copy-who]")) return;
      var direct = b.getAttribute("data-copy-text");
      var sel = b.getAttribute("data-copy");
      var el = sel ? document.querySelector(sel) : null;
      var text = direct || (el ? el.value || el.textContent : "");
      if (!text) return;
      e.preventDefault();
      copyText(text, b);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function () {
      bind();
      bindEasy();
    });
  } else {
    bind();
    bindEasy();
  }
})();
