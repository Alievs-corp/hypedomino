let seed = 11;
const rnd = () => ((seed = (seed * 16807) % 2147483647) / 2147483647);

function flatten(cmds, steps = 6) {
  const contours = []; let cur = null, last = [0, 0];
  for (const c of cmds) {
    if (c.type === 'M') { if (cur && cur.length > 2) contours.push(cur); cur = [[c.x, c.y]]; last = [c.x, c.y]; }
    else if (c.type === 'L') { cur.push([c.x, c.y]); last = [c.x, c.y]; }
    else if (c.type === 'Q') {
      for (let i = 1; i <= steps; i++) { const t = i / steps, u = 1 - t; cur.push([u * u * last[0] + 2 * u * t * c.x1 + t * t * c.x, u * u * last[1] + 2 * u * t * c.y1 + t * t * c.y]); }
      last = [c.x, c.y];
    } else if (c.type === 'C') {
      for (let i = 1; i <= steps; i++) { const t = i / steps, u = 1 - t; cur.push([u ** 3 * last[0] + 3 * u * u * t * c.x1 + 3 * u * t * t * c.x2 + t ** 3 * c.x, u ** 3 * last[1] + 3 * u * u * t * c.y1 + 3 * u * t * t * c.y2 + t ** 3 * c.y]); }
      last = [c.x, c.y];
    } else if (c.type === 'Z') { if (cur && cur.length > 2) contours.push(cur); cur = null; }
  }
  if (cur && cur.length > 2) contours.push(cur);
  return contours;
}

function resample(pts, spacing) {
  const out = [];
  for (let i = 0; i < pts.length; i++) {
    const a = pts[i], b = pts[(i + 1) % pts.length];
    const d = Math.hypot(b[0] - a[0], b[1] - a[1]), n = Math.max(1, Math.round(d / spacing));
    for (let k = 0; k < n; k++) out.push([a[0] + (b[0] - a[0]) * k / n, a[1] + (b[1] - a[1]) * k / n]);
  }
  return out;
}

function sharpen(contour, o) {
  const pts = resample(contour, o.spacing);
  const res = [];
  for (let i = 0; i < pts.length; i++) {
    const p = pts[i], a = pts[(i - 1 + pts.length) % pts.length], b = pts[(i + 1) % pts.length];
    let nx = -(b[1] - a[1]), ny = b[0] - a[0]; const l = Math.hypot(nx, ny) || 1; nx /= l; ny /= l;
    let j = (rnd() - 0.5) * 2 * o.amp;
    const r = rnd();
    if (r < o.chip) j -= o.notch * (0.6 + rnd() * 0.8);
    else if (r > 1 - o.spike) j += o.notch * (0.6 + rnd() * 0.8);
    res.push([p[0] + nx * j, p[1] + ny * j]);
  }
  return res;
}

function toD(contours) {
  return contours.map(c => 'M' + c.map(p => p[0].toFixed(1) + ' ' + p[1].toFixed(1)).join(' L') + ' Z').join(' ');
}

module.exports = { flatten, sharpen, toD, rnd };
