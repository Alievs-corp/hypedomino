import { readFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const textMinimumContrast = 4.5;
const nonTextMinimumContrast = 3;
const levels = ["low", "medium", "high"];
const effectNames = ["tilePlace", "scorePopup", "scoreCascade", "spinnerFormation", "domino", "block", "matchWin", "rankUp", "emoji"];
const budgetFields = ["particles", "drawCalls", "overdrawLayers", "pooled", "cpuMs", "gpuMs"];
const summedFields = [
  ["cpuMs", "effectCpuMs"],
  ["gpuMs", "effectGpuMs"],
  ["drawCalls", "effectDrawCalls"],
  ["particles", "maxParticles"],
];
const scalingFields = ["particles", "drawCalls"];
const requiredGroups = [
  "color", "opacity", "rarity", "typography", "radius", "spacing", "size", "elevation", "glow",
  "motion.duration", "motion.easing", "motion.shake", "effects.levels", "effects.budgets", "effects.concurrentSets", "scoreCascade",
];
const requiredGlyphs = "ƏəĞğIıİiŞşÇçÖöÜü";
const hexPattern = /^#[0-9A-F]{6}$/;
const epsilon = 1e-9;

const isObject = (value) => typeof value === "object" && value !== null && !Array.isArray(value);
const lookup = (root, path) => path.split(".").reduce((node, key) => (isObject(node) ? node[key] : undefined), root);
const round2 = (value) => Math.floor(value * 100) / 100;

function channel(byte) {
  const value = byte / 255;
  return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
}

export function luminance(hex) {
  const [r, g, b] = [1, 3, 5].map((start) => channel(parseInt(hex.slice(start, start + 2), 16)));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

export function contrast(first, second) {
  const [light, dark] = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (light + 0.05) / (dark + 0.05);
}

function backgroundLayers(color) {
  return [
    ["background.base", color.background.base],
    ["background.raised", color.background.raised],
    ["surface.default", color.surface.default],
    ["surface.raised", color.surface.raised],
  ];
}

export function contrastPairs(tokens) {
  const { color } = tokens;
  const layers = backgroundLayers(color);
  const pairs = [];
  for (const [name, value] of Object.entries(color.text)) {
    for (const [layer, background] of layers) pairs.push({ fg: `text.${name}`, bg: layer, fgHex: value, bgHex: background, min: textMinimumContrast });
  }
  for (const [role, shades] of Object.entries(color.role)) {
    pairs.push({ fg: `role.${role}.on`, bg: `role.${role}.fill`, fgHex: shades.on, bgHex: shades.fill, min: textMinimumContrast });
    for (const [layer, background] of layers) pairs.push({ fg: `role.${role}.text`, bg: layer, fgHex: shades.text, bgHex: background, min: textMinimumContrast });
    pairs.push({ fg: `role.${role}.fill`, bg: "background.base", fgHex: shades.fill, bgHex: color.background.base, min: nonTextMinimumContrast });
  }
  return pairs.map((pair) => ({ ...pair, ratio: round2(contrast(pair.fgHex, pair.bgHex)) }));
}

export function contrastTable(tokens) {
  const pairs = contrastPairs(tokens);
  const ratio = (fg, bg) => pairs.find((pair) => pair.fg === fg && pair.bg === bg).ratio.toFixed(2);
  const layerNames = backgroundLayers(tokens.color).map(([name]) => name);
  const textRows = Object.keys(tokens.color.text).map((name) => {
    const cells = layerNames.map((layer) => ratio(`text.${name}`, layer));
    return `| \`text.${name}\` | ${cells.join(" | ")} | - | - |`;
  });
  const roleRows = Object.keys(tokens.color.role).map((role) => {
    const cells = layerNames.map((layer) => ratio(`role.${role}.text`, layer));
    return `| \`role.${role}\` | ${cells.join(" | ")} | ${ratio(`role.${role}.on`, `role.${role}.fill`)} | ${ratio(`role.${role}.fill`, "background.base")} |`;
  });
  return [...textRows, ...roleRows];
}

const colorValues = (node) => Object.values(node).flatMap((value) => (isObject(value) ? colorValues(value) : [value]));

function colorErrors(color) {
  const errors = [];
  const visit = (node, path) => {
    for (const [key, value] of Object.entries(node)) {
      if (isObject(value)) visit(value, `${path}.${key}`);
      else if (!hexPattern.test(value)) errors.push(`${path}.${key}: ${value} is not #RRGGBB`);
    }
  };
  visit(color, "color");
  return errors;
}

function contrastErrors(tokens) {
  return contrastPairs(tokens)
    .filter(({ ratio, min }) => ratio < min)
    .map(({ fg, bg, ratio, min }) => `contrast ${fg} on ${bg} is ${ratio.toFixed(2)}, needs ${min}`);
}

function rarityErrors(rarity) {
  const keys = Object.keys(rarity);
  return keys.length === 1 && keys[0] === "definedBy" ? [] : ["rarity: only a definedBy reference is allowed here, values belong to #5"];
}

function typographyErrors(typography) {
  const errors = [];
  for (const [name, style] of Object.entries(typography.scale)) {
    const family = typography.family[style.family];
    if (!family) errors.push(`typography.scale.${name}: unknown family ${style.family}`);
    else if (!family.weights.includes(style.weight)) errors.push(`typography.scale.${name}: weight ${style.weight} not in ${family.name}`);
    if (style.size < typography.minimumSize) errors.push(`typography.scale.${name}: size ${style.size} under minimum ${typography.minimumSize}`);
  }
  return errors;
}

function motionErrors(motion) {
  const errors = [];
  for (const [name, value] of Object.entries(motion.duration)) {
    if (!Number.isInteger(value) || value < 0) errors.push(`motion.duration.${name}: ${value} is not a non-negative integer`);
  }
  for (const [name, points] of Object.entries(motion.easing)) {
    const valid = Array.isArray(points) && points.length === 4 && points.every(Number.isFinite) && [points[0], points[2]].every((x) => x >= 0 && x <= 1);
    if (!valid) errors.push(`motion.easing.${name}: needs four numbers with x1 and x2 in [0, 1]`);
  }
  return errors;
}

function levelErrors(name, level) {
  const errors = [];
  if (level.baseSceneCpuMs + level.effectCpuMs > level.workTimeP95Ms + epsilon) errors.push(`effects.levels.${name}: CPU base plus effects exceeds ${level.workTimeP95Ms} ms`);
  if (level.baseSceneGpuMs + level.effectGpuMs > level.workTimeP95Ms + epsilon) errors.push(`effects.levels.${name}: GPU base plus effects exceeds ${level.workTimeP95Ms} ms`);
  if (level.baseSceneBatches + level.effectDrawCalls > level.maxBatches) errors.push(`effects.levels.${name}: batches exceed ${level.maxBatches}`);
  return errors;
}

function effectErrors(effects) {
  const errors = [];
  for (const name of effectNames) {
    const effect = effects.budgets[name];
    if (!effect) {
      errors.push(`effects.budgets.${name}: missing`);
      continue;
    }
    if (!Number.isInteger(effect.durationMs) || effect.durationMs <= 0) errors.push(`effects.budgets.${name}.durationMs: must be a positive integer`);
    for (const levelName of levels) {
      const budget = effect[levelName];
      const level = effects.levels[levelName];
      if (!budget) {
        errors.push(`effects.budgets.${name}.${levelName}: missing`);
        continue;
      }
      for (const field of budgetFields) {
        if (typeof budget[field] !== "number" || budget[field] < 0) errors.push(`effects.budgets.${name}.${levelName}.${field}: must be a non-negative number`);
      }
      if (budget.overdrawLayers > level.maxOverdrawLayers) errors.push(`effects.budgets.${name}.${levelName}: overdraw ${budget.overdrawLayers} over ${level.maxOverdrawLayers}`);
    }
    for (const field of scalingFields) {
      const values = levels.map((levelName) => effect[levelName]?.[field]);
      if (values.some((value, index) => index > 0 && value < values[index - 1])) errors.push(`effects.budgets.${name}.${field}: must not decrease from low to high`);
    }
  }
  for (const [setName, members] of Object.entries(effects.concurrentSets)) {
    for (const member of members) if (!effects.budgets[member]) errors.push(`effects.concurrentSets.${setName}: unknown effect ${member}`);
    for (const levelName of levels) {
      const level = effects.levels[levelName];
      for (const [field, limitField] of summedFields) {
        const total = members.reduce((sum, member) => sum + (effects.budgets[member]?.[levelName]?.[field] ?? 0), 0);
        if (total > level[limitField] + epsilon) errors.push(`effects.concurrentSets.${setName}.${levelName}: ${field} ${round2(total)} over ${limitField} ${level[limitField]}`);
      }
    }
  }
  return errors;
}

function cascadeErrors(tokens) {
  const { steps, chainBoost, popIn, popOut } = tokens.scoreCascade;
  const errors = [];
  steps.forEach((step, index) => {
    const previous = steps[index - 1];
    if (step.step !== index + 1) errors.push(`scoreCascade.steps[${index}]: step must be ${index + 1}`);
    if (step.minPoints % 5 !== 0) errors.push(`scoreCascade.steps[${index}]: minPoints must be a multiple of 5`);
    if (!tokens.glow[step.glow]) errors.push(`scoreCascade.steps[${index}]: unknown glow ${step.glow}`);
    if (!tokens.motion.shake[step.shake]) errors.push(`scoreCascade.steps[${index}]: unknown shake ${step.shake}`);
    if (step.particleShare < 0 || step.particleShare > 1) errors.push(`scoreCascade.steps[${index}]: particleShare outside [0, 1]`);
    if (!previous) return;
    if (step.minPoints <= previous.minPoints) errors.push(`scoreCascade.steps[${index}]: minPoints must increase`);
    for (const field of ["textScale", "particleShare", "holdMs"]) {
      if (step[field] < previous[field]) errors.push(`scoreCascade.steps[${index}]: ${field} must not decrease`);
    }
  });
  chainBoost.forEach((boost, index) => {
    if (boost.placement !== index + 1) errors.push(`scoreCascade.chainBoost[${index}]: placement must be ${index + 1}`);
    if (index > 0 && boost.addSteps < chainBoost[index - 1].addSteps) errors.push(`scoreCascade.chainBoost[${index}]: addSteps must not decrease`);
  });
  for (const [name, transition] of Object.entries({ popIn, popOut })) {
    if (tokens.motion.duration[transition.duration] === undefined) errors.push(`scoreCascade.${name}: unknown duration ${transition.duration}`);
    if (!tokens.motion.easing[transition.easing]) errors.push(`scoreCascade.${name}: unknown easing ${transition.easing}`);
  }
  const holdMax = Math.max(...steps.map((step) => step.holdMs));
  const popupMax = tokens.motion.duration[popIn.duration] + holdMax + tokens.motion.duration[popOut.duration];
  if (tokens.effects.budgets.scorePopup?.durationMs !== popupMax) errors.push(`effects.budgets.scorePopup.durationMs must equal pop in + longest hold + pop out (${popupMax})`);
  return errors;
}

export function stepForPoints(tokens, points, placement = 1) {
  const { steps, chainBoost } = tokens.scoreCascade;
  const base = steps.filter((step) => points >= step.minPoints).length;
  if (base === 0) return 0;
  const boost = chainBoost[Math.min(placement, chainBoost.length) - 1].addSteps;
  return Math.min(base + boost, steps.length);
}

export function checkTokens(tokens, specText) {
  const missing = requiredGroups.filter((path) => lookup(tokens, path) === undefined).map((path) => `missing group ${path}`);
  if (missing.length > 0) return missing;
  const errors = [
    ...colorErrors(tokens.color),
    ...contrastErrors(tokens),
    ...rarityErrors(tokens.rarity),
    ...typographyErrors(tokens.typography),
    ...motionErrors(tokens.motion),
    ...levels.flatMap((name) => (tokens.effects.levels[name] ? levelErrors(name, tokens.effects.levels[name]) : [`effects.levels.${name}: missing`])),
  ];
  if (errors.length > 0) return errors;
  errors.push(...effectErrors(tokens.effects), ...cascadeErrors(tokens));
  for (const row of contrastTable(tokens)) if (!specText.includes(row)) errors.push(`spec contrast table is missing or stale: ${row}`);
  if (!specText.includes(requiredGlyphs)) errors.push("spec does not list the required Azerbaijani and Turkish glyphs");
  for (const hex of colorValues(tokens.color)) if (!specText.includes(`\`${hex}\``)) errors.push(`spec does not list colour ${hex}`);
  return errors;
}

export const loadTokens = (path) => JSON.parse(readFileSync(path, "utf8"));
export const loadSpec = (path) => readFileSync(path, "utf8");

function main(args) {
  const tokens = loadTokens("docs/design-tokens.json");
  if (args.includes("--table")) {
    console.log(contrastTable(tokens).join("\n"));
    return;
  }
  const errors = checkTokens(tokens, loadSpec("docs/ui-motion.md"));
  for (const error of errors) console.error(error);
  if (errors.length > 0) process.exit(1);
  console.log("design tokens ok");
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main(process.argv.slice(2));
