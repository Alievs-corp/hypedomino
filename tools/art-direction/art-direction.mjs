import { readFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const colorLiteralPatterns = [/#[0-9A-Fa-f]{6}(?:[0-9A-Fa-f]{2})?\b/g, /\b(?:rgba?|hsla?)\s*\(/gi];
const codeSpanPattern = /`([^`\n]+)`/g;
const tokenPathPattern = /^[A-Za-z][A-Za-z0-9]*(?:\.[A-Za-z0-9]+)*$/;

const isObject = (value) => typeof value === "object" && value !== null && !Array.isArray(value);
const lookup = (root, path) => path.split(".").reduce((node, key) => (isObject(node) ? node[key] : undefined), root);

function colorLiteralErrors(text) {
  return colorLiteralPatterns.flatMap((pattern) => [...text.matchAll(pattern)].map((match) => `colour literal ${match[0]} found; name a token instead`));
}

function tokenPathErrors(tokens, text) {
  const groups = new Set(Object.keys(tokens).filter((key) => key !== "meta"));
  const paths = [...text.matchAll(codeSpanPattern)].map((match) => match[1]);
  const tokenPaths = paths.filter((path) => tokenPathPattern.test(path) && groups.has(path.split(".")[0]));
  return [...new Set(tokenPaths)].filter((path) => lookup(tokens, path) === undefined).map((path) => `unknown token ${path}`);
}

export function checkArtDirection(tokens, text) {
  return [...colorLiteralErrors(text), ...tokenPathErrors(tokens, text)];
}

function main() {
  const tokens = JSON.parse(readFileSync("docs/design-tokens.json", "utf8"));
  const errors = checkArtDirection(tokens, readFileSync("docs/art-direction.md", "utf8"));
  for (const error of errors) console.error(error);
  if (errors.length > 0) process.exit(1);
  console.log("art direction ok");
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main();
