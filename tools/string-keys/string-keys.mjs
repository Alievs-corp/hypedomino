import { readdirSync, readFileSync } from "node:fs";
import { join, relative } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

export const clientRoot = "client/Assets/HypeDomino";
export const builtinStringsPath = `${clientRoot}/UI/Localization/Resources/builtin-strings.json`;
export const requiredLanguages = ["en", "az"];
const nonRuntimeFolders = new Set(["Editor", "Generated", "Tests"]);
const tokenFiles = new Set(["DesignTokens.cs"]);
const keyLiteral = /"([a-z0-9_]+(?:\.[a-z0-9_]+){2,})"/g;
const resourceKey = /"\w*Key"\s*:\s*"([a-z0-9_]+(?:\.[a-z0-9_]+)+)"/g;

function filesUnder(folder, accept) {
  const found = [];
  for (const entry of readdirSync(folder, { withFileTypes: true })) {
    const path = join(folder, entry.name);
    if (entry.isDirectory()) {
      if (!nonRuntimeFolders.has(entry.name)) found.push(...filesUnder(path, accept));
    } else if (accept(path, entry.name)) found.push(path);
  }
  return found;
}

export function codeKeys(source) {
  return [...source.matchAll(keyLiteral)].map((match) => match[1]);
}

export function resourceKeys(json) {
  return [...json.matchAll(resourceKey)].map((match) => match[1]);
}

function isRuntimeSource(_, name) {
  return name.endsWith(".cs") && !tokenFiles.has(name);
}

function isResourceData(path, name) {
  return name.endsWith(".json") && path.split(/[\\/]/).includes("Resources") && name !== "builtin-strings.json";
}

export function referencedKeys(repoRoot) {
  const root = join(repoRoot, clientRoot);
  const references = new Map();
  const note = (key, path) => {
    if (!references.has(key)) references.set(key, relative(repoRoot, path).replaceAll("\\", "/"));
  };
  for (const path of filesUnder(root, isRuntimeSource)) for (const key of codeKeys(readFileSync(path, "utf8"))) note(key, path);
  for (const path of filesUnder(root, isResourceData)) for (const key of resourceKeys(readFileSync(path, "utf8"))) note(key, path);
  return references;
}

export function missingTexts(references, entries) {
  const byKey = new Map(entries.map((entry) => [entry.key, entry]));
  const errors = [];
  for (const [key, path] of [...references].sort(([a], [b]) => a.localeCompare(b))) {
    const entry = byKey.get(key);
    const missing = requiredLanguages.filter((language) => !entry?.[language]);
    if (missing.length > 0) errors.push(`${key} (${path}) has no ${missing.join(" or ")} text in ${builtinStringsPath}`);
  }
  return errors;
}

export function checkStringKeys(repoRoot) {
  const { entries } = JSON.parse(readFileSync(join(repoRoot, builtinStringsPath), "utf8"));
  return missingTexts(referencedKeys(repoRoot), entries);
}

function main() {
  const errors = checkStringKeys(fileURLToPath(new URL("../../", import.meta.url)));
  for (const error of errors) console.error(error);
  if (errors.length > 0) process.exitCode = 1;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main();
