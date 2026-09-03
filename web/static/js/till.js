(function () {
  var out = document.querySelector("[data-preview]");
  if (!out) return;
  function num(sel) {
    var el = document.querySelector(sel);
    if (!el) return 0;
    return Number(String(el.value || "").replace(",", "."));
  }
  function preview() {
    var grams = num("[name=grams]");
    var kas = num("[name=kas]");
    var fiat = num("[name=fiat]");
    var rate = num("[name=rate]");
    var n = 0;
    if (grams > 0) n = Math.floor(grams);
    else if (kas > 0) n = Math.floor(kas * 1000000);
    else if (fiat > 0 && rate > 0) n = Math.floor((fiat / rate) * 1000000);
    if (n > 0) out.textContent = n.toLocaleString("en-US") + " GRAM";
    else out.textContent = "grams appear here";
  }
  document.querySelectorAll("[name=fiat],[name=rate],[name=grams],[name=kas]").forEach(function (el) {
    el.addEventListener("input", preview);
  });
  preview();
})();
