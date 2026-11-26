import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { pathToFileURL } from "node:url";

export const manifestName = "files.json";
export const uvcsRepository = "hypedomino-art";
export const sourceArtExtensions = ["psd", "psb", "ai", "aep", "prproj", "blend", "fbx", "obj", "glb", "exr", "tif", "tiff", "png", "jpg", "jpeg", "webp", "mp4", "mov", "wav", "aif", "aiff", "flac", "mp3", "ogg", "zip"];
const lfsPointer = /^version https:\/\/git-lfs\.github\.com\/spec\/v1\noid sha256:([0-9a-f]{64})\nsize \d+\n?$/;
const sha256Pattern = /^[0-9a-f]{64}$/;
const entryKeys = ["path", "sha256", "changeset"];

export const isSourceArt = (path) => sourceArtExtensions.includes(path.slice(path.lastIndexOf(".") + 1).toLowerCase());

export function digestOf(bytes) {
  const pointer = bytes.length < 200 ? bytes.toString("latin1").match(lfsPointer) : null;
  if (pointer) return { sha256: pointer[1], pointer: true };
  return { sha256: createHash("sha256").update(bytes).digest("hex"), pointer: false };
}

function listTree(directory, prefix = "") {
  return readdirSync(join(directory, prefix), { withFileTypes: true }).flatMap((entry) => {
    const path = prefix ? `${prefix}/${entry.name}` : entry.name;
    return entry.isDirectory() ? listTree(directory, path) : [path];
  });
}

export function findItems(root) {
  const absolute = resolve(root);
  if (existsSync(join(absolute, "provenance.md"))) return [absolute];
  return listTree(absolute)
    .filter((path) => path.endsWith("/provenance.md"))
    .map((path) => join(absolute, dirname(path)))
    .sort();
}

const sourceArtOnDisk = (item) => listTree(item).filter(isSourceArt).sort();

function pathErrors(path) {
  if (typeof path !== "string" || path.length === 0) return ["path must be a non-empty string"];
  if (path.includes("\\") || path.startsWith("/")) return [`${path}: path must be relative with forward slashes`];
  if (path.split("/").some((segment) => segment === "" || segment === "." || segment === "..")) return [`${path}: path has an empty, . or .. segment`];
  if (path === manifestName || !isSourceArt(path)) return [`${path}: not a source art file (docs/adr/0015-source-art-in-unity-version-control.md)`];
  return [];
}

function entryErrors(entry) {
  if (entry === null || typeof entry !== "object" || Array.isArray(entry)) return ["every entry must be an object"];
  const keys = Object.keys(entry).sort();
  if (keys.join() !== [...entryKeys].sort().join()) return [`${entry.path}: entry needs exactly ${entryKeys.join(", ")}`];
  const errors = pathErrors(entry.path);
  if (!sha256Pattern.test(entry.sha256)) errors.push(`${entry.path}: sha256 must be 64 lowercase hex digits`);
  if (!Number.isInteger(entry.changeset) || entry.changeset < 1) errors.push(`${entry.path}: changeset must be a whole number from 1`);
  return errors;
}

export function manifestErrors(manifest) {
  if (manifest === null || typeof manifest !== "object" || Array.isArray(manifest) || Object.keys(manifest).join() !== "files" || !Array.isArray(manifest.files)) return ["manifest must be an object with only a files array"];
  const errors = manifest.files.flatMap(entryErrors);
  const paths = manifest.files.map((entry) => entry?.path);
  paths.forEach((path, index) => {
    if (index > 0 && !(paths[index - 1] < path)) errors.push(`${path}: files must be sorted by path without duplicates`);
  });
  return errors;
}

export function readManifest(item) {
  const path = join(item, manifestName);
  if (!existsSync(path)) return undefined;
  try {
    return JSON.parse(readFileSync(path, "utf8"));
  } catch (error) {
    return { invalid: error.message };
  }
}

const digestOnDisk = (item, path) => digestOf(readFileSync(join(item, path)));

export function buildManifest(item, previous, { changeset, prune = false } = {}) {
  const earlier = new Map((previous?.files ?? []).map((entry) => [entry.path, entry]));
  const onDisk = sourceArtOnDisk(item);
  const errors = [];
  const files = onDisk.map((path) => {
    const { sha256 } = digestOnDisk(item, path);
    const kept = earlier.get(path);
    if (kept?.sha256 === sha256) return kept;
    if (changeset === undefined) errors.push(`${path}: new or changed, pass --changeset <n> of the Unity Version Control check-in`);
    return { path, sha256, changeset };
  });
  const absent = prune ? [] : [...earlier.values()].filter((entry) => !onDisk.includes(entry.path));
  const all = [...files, ...absent].sort((a, b) => (a.path < b.path ? -1 : 1));
  return { manifest: { files: all }, errors };
}

const serialize = (manifest) => `${JSON.stringify(manifest, null, 2)}\n`;

export function writeItem(item, options) {
  const previous = readManifest(item);
  if (previous?.invalid) return { errors: [`${manifestName}: ${previous.invalid}`] };
  const { manifest, errors } = buildManifest(item, previous, options);
  if (errors.length > 0) return { errors };
  if (manifest.files.length === 0 && !previous) return { errors: [], change: "none" };
  const text = serialize(manifest);
  const path = join(item, manifestName);
  if (previous && readFileSync(path, "utf8") === text) return { errors: [], change: "unchanged", count: manifest.files.length };
  writeFileSync(path, text);
  return { errors: [], change: previous ? "updated" : "written", count: manifest.files.length };
}

export function verifyItem(item, { complete = false } = {}) {
  const manifest = readManifest(item);
  if (!manifest) return { errors: [], skipped: true };
  if (manifest.invalid) return { errors: [`${manifestName}: ${manifest.invalid}`] };
  const errors = manifestErrors(manifest);
  if (errors.length > 0) return { errors };
  const listed = new Map(manifest.files.map((entry) => [entry.path, entry]));
  for (const path of sourceArtOnDisk(item)) {
    const entry = listed.get(path);
    if (!entry) errors.push(`${path}: on disk but not in ${manifestName}, run asset-manifest write`);
    else if (digestOnDisk(item, path).sha256 !== entry.sha256) errors.push(`${path}: sha256 differs from ${manifestName}`);
  }
  if (complete) manifest.files.filter((entry) => !existsSync(join(item, entry.path))).forEach((entry) => errors.push(`${entry.path}: listed but not on disk`));
  return { errors, count: manifest.files.length };
}

export function listedPaths(root, files) {
  return files
    .filter((path) => path === manifestName || path.endsWith(`/${manifestName}`))
    .flatMap((path) => {
      const manifest = readManifest(join(root, dirname(path)));
      if (!manifest || manifest.invalid || !Array.isArray(manifest.files)) return [];
      const prefix = dirname(path) === "." ? "" : `${dirname(path).replaceAll("\\", "/")}/`;
      return manifest.files.map((entry) => `${prefix}${entry.path}`);
    });
}

export function serverPath(item, path) {
  const segments = resolve(item).split(/[\\/]/);
  const at = segments.lastIndexOf("assets-src");
  if (at < 0) throw new Error(`${item} is not under assets-src`);
  return `/${[...segments.slice(at + 1), path].join("/")}`;
}

export const revisionSpec = (item, entry, server) => `serverpath:${serverPath(item, entry.path)}#cs:${entry.changeset}@${uvcsRepository}@${server}`;

export const runCm = (args) => execFileSync("cm", args, { stdio: ["ignore", "inherit", "inherit"] });

export function pullItem(item, { server, only = "", run = runCm }) {
  const manifest = readManifest(item);
  if (!manifest) return { errors: [], pulled: 0 };
  const errors = manifest.invalid ? [`${manifestName}: ${manifest.invalid}`] : manifestErrors(manifest);
  if (errors.length > 0) return { errors, pulled: 0 };
  let pulled = 0;
  for (const entry of manifest.files.filter((file) => file.path.startsWith(only))) {
    const target = join(item, entry.path);
    if (existsSync(target)) {
      const present = digestOnDisk(item, entry.path);
      if (!present.pointer && present.sha256 === entry.sha256) continue;
    }
    mkdirSync(dirname(target), { recursive: true });
    run(["cat", revisionSpec(item, entry, server), `--file=${target}`]);
    pulled++;
    if (digestOnDisk(item, entry.path).sha256 !== entry.sha256) errors.push(`${entry.path}: changeset ${entry.changeset} does not have the sha256 of ${manifestName}`);
  }
  return { errors, pulled };
}

export function parseArguments(argv) {
  const [command, ...rest] = argv;
  const options = { command, paths: [], prune: false, complete: false };
  for (let at = 0; at < rest.length; at++) {
    const argument = rest[at];
    if (argument === "--changeset") {
      const value = Number(rest[++at]);
      if (!Number.isInteger(value) || value < 1) throw new Error("--changeset needs a whole number from 1");
      options.changeset = value;
    } else if (argument === "--prune") options.prune = true;
    else if (argument === "--complete") options.complete = true;
    else if (argument.startsWith("--")) throw new Error(`unknown option ${argument}`);
    else options.paths.push(argument);
  }
  return options;
}

const usage = "usage: asset-manifest.mjs write [--changeset <n>] [--prune] <assets-src or item folder>...\n       asset-manifest.mjs verify [--complete] <assets-src or item folder>...";

function main(argv) {
  const options = parseArguments(argv);
  if (!["write", "verify"].includes(options.command) || options.paths.length === 0) throw new Error(usage);
  const items = options.paths.flatMap(findItems);
  let failed = false;
  let manifests = 0;
  for (const item of items) {
    const result = options.command === "write" ? writeItem(item, options) : verifyItem(item, options);
    const name = relative(process.cwd(), item).replaceAll("\\", "/");
    for (const error of result.errors) console.error(`${name}: ${error}`);
    failed ||= result.errors.length > 0;
    if (result.count !== undefined) manifests++;
    if (result.change === "written" || result.change === "updated") console.log(`${name}/${manifestName}: ${result.change}, ${result.count} files`);
  }
  if (failed) return 1;
  console.log(`asset manifest ${options.command} ok (${items.length} items, ${manifests} with ${manifestName})`);
  return 0;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    process.exitCode = main(process.argv.slice(2));
  } catch (error) {
    console.error(error.message);
    process.exitCode = 2;
  }
}
