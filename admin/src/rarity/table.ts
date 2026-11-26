export type RarityTier = { ordinal: number; nameKey: string; fill: string; on: string; text: string; pips: number; glow: string; flow: boolean };

export type RarityTable = { diamondMinSize: number; flowPeriodMs: number; tiers: Record<string, RarityTier> };

export type TierFields = { id: string; ordinal: string; fill: string; on: string; text: string; pips: string; glow: string; flow: boolean };

export type TableFields = { diamondMinSize: string; flowPeriodMs: string; tiers: TierFields[] };

export const shades = ["fill", "on", "text"] as const;

export const maxPips = 6;

const positiveInteger = /^[1-9]\d*$/;
const hexColor = /^#[0-9A-Fa-f]{6}$/;

export function fieldsFromTable(table: RarityTable): TableFields {
  const tiers = Object.entries(table.tiers)
    .sort(([, a], [, b]) => a.ordinal - b.ordinal)
    .map(([id, tier]) => ({ id, ordinal: String(tier.ordinal), fill: tier.fill, on: tier.on, text: tier.text, pips: String(tier.pips), glow: tier.glow, flow: tier.flow }));
  return { diamondMinSize: String(table.diamondMinSize), flowPeriodMs: String(table.flowPeriodMs), tiers };
}

export function tableFromFields(fields: TableFields): RarityTable {
  const tiers: Record<string, RarityTier> = {};
  for (const tier of fields.tiers) {
    tiers[tier.id] = { ordinal: Number(tier.ordinal), nameKey: `rarity.${tier.id}.name`, fill: tier.fill, on: tier.on, text: tier.text, pips: Number(tier.pips), glow: tier.glow, flow: tier.flow };
  }
  return { diamondMinSize: Number(fields.diamondMinSize), flowPeriodMs: Number(fields.flowPeriodMs), tiers };
}

export function badFields(fields: TableFields, glowLevels: readonly string[]): string[] {
  const bad: string[] = (["diamondMinSize", "flowPeriodMs"] as const).filter((name) => !positiveInteger.test(fields[name]));
  for (const tier of fields.tiers) {
    if (!positiveInteger.test(tier.ordinal)) bad.push(`${tier.id}.ordinal`);
    for (const shade of shades) if (!hexColor.test(tier[shade])) bad.push(`${tier.id}.${shade}`);
    if (!positiveInteger.test(tier.pips) || Number(tier.pips) > maxPips) bad.push(`${tier.id}.pips`);
    if (!glowLevels.includes(tier.glow)) bad.push(`${tier.id}.glow`);
  }
  return bad;
}
