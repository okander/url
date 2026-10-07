package server

// indexHTML is the landing page served at "/". It is a plain string constant
// (no embed, no static-file config) so it works identically locally and on
// Vercel. The page only calls the same public API documented in the README.
const indexHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Shortener</title>
<style>
:root{--bg:#fff;--fg:#1a1a1a;--muted:#666;--card:#f5f5f7;--accent:#2563eb;--border:#d4d4d8}
@media (prefers-color-scheme:dark){:root{--bg:#111113;--fg:#f4f4f5;--muted:#a1a1aa;--card:#1c1c1f;--accent:#60a5fa;--border:#3f3f46}}
*{box-sizing:border-box}
body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;background:var(--bg);color:var(--fg);font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;padding:1rem}
main{width:100%;max-width:32rem;background:var(--card);border:1px solid var(--border);border-radius:12px;padding:1.5rem}
h1{margin:0 0 .25rem;font-size:1.5rem}
p.sub{margin:0 0 1.25rem;color:var(--muted);font-size:.95rem}
label{display:block;margin:.75rem 0 .25rem;font-size:.85rem;color:var(--muted)}
input{width:100%;padding:.65rem .75rem;border:1px solid var(--border);border-radius:8px;background:var(--bg);color:var(--fg);font-size:1rem}
button{margin-top:1rem;padding:.65rem 1rem;border:0;border-radius:8px;background:var(--accent);color:#fff;font-size:1rem;cursor:pointer}
button:disabled{opacity:.6;cursor:default}
#result{margin-top:1rem;word-break:break-all}
#result a{color:var(--accent)}
.err{color:#dc2626}
.hint{margin-top:1rem;font-size:.8rem;color:var(--muted)}
</style>
</head>
<body>
<main>
<h1>Shortener</h1>
<p class="sub">Paste a long link, get a short one.</p>
<label for="url">Long URL</label>
<input id="url" type="url" placeholder="https://example.com/a/very/long/path" autocomplete="off">
<label for="key">API key</label>
<input id="key" type="password" placeholder="Required to create links" autocomplete="off">
<button id="go">Shorten</button>
<div id="result" aria-live="polite"></div>
<p class="hint">Short links work for everyone. An API key is only needed to create them.</p>
</main>
<script>
(function () {
  var urlEl = document.getElementById("url");
  var keyEl = document.getElementById("key");
  var btn = document.getElementById("go");
  var out = document.getElementById("result");

  function show(node) { out.textContent = ""; out.appendChild(node); }
  function message(text, cls) {
    var p = document.createElement("p");
    p.textContent = text;
    if (cls) p.className = cls;
    show(p);
  }

  async function shorten() {
    var url = urlEl.value.trim();
    if (!url) { message("Enter a URL first.", "err"); return; }
    btn.disabled = true;
    try {
      var headers = { "Content-Type": "application/json" };
      var key = keyEl.value.trim();
      if (key) headers["Authorization"] = "Bearer " + key;
      var res = await fetch("/links", { method: "POST", headers: headers, body: JSON.stringify({ url: url }) });
      var data = await res.json().catch(function () { return {}; });
      if (!res.ok) { message(data.error || ("Request failed (" + res.status + ")"), "err"); return; }

      var wrap = document.createElement("div");
      var a = document.createElement("a");
      a.href = data.short_url;
      a.textContent = data.short_url;
      a.target = "_blank";
      a.rel = "noopener";
      var copy = document.createElement("button");
      copy.textContent = "Copy";
      copy.onclick = function () {
        navigator.clipboard.writeText(data.short_url).then(function () { copy.textContent = "Copied"; });
      };
      wrap.appendChild(a);
      wrap.appendChild(document.createElement("br"));
      wrap.appendChild(copy);
      show(wrap);
    } catch (e) {
      message("Network error. Try again.", "err");
    } finally {
      btn.disabled = false;
    }
  }

  btn.addEventListener("click", shorten);
  urlEl.addEventListener("keydown", function (e) { if (e.key === "Enter") shorten(); });
})();
</script>
</body>
</html>
`
