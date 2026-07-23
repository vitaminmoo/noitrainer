package main

// Self-contained explorer page: the rendered world as a pannable/zoomable
// canvas, with a hierarchy of chunks, rigid bodies and joints beside it.
// The PNG is inlined as a data URI and the index as JSON, so the file works
// offline with no sibling assets.

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func writeExplorer(path, pngPath string, index jsonWorld) error {
	pngBytes, err := os.ReadFile(pngPath)
	if err != nil {
		return err
	}
	indexJSON, err := json.Marshal(index)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString(explorerHead)
	fmt.Fprintf(&b, "<script>\nconst WORLD = %s;\nconst IMG = \"data:image/png;base64,%s\";\n</script>\n",
		indexJSON, base64.StdEncoding.EncodeToString(pngBytes))
	b.WriteString(explorerBody)

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

const explorerHead = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Noita save explorer</title>
<style>
  :root {
    --bg: #0e0f13; --panel: #16181f; --line: #262a35;
    --text: #e6e8ee; --dim: #9aa1b1; --accent: #7cc7ff; --warn: #ffcc66;
  }
  * { box-sizing: border-box; }
  html, body { margin: 0; height: 100%; }
  body {
    background: var(--bg); color: var(--text); display: flex; height: 100vh;
    font: 13px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  }
  #side {
    width: 340px; flex: none; background: var(--panel);
    border-right: 1px solid var(--line); display: flex; flex-direction: column;
  }
  #side h1 { font-size: 13px; margin: 0; padding: 12px 14px; border-bottom: 1px solid var(--line); }
  #side h1 small { color: var(--dim); font-weight: normal; display: block; margin-top: 2px; }
  #tabs { display: flex; border-bottom: 1px solid var(--line); }
  #tabs button {
    flex: 1; background: none; border: 0; border-bottom: 2px solid transparent;
    color: var(--dim); padding: 8px 4px; cursor: pointer; font: inherit;
  }
  #tabs button.on { color: var(--accent); border-bottom-color: var(--accent); }
  #list { overflow-y: auto; flex: 1; }
  .row {
    padding: 6px 14px; border-bottom: 1px solid var(--line);
    cursor: pointer; display: flex; gap: 8px; align-items: baseline;
  }
  .row:hover { background: #1d2029; }
  .row.sel { background: #23304a; }
  .row .name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .row .n { color: var(--dim); font-size: 11px; }
  .kids { display: none; }
  .kids.open { display: block; }
  .kid { padding-left: 30px; font-size: 12px; color: var(--dim); }
  .kid:hover { color: var(--text); }
  .sw { width: 10px; height: 10px; border-radius: 2px; flex: none; border: 1px solid #0006; }
  #main { flex: 1; position: relative; overflow: hidden; }
  canvas { display: block; cursor: grab; }
  canvas.drag { cursor: grabbing; }
  #hud, #detail {
    position: absolute; background: #0e0f13e8; border: 1px solid var(--line);
    border-radius: 6px; padding: 8px 10px; pointer-events: none;
  }
  #hud { top: 10px; left: 10px; color: var(--dim); }
  #hud b { color: var(--text); font-weight: normal; }
  #detail { right: 10px; top: 10px; max-width: 300px; display: none; }
  #detail div { color: var(--dim); }
  #detail div span { color: var(--text); }
  #toolbar {
    position: absolute; bottom: 10px; left: 10px; display: flex; gap: 6px; flex-wrap: wrap;
  }
  #toolbar label {
    background: #0e0f13e8; border: 1px solid var(--line); border-radius: 6px;
    padding: 5px 8px; cursor: pointer; user-select: none;
  }
  #toolbar input { vertical-align: -1px; margin-right: 4px; }
  .warn { color: var(--warn); }
</style>
</head>
<body>
`

const explorerBody = `
<div id="side">
  <h1>Noita save explorer<small id="sub"></small></h1>
  <div id="tabs">
    <button data-tab="chunks" class="on">Chunks</button>
    <button data-tab="bodies">Bodies</button>
    <button data-tab="materials">Materials</button>
  </div>
  <div id="list"></div>
</div>
<div id="main">
  <canvas id="cv"></canvas>
  <div id="hud"></div>
  <div id="detail"></div>
  <div id="toolbar">
    <label><input type="checkbox" id="tGrid" checked>chunk grid</label>
    <label><input type="checkbox" id="tBodies" checked>body markers</label>
    <label><input type="checkbox" id="tLabels">labels</label>
    <label><input type="checkbox" id="tSmooth">smooth</label>
  </div>
</div>
<script>
const cv = document.getElementById('cv'), ctx = cv.getContext('2d');
const B = WORLD.bounds;
const img = new Image();
let view = { x: B.x, y: B.y, scale: 1 }; // world coords at canvas origin
let sel = null, hoverBody = null;

// Flatten bodies once, tagged with their owning chunk.
const bodies = [];
for (const c of WORLD.chunks) for (const o of (c.objects || [])) bodies.push({...o, chunk: c});
const jointCount = WORLD.totals.joints;

document.getElementById('sub').textContent =
  WORLD.totals.chunks + ' chunks · ' + WORLD.totals.objects + ' bodies · ' +
  jointCount + ' joints · ' + WORLD.totals.solidCells.toLocaleString() + ' solid cells';

function resize() {
  const m = document.getElementById('main');
  cv.width = m.clientWidth;
  cv.height = m.clientHeight;
  draw();
}
window.addEventListener('resize', resize);

function fit() {
  const s = Math.min(cv.width / B.w, cv.height / B.h) * 0.95;
  view.scale = s;
  view.x = B.x + B.w / 2 - cv.width / (2 * s);
  view.y = B.y + B.h / 2 - cv.height / (2 * s);
}
const w2s = (wx, wy) => [(wx - view.x) * view.scale, (wy - view.y) * view.scale];
const s2w = (sx, sy) => [sx / view.scale + view.x, sy / view.scale + view.y];

function draw() {
  ctx.setTransform(1, 0, 0, 1, 0, 0);
  ctx.fillStyle = '#07080b';
  ctx.fillRect(0, 0, cv.width, cv.height);

  ctx.imageSmoothingEnabled = document.getElementById('tSmooth').checked;
  const [ox, oy] = w2s(B.x, B.y);
  ctx.drawImage(img, ox, oy, B.w * view.scale, B.h * view.scale);

  if (document.getElementById('tGrid').checked) {
    ctx.lineWidth = 1;
    for (const c of WORLD.chunks) {
      const [x, y] = w2s(c.x, c.y);
      const s = WORLD.chunkSize * view.scale;
      ctx.strokeStyle = (sel && sel.kind === 'chunk' && sel.item === c) ? '#7cc7ff' : '#ffffff22';
      ctx.strokeRect(x, y, s, s);
      if (document.getElementById('tLabels').checked && s > 90) {
        ctx.fillStyle = '#ffffff66';
        ctx.fillText(c.x + ',' + c.y, x + 4, y + 12);
      }
    }
  }

  if (document.getElementById('tBodies').checked) {
    for (const b of bodies) {
      const [x, y] = w2s(b.x, b.y);
      const on = (sel && sel.kind === 'body' && sel.item === b) || hoverBody === b;
      const r = Math.max(3, Math.min(b.w, b.h) * view.scale / 2);
      ctx.beginPath();
      ctx.arc(x, y, r, 0, Math.PI * 2);
      ctx.strokeStyle = on ? '#7cc7ff' : (b.isStatic ? '#ffcc6699' : '#ff7c7c99');
      ctx.lineWidth = on ? 2 : 1;
      ctx.stroke();
      if (on) {
        ctx.beginPath(); ctx.moveTo(x - 12, y); ctx.lineTo(x + 12, y);
        ctx.moveTo(x, y - 12); ctx.lineTo(x, y + 12); ctx.stroke();
      }
    }
  }
}

// ---- interaction ----------------------------------------------------------
let drag = null;
cv.addEventListener('mousedown', e => {
  drag = { sx: e.offsetX, sy: e.offsetY, vx: view.x, vy: view.y };
  cv.classList.add('drag');
});
window.addEventListener('mouseup', () => { drag = null; cv.classList.remove('drag'); });
cv.addEventListener('mousemove', e => {
  if (drag) {
    view.x = drag.vx - (e.offsetX - drag.sx) / view.scale;
    view.y = drag.vy - (e.offsetY - drag.sy) / view.scale;
  }
  const [wx, wy] = s2w(e.offsetX, e.offsetY);
  const cs = WORLD.chunkSize;
  const cx = Math.floor(wx / cs) * cs, cy = Math.floor(wy / cs) * cs;
  document.getElementById('hud').innerHTML =
    'world <b>' + Math.floor(wx) + ', ' + Math.floor(wy) + '</b><br>' +
    'chunk <b>' + cx + ', ' + cy + '</b><br>zoom <b>' + view.scale.toFixed(2) + '×</b>';

  // Nearest body within a few screen pixels.
  let best = null, bestD = 14 / view.scale;
  for (const b of bodies) {
    const d = Math.hypot(b.x - wx, b.y - wy);
    if (d < bestD) { bestD = d; best = b; }
  }
  if (best !== hoverBody) { hoverBody = best; }
  draw();
});
cv.addEventListener('wheel', e => {
  e.preventDefault();
  const [wx, wy] = s2w(e.offsetX, e.offsetY);
  const k = Math.exp(-e.deltaY * 0.0015);
  view.scale = Math.max(0.02, Math.min(40, view.scale * k));
  // Keep the cursor anchored to the same world point.
  view.x = wx - e.offsetX / view.scale;
  view.y = wy - e.offsetY / view.scale;
  draw();
}, { passive: false });
cv.addEventListener('click', () => {
  if (hoverBody) select('body', hoverBody, false);
});
for (const id of ['tGrid', 'tBodies', 'tLabels', 'tSmooth'])
  document.getElementById(id).addEventListener('change', draw);

function focusOn(wx, wy, scale) {
  if (scale) view.scale = scale;
  view.x = wx - cv.width / (2 * view.scale);
  view.y = wy - cv.height / (2 * view.scale);
  draw();
}

function select(kind, item, move = true) {
  sel = { kind, item };
  const d = document.getElementById('detail');
  d.style.display = 'block';
  if (kind === 'chunk') {
    d.innerHTML = '<b>' + item.file + '</b>' +
      row('origin', item.x + ', ' + item.y) +
      row('version', item.version) +
      row('solid cells', item.solidCells.toLocaleString()) +
      row('materials', item.materials.length) +
      row('custom colors', item.customColors.toLocaleString()) +
      row('bodies', (item.objects||[]).length) +
      row('joints', (item.joints||[]).length);
    if (move) focusOn(item.x + WORLD.chunkSize / 2, item.y + WORLD.chunkSize / 2,
                      Math.max(view.scale, 0.35));
  } else if (kind === 'body') {
    const js = (item.chunk.joints||[]).filter(j => j.bodyA === item.id || j.bodyB === item.id);
    d.innerHTML = '<b>body ' + item.id + '</b>' +
      row('material', item.material) +
      row('position', item.x.toFixed(1) + ', ' + item.y.toFixed(1)) +
      row('rotation', (item.rot * 180 / Math.PI).toFixed(1) + '°') +
      row('image', item.w + '×' + item.h) +
      row('shape', item.isCircle ? 'circle r=' + item.radius.toFixed(2) : 'pixel body') +
      row('static', item.isStatic) +
      row('z', item.z) +
      row('joints', js.length ? js.map(j => j.kind).join(', ') : 'none') +
      row('chunk', item.chunk.file);
    if (move) focusOn(item.x, item.y, Math.max(view.scale, 2));
  } else if (kind === 'material') {
    d.innerHTML = '<b>' + item.name + '</b>' +
      row('cells', item.cells.toLocaleString()) +
      row('color', item.color) +
      (item.known ? '' : '<div class="warn">not found in materials.xml</div>');
  }
  draw();
}
const row = (k, v) => '<div>' + k + ' <span>' + v + '</span></div>';

// ---- sidebar --------------------------------------------------------------
let tab = 'chunks';
for (const b of document.querySelectorAll('#tabs button')) {
  b.addEventListener('click', () => {
    tab = b.dataset.tab;
    for (const o of document.querySelectorAll('#tabs button')) o.classList.toggle('on', o === b);
    renderList();
  });
}

function renderList() {
  const el = document.getElementById('list');
  el.innerHTML = '';
  if (tab === 'chunks') {
    for (const c of WORLD.chunks) {
      const r = mkRow(c.file, (c.objects||[]).length + 'b/' + (c.joints||[]).length + 'j');
      const kids = document.createElement('div');
      kids.className = 'kids';
      for (const o of (c.objects || [])) {
        const k = mkRow('body ' + o.id.slice(-6) + ' · ' + (o.material || '?'), o.w + '×' + o.h);
        k.classList.add('kid');
        k.onclick = ev => { ev.stopPropagation(); select('body', bodies.find(b => b.id === o.id && b.chunk === c)); };
        kids.appendChild(k);
      }
      for (const j of (c.joints || [])) {
        const k = mkRow('joint ' + j.kind, j.local ? '' : 'cross-chunk');
        k.classList.add('kid');
        if (!j.local) k.classList.add('warn');
        kids.appendChild(k);
      }
      r.onclick = () => {
        kids.classList.toggle('open');
        select('chunk', c);
        markSel(r);
      };
      el.appendChild(r); el.appendChild(kids);
    }
  } else if (tab === 'bodies') {
    const sorted = [...bodies].sort((a, b) => b.pixels - a.pixels);
    for (const o of sorted) {
      const r = mkRow((o.material || '?') + ' · ' + o.id.slice(-6), o.w + '×' + o.h);
      r.onclick = () => { select('body', o); markSel(r); };
      el.appendChild(r);
    }
  } else {
    for (const m of WORLD.materials) {
      const r = mkRow(m.name, m.cells.toLocaleString());
      const sw = document.createElement('span');
      sw.className = 'sw'; sw.style.background = m.color;
      r.prepend(sw);
      if (!m.known) r.classList.add('warn');
      r.onclick = () => { select('material', m); markSel(r); };
      el.appendChild(r);
    }
  }
}
function mkRow(name, n) {
  const r = document.createElement('div');
  r.className = 'row';
  r.innerHTML = '<span class="name"></span><span class="n"></span>';
  r.querySelector('.name').textContent = name;
  r.querySelector('.n').textContent = n;
  return r;
}
function markSel(r) {
  for (const o of document.querySelectorAll('.row.sel')) o.classList.remove('sel');
  r.classList.add('sel');
}

img.onload = () => { resize(); fit(); draw(); };
img.src = IMG;
renderList();
</script>
</body>
</html>
`
