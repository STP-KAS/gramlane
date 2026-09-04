(function () {
  var root = document.querySelector("[data-telegram]");
  if (!root) return;
  var KEY = "gramlaneTelegramSecret";
  var room = root.getAttribute("data-room") || "board";
  var fromEl = document.querySelector("[data-wallet-addr]");
  var secretEl = document.getElementById("tgSecret");
  var form = document.getElementById("tgForm");
  var textEl = document.getElementById("tgText");

  function secret() {
    return (secretEl && secretEl.value ? secretEl.value : "").trim();
  }

  function persistSecret() {
    try {
      if (secret()) sessionStorage.setItem(KEY, secret());
    } catch (_) {}
  }

  function loadSecret() {
    try {
      var s = sessionStorage.getItem(KEY);
      if (s && secretEl && !secretEl.value) secretEl.value = s;
    } catch (_) {}
  }

  function hex(buf) {
    var u = new Uint8Array(buf);
    var o = "";
    for (var i = 0; i < u.length; i++) o += u[i].toString(16).padStart(2, "0");
    return o;
  }

  function unhex(s) {
    s = String(s || "");
    var u = new Uint8Array(s.length / 2);
    for (var i = 0; i < u.length; i++) u[i] = parseInt(s.substr(i * 2, 2), 16);
    return u;
  }

  async function keyOf(pass, roomName) {
    var enc = new TextEncoder();
    var base = await crypto.subtle.importKey("raw", enc.encode(pass), "PBKDF2", false, ["deriveKey"]);
    return crypto.subtle.deriveKey(
      { name: "PBKDF2", salt: enc.encode("gramlane-telegram:" + roomName), iterations: 120000, hash: "SHA-256" },
      base,
      { name: "AES-GCM", length: 256 },
      false,
      ["encrypt", "decrypt"]
    );
  }

  async function seal(pass, roomName, plain) {
    var k = await keyOf(pass, roomName);
    var nonce = crypto.getRandomValues(new Uint8Array(12));
    var ct = await crypto.subtle.encrypt({ name: "AES-GCM", iv: nonce }, k, new TextEncoder().encode(plain));
    return { nonce: hex(nonce), box: hex(ct) };
  }

  async function openBox(pass, roomName, nonce, box) {
    var k = await keyOf(pass, roomName);
    var pt = await crypto.subtle.decrypt({ name: "AES-GCM", iv: unhex(nonce) }, k, unhex(box));
    return new TextDecoder().decode(pt);
  }

  async function paint() {
    var pass = secret();
    document.querySelectorAll(".kc-bubble[data-box]").forEach(async function (el) {
      var line = el.querySelector(".tg-plain");
      if (!line) return;
      if (!pass) {
        line.textContent = "locked — type the shared secret";
        return;
      }
      try {
        line.textContent = await openBox(pass, room, el.getAttribute("data-nonce"), el.getAttribute("data-box"));
      } catch (_) {
        line.textContent = "wrong secret";
      }
    });
  }

  loadSecret();
  paint();
  if (secretEl) {
    secretEl.addEventListener("input", function () {
      persistSecret();
      paint();
    });
  }

  if (form) {
    form.addEventListener("submit", async function (e) {
      e.preventDefault();
      var pass = secret();
      var text = (textEl && textEl.value ? textEl.value : "").trim();
      if (!pass) {
        alert("Set a shared secret first. Both sides type the same one.");
        return;
      }
      if (!text) return;
      var from = (fromEl && (fromEl.value || fromEl.textContent) ? fromEl.value || fromEl.textContent : "").trim();
      try {
        var sealed = await seal(pass, room, text);
        var res = await fetch("/api/telegram", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ room: room, from: from, nonce: sealed.nonce, box: sealed.box }),
        });
        var j = await res.json();
        if (!j || !j.ok) {
          alert((j && j.error) || "could not send");
          return;
        }
        location.reload();
      } catch (err) {
        alert(err && err.message ? err.message : String(err));
      }
    });
  }
})();
