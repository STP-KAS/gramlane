/* Gramlane phone shell. Network first so the desk stays in sync. */
self.addEventListener("install", function (e) {
  self.skipWaiting();
});
self.addEventListener("activate", function (e) {
  e.waitUntil(self.clients.claim());
});
self.addEventListener("fetch", function (e) {
  var u = new URL(e.request.url);
  if (e.request.method !== "GET") return;
  if (u.origin !== self.location.origin) return;
  if (u.pathname.indexOf("/api/") === 0) return;
  if (u.pathname.indexOf("/static/") === 0) {
    e.respondWith(
      caches.open("gramlane-static-v1").then(function (c) {
        return c.match(e.request).then(function (hit) {
          var live = fetch(e.request).then(function (res) {
            if (res && res.ok) c.put(e.request, res.clone());
            return res;
          });
          return hit || live;
        });
      })
    );
    return;
  }
  e.respondWith(
    fetch(e.request).catch(function () {
      return caches.match(e.request);
    })
  );
});
