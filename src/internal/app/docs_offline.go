package app

import (
	"encoding/json"
	"fmt"
	"html"
	"io/fs"

	"gopkg.in/yaml.v3"
)

// offlineDocsHTML returns a self-contained HTML page that renders the OpenAPI spec
// with no external scripts or CDN, so /docs works without internet.
func offlineDocsHTML(specRoot fs.FS) ([]byte, error) {
	data, err := fs.ReadFile(specRoot, "openapi.yaml")
	if err != nil {
		return nil, err
	}
	var specMap map[string]interface{}
	if err := yaml.Unmarshal(data, &specMap); err != nil {
		return nil, err
	}
	specJSON, err := json.Marshal(specMap)
	if err != nil {
		return nil, err
	}
	// Safe for embedding in HTML: escape so </script> in spec doesn't break the page
	specEscaped := html.EscapeString(string(specJSON))
	const tpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1"/>
  <title>ETRUSCAN API Reference</title>
  <style>
    * { box-sizing: border-box; }
    body { font-family: system-ui, sans-serif; margin: 0; background: #0f0f14; color: #e4e4e7; line-height: 1.5; }
    .wrap { max-width: 960px; margin: 0 auto; padding: 1.5rem; }
    h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
    .tag { margin-top: 2rem; padding-top: 1rem; border-top: 1px solid #333; }
    .tag h2 { font-size: 1.2rem; color: #a78bfa; margin-bottom: 0.75rem; }
    .op { margin: 0.75rem 0; padding: 0.75rem; background: #1a1a24; border-radius: 6px; border-left: 3px solid #6366f1; }
    .op .method { font-weight: 600; color: #818cf8; }
    .op .path { font-family: ui-monospace, monospace; }
    .op .summary { color: #94a3b8; font-size: 0.9rem; margin-top: 0.25rem; }
    .op details { margin-top: 0.5rem; }
    .op summary { cursor: pointer; color: #a5b4fc; font-size: 0.9rem; }
    pre { margin: 0.5rem 0 0; padding: 0.75rem; background: #12121a; border-radius: 4px; overflow: auto; font-size: 0.8rem; }
    .note { font-size: 0.85rem; color: #64748b; margin-top: 1rem; }
  </style>
</head>
<body>
  <div class="wrap">
    <h1>ETRUSCAN Platform API</h1>
    <p class="note">Offline API reference — no external resources.</p>
    <div id="app"></div>
  </div>
  <script type="application/json" id="spec">%s</script>
  <script>
(function() {
  var raw = document.getElementById('spec').textContent;
  var spec;
  try { spec = JSON.parse(raw); } catch (e) { document.getElementById('app').innerHTML = '<p>Invalid spec</p>'; return; }
  var paths = spec.paths || {};
  var pathKeys = Object.keys(paths).sort();
  var byTag = {};
  pathKeys.forEach(function(path) {
    var ops = paths[path];
    Object.keys(ops).forEach(function(method) {
      if (method === 'parameters' || method === 'summary') return;
      var op = ops[method];
      var tags = (op.tags || ['Default']);
      tags.forEach(function(tag) {
        if (!byTag[tag]) byTag[tag] = [];
        byTag[tag].push({ method: method.toUpperCase(), path: path, op: op });
      });
    });
  });
  var tagOrder = Object.keys(byTag).sort();
  var html = '';
  tagOrder.forEach(function(tag) {
    html += '<div class="tag"><h2>' + escapeHtml(tag) + '</h2>';
    byTag[tag].forEach(function(item) {
      var s = (item.op.summary || item.method + ' ' + item.path);
      html += '<div class="op">';
      html += '<span class="method">' + escapeHtml(item.method) + '</span> ';
      html += '<span class="path">' + escapeHtml(item.path) + '</span>';
      html += '<div class="summary">' + escapeHtml(s) + '</div>';
      html += '<details><summary>Details</summary><pre>' + escapeHtml(JSON.stringify({ requestBody: item.op.requestBody, responses: item.op.responses }, null, 2)) + '</pre></details>';
      html += '</div>';
    });
    html += '</div>';
  });
  document.getElementById('app').innerHTML = html || '<p>No operations.</p>';
  function escapeHtml(s) {
    if (typeof s !== 'string') return '';
    var d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }
})();
  </script>
</body>
</html>`
	return []byte(fmt.Sprintf(tpl, specEscaped)), nil
}
