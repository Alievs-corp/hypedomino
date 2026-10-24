const fs = require('fs');
const opentype = require(process.env.OT || 'opentype.js');
const fontFile = process.argv[2] || 'fonts/Kanit-BlackItalic.ttf';
const out = process.argv[3] || 'hypedomino-logo';
const font = opentype.parse(fs.readFileSync(fontFile).buffer.slice(fs.readFileSync(fontFile).byteOffset, fs.readFileSync(fontFile).byteOffset + fs.readFileSync(fontFile).byteLength));

let seed = 7;
const rnd = () => ((seed = (seed * 16807) % 2147483647) / 2147483647);

const dx = 46, dy = 46;
const P = {
  1: [[0, 0]], 2: [[-dx, -dy], [dx, dy]], 3: [[-dx, -dy], [0, 0], [dx, dy]],
  4: [[-dx, -dy], [dx, -dy], [-dx, dy], [dx, dy]], 5: [[-dx, -dy], [dx, -dy], [0, 0], [-dx, dy], [dx, dy]],
  6: [[-dx, -dy], [dx, -dy], [-dx, 0], [dx, 0], [-dx, dy], [dx, dy]],
};
const W = 200, H = 350, R = 40;
const pips = (n, cy) => P[n].map(([x, y]) => `<circle cx="${x}" cy="${cy + y}" r="19" fill="url(#pip)"/><circle cx="${x - 6}" cy="${cy + y - 7}" r="5" fill="#fff" opacity=".45"/>`).join('');
function tile(top, bottom, tx, ty, rot) {
  return `<g transform="translate(${tx} ${ty}) rotate(${rot}) rotate(-90)">
  <rect x="${-W / 2 + 14}" y="${-H / 2 + 26}" width="${W}" height="${H}" rx="${R}" fill="url(#ext)"/>
  <rect x="${-W / 2}" y="${-H / 2}" width="${W}" height="${H}" rx="${R}" fill="url(#face)" stroke="url(#rim)" stroke-width="9"/>
  <rect x="${-W / 2 + 10}" y="${-H / 2 + 10}" width="${W - 20}" height="${H - 20}" rx="${R - 10}" fill="none" stroke="#fff" stroke-opacity=".7" stroke-width="3"/>
  <rect x="${-W / 2 + 34}" y="-4" width="${W - 68}" height="8" rx="4" fill="#8a7bd8" opacity=".75"/>
  <circle cx="0" cy="0" r="9" fill="url(#rim)"/>
  ${pips(top, -88)}${pips(bottom, 88)}
</g>`;
}
const streak = (d, g) => `<path d="${d}" fill="url(#${g})"/>`;

function boltPoints(x1, y1, x2, y2, segs, jitter) {
  const len = Math.hypot(x2 - x1, y2 - y1), nx = -(y2 - y1) / len, ny = (x2 - x1) / len;
  const pts = [[x1, y1]];
  for (let i = 1; i < segs; i++) {
    const t = i / segs, j = (rnd() - 0.5) * 2 * jitter;
    pts.push([x1 + (x2 - x1) * t + nx * j, y1 + (y2 - y1) * t + ny * j]);
  }
  pts.push([x2, y2]);
  return pts;
}
const poly = pts => 'M' + pts.map(p => p[0].toFixed(1) + ' ' + p[1].toFixed(1)).join(' L');
function bolt(x1, y1, x2, y2, segs = 9, jitter = 22, branches = 2) {
  const pts = boltPoints(x1, y1, x2, y2, segs, jitter);
  let d = poly(pts);
  for (let b = 0; b < branches; b++) {
    const i = 2 + Math.floor(rnd() * (segs - 4)), [bx, by] = pts[i];
    const ang = Math.atan2(y2 - y1, x2 - x1) + (rnd() - 0.5) * 1.8, l = 40 + rnd() * 60;
    d += ' ' + poly(boltPoints(bx, by, bx + Math.cos(ang) * l, by + Math.sin(ang) * l, 4, 10));
  }
  return `<g fill="none" stroke-linejoin="round" stroke-linecap="round">
  <path d="${d}" stroke="#18c8ff" stroke-width="16" opacity=".55" filter="url(#glow2)"/>
  <path d="${d}" stroke="#2ee6ff" stroke-width="6"/>
  <path d="${d}" stroke="#ffffff" stroke-width="2.4"/>
</g>`;
}
const spark = (x, y, s) => `<path d="M${x} ${y - s} L${x + s * .22} ${y - s * .22} L${x + s} ${y} L${x + s * .22} ${y + s * .22} L${x} ${y + s} L${x - s * .22} ${y + s * .22} L${x - s} ${y} L${x - s * .22} ${y - s * .22} Z" fill="#fff" filter="url(#glow2)"/><path d="M${x} ${y - s} L${x + s * .22} ${y - s * .22} L${x + s} ${y} L${x + s * .22} ${y + s * .22} L${x} ${y + s} L${x - s * .22} ${y + s * .22} L${x - s} ${y} L${x - s * .22} ${y - s * .22} Z" fill="#9ff3ff"/>`;

const SIZE = 205, WORD = 'HypeDomino';
const adv = font.getAdvanceWidth(WORD, SIZE);
const wordPath = font.getPath(WORD, -adv / 2, 0, SIZE).toPathData(2);
const half = adv / 2;
const cuts = [-62, -34].map(y => `<path d="M${-half - 20} ${y + 3} L${half + 20} ${y - 11} L${half + 20} ${y - 3} L${-half - 20} ${y + 11} Z" fill="#000"/>`).join('');

const bolts = [
  bolt(820, 322, 925, 280, 6, 16, 1),
  bolt(560, 255, 745, 150, 9, 24, 2),
  bolt(1170, 400, 1350, 310, 9, 22, 2),
  bolt(110, 640, 430, 585, 11, 24, 2),
  bolt(1180, 585, 1490, 540, 11, 24, 2),
  bolt(740, 545, 830, 610, 5, 14, 1),
].join('\n');
const sparks = [[560, 255, 22], [1350, 310, 18], [110, 640, 20], [1490, 540, 22], [820, 322, 16], [925, 280, 14], [1190, 150, 14], [330, 520, 12]].map(s => spark(...s)).join('');

const body = (bg) => `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1600 900" width="1600" height="900">
<defs>
<linearGradient id="face" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="#cfdcff"/></linearGradient>
<linearGradient id="rim" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#2ee6ff"/><stop offset=".5" stop-color="#ff3ddc"/><stop offset="1" stop-color="#5a3bff"/></linearGradient>
<linearGradient id="ext" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#8a2cff"/><stop offset="1" stop-color="#2a1a9a"/></linearGradient>
<radialGradient id="pip" cx=".35" cy=".3" r=".9"><stop offset="0" stop-color="#4a3bb0"/><stop offset="1" stop-color="#0d0830"/></radialGradient>
<linearGradient id="txt" gradientUnits="userSpaceOnUse" x1="${-half}" y1="-150" x2="${half}" y2="30"><stop offset="0" stop-color="#ffffff"/><stop offset=".34" stop-color="#bfeaff"/><stop offset=".44" stop-color="#ff4fe0"/><stop offset="1" stop-color="#7a3bff"/></linearGradient>
<linearGradient id="outer" gradientUnits="userSpaceOnUse" x1="${-half}" y1="0" x2="${half}" y2="0"><stop offset="0" stop-color="#18c8ff"/><stop offset=".45" stop-color="#3b6bff"/><stop offset=".6" stop-color="#ff2fd0"/><stop offset="1" stop-color="#7a3bff"/></linearGradient>
<linearGradient id="ext2" gradientUnits="userSpaceOnUse" x1="${-half}" y1="0" x2="${half}" y2="0"><stop offset="0" stop-color="#14309a"/><stop offset=".5" stop-color="#8a1fa0"/><stop offset="1" stop-color="#3b1a9a"/></linearGradient>
<linearGradient id="s1" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#ff2fd0"/><stop offset="1" stop-color="#ff2fd0" stop-opacity="0"/></linearGradient>
<linearGradient id="s2" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#2ee6ff"/><stop offset="1" stop-color="#2ee6ff" stop-opacity="0"/></linearGradient>
<linearGradient id="s3" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#9a3bff"/><stop offset="1" stop-color="#9a3bff" stop-opacity="0"/></linearGradient>
<radialGradient id="halo" cx=".5" cy=".5" r=".5"><stop offset="0" stop-color="#2a6bff" stop-opacity=".5"/><stop offset=".6" stop-color="#8a2cff" stop-opacity=".25"/><stop offset="1" stop-color="#8a2cff" stop-opacity="0"/></radialGradient>
<filter id="glow" x="-20%" y="-60%" width="140%" height="220%"><feGaussianBlur stdDeviation="16"/></filter>
<filter id="glow2" x="-50%" y="-50%" width="200%" height="200%"><feGaussianBlur stdDeviation="6"/></filter>
<mask id="cutmask" maskUnits="userSpaceOnUse" x="${-half - 60}" y="-260" width="${adv + 120}" height="420"><rect x="${-half - 60}" y="-260" width="${adv + 120}" height="420" fill="#fff"/>${cuts}</mask>
</defs>
${bg}
<ellipse cx="800" cy="360" rx="620" ry="330" fill="url(#halo)"/>
${streak('M60 560 L500 470 L500 500 Z', 's1')}
${streak('M130 470 L520 380 L520 404 Z', 's2')}
${streak('M20 640 L470 540 L470 560 Z', 's3')}
${streak('M200 395 L560 300 L560 322 Z', 's2')}
${tile(5, 5, 1050, 225, -22)}
${tile(6, 3, 675, 380, -22)}
<g transform="translate(800 790)">
<g filter="url(#glow)" opacity=".95"><path d="${wordPath}" fill="none" stroke="url(#outer)" stroke-width="34" stroke-linejoin="round"/></g>
<path d="${wordPath}" fill="none" stroke="url(#outer)" stroke-width="22" stroke-linejoin="round"/>
<g mask="url(#cutmask)">
<path d="${wordPath}" transform="translate(5 9)" fill="url(#ext2)" stroke="url(#ext2)" stroke-width="6" stroke-linejoin="round"/>
<path d="${wordPath}" fill="none" stroke="#0a0626" stroke-width="10" stroke-linejoin="round"/>
<path d="${wordPath}" fill="url(#txt)"/>
</g>
<path d="${wordPath}" fill="none" stroke="#fff" stroke-opacity=".55" stroke-width="1.6" stroke-linejoin="round"/>
</g>
${bolts}
${sparks}
<path d="M240 862 Q800 826 1380 850 Q800 858 240 862Z" fill="url(#outer)"/>
</svg>`;
fs.writeFileSync(out + '.svg', body('<rect width="1600" height="900" fill="#070a24"/>'));
fs.writeFileSync(out + '-transparent.svg', body(''));
