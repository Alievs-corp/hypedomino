const fs = require('fs');
const dx = 46, dy = 46;
const P = {
  1: [[0, 0]], 2: [[-dx, -dy], [dx, dy]], 3: [[-dx, -dy], [0, 0], [dx, dy]],
  4: [[-dx, -dy], [dx, -dy], [-dx, dy], [dx, dy]], 5: [[-dx, -dy], [dx, -dy], [0, 0], [-dx, dy], [dx, dy]],
  6: [[-dx, -dy], [dx, -dy], [-dx, 0], [dx, 0], [-dx, dy], [dx, dy]],
};
const W = 200, H = 350, R = 40;
function pips(n, cy) {
  return P[n].map(([x, y]) => `<circle cx="${x}" cy="${cy + y}" r="19" fill="url(#pip)"/><circle cx="${x - 6}" cy="${cy + y - 7}" r="5" fill="#fff" opacity=".45"/>`).join('');
}
function tile(id, top, bottom, tx, ty, rot) {
  return `<g transform="translate(${tx} ${ty}) rotate(${rot})">
  <rect x="${-W / 2 + 16}" y="${-H / 2 + 24}" width="${W}" height="${H}" rx="${R}" fill="url(#ext)"/>
  <rect x="${-W / 2}" y="${-H / 2}" width="${W}" height="${H}" rx="${R}" fill="url(#face)" stroke="url(#rim)" stroke-width="9"/>
  <rect x="${-W / 2 + 10}" y="${-H / 2 + 10}" width="${W - 20}" height="${H - 20}" rx="${R - 10}" fill="none" stroke="#fff" stroke-opacity=".7" stroke-width="3"/>
  <rect x="${-W / 2 + 34}" y="-4" width="${W - 68}" height="8" rx="4" fill="#8a7bd8" opacity=".75"/>
  <circle cx="0" cy="0" r="9" fill="url(#rim)"/>
  ${pips(top, -88)}${pips(bottom, 88)}
</g>`;
}
const streak = (d, g) => `<path d="${d}" fill="url(#${g})"/>`;
const text = (attrs) => `<text x="0" y="0" text-anchor="middle" font-family="'Arial Black','Segoe UI Black',Impact,sans-serif" font-weight="900" font-size="204" letter-spacing="-4" ${attrs}>HypeDomino</text>`;
const body = (bg) => `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1600 900" width="1600" height="900">
<defs>
<linearGradient id="face" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="#cfdcff"/></linearGradient>
<linearGradient id="rim" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ff3ddc"/><stop offset="1" stop-color="#5a3bff"/></linearGradient>
<linearGradient id="ext" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#8a2cff"/><stop offset="1" stop-color="#2a1a9a"/></linearGradient>
<radialGradient id="pip" cx=".35" cy=".3" r=".9"><stop offset="0" stop-color="#4a3bb0"/><stop offset="1" stop-color="#0d0830"/></radialGradient>
<linearGradient id="txt" gradientUnits="userSpaceOnUse" x1="-560" y1="0" x2="560" y2="0"><stop offset="0" stop-color="#ffffff"/><stop offset=".36" stop-color="#cfe0ff"/><stop offset=".43" stop-color="#ff4fe0"/><stop offset="1" stop-color="#7a3bff"/></linearGradient>
<linearGradient id="outer" gradientUnits="userSpaceOnUse" x1="-560" y1="0" x2="560" y2="0"><stop offset="0" stop-color="#3b6bff"/><stop offset=".5" stop-color="#ff2fd0"/><stop offset="1" stop-color="#7a3bff"/></linearGradient>
<linearGradient id="s1" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#ff2fd0"/><stop offset="1" stop-color="#ff2fd0" stop-opacity="0"/></linearGradient>
<linearGradient id="s2" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#4a6bff"/><stop offset="1" stop-color="#4a6bff" stop-opacity="0"/></linearGradient>
<linearGradient id="s3" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#9a3bff"/><stop offset="1" stop-color="#9a3bff" stop-opacity="0"/></linearGradient>
<radialGradient id="halo" cx=".5" cy=".5" r=".5"><stop offset="0" stop-color="#8a2cff" stop-opacity=".55"/><stop offset="1" stop-color="#8a2cff" stop-opacity="0"/></radialGradient>
<filter id="glow" x="-20%" y="-40%" width="140%" height="180%"><feGaussianBlur stdDeviation="14"/></filter>
</defs>
${bg}
<ellipse cx="800" cy="330" rx="520" ry="300" fill="url(#halo)"/>
${streak('M180 360 L650 230 L650 262 Z', 's2')}
${streak('M120 430 L660 300 L660 340 Z', 's1')}
${streak('M240 300 L640 190 L640 214 Z', 's3')}
${streak('M1480 330 L1000 210 L1000 240 Z', 's1')}
${streak('M1520 420 L1010 290 L1010 326 Z', 's2')}
${streak('M1400 250 L1010 160 L1010 182 Z', 's3')}
${tile("a", 6, 3, 625, 320, -15)}
${tile("b", 5, 5, 975, 300, 12)}
<g transform="translate(800 740) skewX(-12)">
<g filter="url(#glow)" opacity=".85">${text('fill="none" stroke="url(#outer)" stroke-width="40"')}</g>
${text('fill="none" stroke="url(#outer)" stroke-width="42" stroke-linejoin="round"')}
${text('fill="none" stroke="#0a0626" stroke-width="22" stroke-linejoin="round"')}
${text('fill="url(#txt)"')}
</g>
<path d="M230 842 Q800 800 1390 826 Q800 836 230 842Z" fill="url(#outer)"/>
</svg>`;
fs.writeFileSync('hypedomino-logo.svg', body('<rect width="1600" height="900" fill="#070a24"/>'));
fs.writeFileSync('hypedomino-logo-transparent.svg', body(''));
