const fs = require('fs');
const path = require('path');
const opentype = require(process.env.OT || 'opentype.js');
const { flatten, sharpen, toD } = require('./sharpen.js');
const fontFile = process.argv[2] || path.join(__dirname, 'fonts/Kanit-BlackItalic.ttf');
const out = process.argv[3] || 'hypedomino-logo';
const raw = fs.readFileSync(fontFile);
const font = opentype.parse(raw.buffer.slice(raw.byteOffset, raw.byteOffset + raw.byteLength));

const dx = 46, dy = 46;
const P = {
  1: [[0, 0]], 2: [[-dx, -dy], [dx, dy]], 3: [[-dx, -dy], [0, 0], [dx, dy]],
  4: [[-dx, -dy], [dx, -dy], [-dx, dy], [dx, dy]], 5: [[-dx, -dy], [dx, -dy], [0, 0], [-dx, dy], [dx, dy]],
  6: [[-dx, -dy], [dx, -dy], [-dx, 0], [dx, 0], [-dx, dy], [dx, dy]],
};
const W = 200, H = 350, R = 40;
const pips = (n, cy) => P[n].map(([x, y]) => `<circle cx="${x}" cy="${cy + y}" r="19" fill="url(#pip)"/><circle cx="${x - 6}" cy="${cy + y - 7}" r="5" fill="#fff" opacity=".45"/>`).join('');
function tile(top, bottom, tx, ty, rot, scale, shade) {
  return `<g transform="translate(${tx} ${ty}) rotate(${rot}) scale(${scale})">
  <rect x="${-W / 2 + 26}" y="${-H / 2 + 16}" width="${W}" height="${H}" rx="${R}" fill="url(#ext)"/>
  <rect x="${-W / 2 + 13}" y="${-H / 2 + 8}" width="${W}" height="${H}" rx="${R}" fill="url(#ext)"/>
  <rect x="${-W / 2}" y="${-H / 2}" width="${W}" height="${H}" rx="${R}" fill="url(#face)" stroke="url(#rim)" stroke-width="9"/>
  <rect x="${-W / 2 + 10}" y="${-H / 2 + 10}" width="${W - 20}" height="${H - 20}" rx="${R - 10}" fill="none" stroke="#fff" stroke-opacity=".7" stroke-width="3"/>
  <rect x="${-W / 2 + 24}" y="-4" width="${W - 48}" height="8" rx="4" fill="#8a7bd8" opacity=".8"/>
  <circle cx="0" cy="0" r="9" fill="url(#rim)"/>
  ${pips(top, -88)}${pips(bottom, 88)}
  ${shade ? `<rect x="${-W / 2}" y="${-H / 2}" width="${W}" height="${H}" rx="${R}" fill="#0a0626" opacity="${shade}"/>` : ''}
</g>`;
}

const SIZE = 215, WORD = 'HypeDomino';
const adv = font.getAdvanceWidth(WORD, SIZE);
const half = adv / 2;
const glyphPath = font.getPath(WORD, -half, 0, SIZE);
const wordD = toD(flatten(glyphPath.commands, 8).map(c => sharpen(c, { spacing: 22, amp: 0.25, chip: 0.012, spike: 0.006, notch: 9 })));

const advAt = t => font.getAdvanceWidth(t, SIZE);
const slashX = [-half + advAt('Hype') + 6, -half + advAt('HypeD') + 4];
const slashes = slashX.map(x => `<path d="M${x + 70} -190 L${x + 62} -190 L${x - 38} 70 L${x - 30} 70 Z" fill="#000"/>`).join('');
const blades = `
<path d="M${-half - 170} -4 L${-half + 60} -84 L${-half + 60} -36 Z"/>
<path d="M${-half - 100} 50 L${-half + 40} 6 L${-half + 40} 34 Z"/>
<path d="M${half - 10} -2 L${half + 170} -70 L${half + 30} -30 Z"/>`;

const streaks = [
  'M180 300 L450 270 L580 292 L430 312 L560 328 L400 344 L580 370 L300 372 Z',
  'M110 400 L420 360 L540 382 L400 400 L540 416 L380 434 L580 462 L220 456 Z',
  'M260 215 L500 205 L620 226 L470 236 L600 252 L430 262 Z',
];
const sg = ['s1', 's2', 's3'];

const body = (bg) => `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1600 900" width="1600" height="900">
<defs>
<linearGradient id="face" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="#cfdcff"/></linearGradient>
<linearGradient id="rim" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#2ee6ff"/><stop offset=".5" stop-color="#ff3ddc"/><stop offset="1" stop-color="#5a3bff"/></linearGradient>
<linearGradient id="ext" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ff3ddc"/><stop offset=".5" stop-color="#8a2cff"/><stop offset="1" stop-color="#2a1a9a"/></linearGradient>
<radialGradient id="pip" cx=".35" cy=".3" r=".9"><stop offset="0" stop-color="#4a3bb0"/><stop offset="1" stop-color="#0d0830"/></radialGradient>
<linearGradient id="txt" gradientUnits="userSpaceOnUse" x1="${-half}" y1="-160" x2="${half}" y2="40"><stop offset="0" stop-color="#ffffff"/><stop offset=".3" stop-color="#bfeeff"/><stop offset=".42" stop-color="#33d0ff"/><stop offset=".7" stop-color="#3b7bff"/><stop offset="1" stop-color="#9a3bff"/></linearGradient>
<linearGradient id="hi" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#fff" stop-opacity=".9"/><stop offset=".5" stop-color="#fff" stop-opacity="0"/></linearGradient>
<linearGradient id="outer" gradientUnits="userSpaceOnUse" x1="${-half}" y1="0" x2="${half}" y2="0"><stop offset="0" stop-color="#18c8ff"/><stop offset=".5" stop-color="#3b6bff"/><stop offset="1" stop-color="#b03bff"/></linearGradient>
<linearGradient id="s1" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#ff2fd0"/><stop offset="1" stop-color="#ff2fd0" stop-opacity="0"/></linearGradient>
<linearGradient id="s2" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#2ee6ff"/><stop offset="1" stop-color="#2ee6ff" stop-opacity="0"/></linearGradient>
<linearGradient id="s3" x1="1" y1="0" x2="0" y2="0"><stop offset="0" stop-color="#9a3bff"/><stop offset="1" stop-color="#9a3bff" stop-opacity="0"/></linearGradient>
<radialGradient id="halo" cx=".5" cy=".5" r=".5"><stop offset="0" stop-color="#2a6bff" stop-opacity=".5"/><stop offset=".6" stop-color="#8a2cff" stop-opacity=".25"/><stop offset="1" stop-color="#8a2cff" stop-opacity="0"/></radialGradient>
<filter id="glow" x="-20%" y="-60%" width="140%" height="220%"><feGaussianBlur stdDeviation="16"/></filter>
<mask id="cutmask" maskUnits="userSpaceOnUse" x="${-half - 200}" y="-260" width="${adv + 400}" height="420"><rect x="${-half - 200}" y="-260" width="${adv + 400}" height="420" fill="#fff"/>${slashes}</mask>
</defs>
${bg}
<ellipse cx="800" cy="340" rx="640" ry="340" fill="url(#halo)"/>
${streaks.map((d, i) => `<path d="${d}" fill="url(#${sg[i]})"/>`).join('\n')}
${tile(5, 5, 935, 335, -11, 0.95, 0)}
${tile(6, 3, 765, 295, -17, 1.15, 0)}
<g transform="translate(800 790)">
<g transform="skewX(-4)">
<g filter="url(#glow)" opacity=".95"><path d="${wordD}" fill="none" stroke="url(#outer)" stroke-width="36" stroke-linejoin="round"/><g fill="url(#outer)">${blades}</g></g>
<path d="${wordD}" fill="none" stroke="url(#outer)" stroke-width="24" stroke-linejoin="round"/>
<g fill="url(#outer)" stroke="url(#outer)" stroke-width="12" stroke-linejoin="round">${blades}</g>
<g mask="url(#cutmask)">
<path d="${wordD}" fill="none" stroke="#0a0626" stroke-width="11" stroke-linejoin="round"/>
<path d="${wordD}" fill="url(#txt)"/>
</g>
<g fill="url(#txt)" stroke="#0a0626" stroke-width="3" stroke-linejoin="round">${blades}</g>
</g>
</g>
<path d="M240 862 Q800 826 1380 850 Q800 858 240 862Z" fill="url(#outer)"/>
</svg>`;
fs.writeFileSync(out + '.svg', body('<rect width="1600" height="900" fill="#070a24"/>'));
fs.writeFileSync(out + '-transparent.svg', body(''));
