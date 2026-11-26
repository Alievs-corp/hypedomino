import { after, test } from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, unlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { digestOf, findItems, listedPaths, manifestErrors, manifestName, parseArguments, pullItem, readManifest, revisionSpec, sourceArtExtensions, verifyItem, writeItem } from "./asset-manifest.mjs";

const temporaryRoots = [];
after(() => temporaryRoots.forEach((root) => rmSync(root, { recursive: true, force: true })));
const sha = (text) => createHash("sha256").update(text).digest("hex");
const pointer = (text) => `version https://git-lfs.github.com/spec/v1\noid sha256:${sha(text)}\nsize ${text.length}\n`;

function put(root, path, content) {
  mkdirSync(dirname(join(root, path)), { recursive: true });
  writeFileSync(join(root, path), content);
}

function tree(files) {
  const root = mkdtempSync(join(tmpdir(), "asset-manifest-"));
  temporaryRoots.push(root);
  const assets = join(root, "assets-src");
  for (const [path, content] of Object.entries(files)) put(assets, path, content);
  return assets;
}

function avatar() {
  const assets = tree({
    "cosmetic/avatar/x/provenance.md": "# avatar.x\n",
    "cosmetic/avatar/x/export/static.png": "static",
    "cosmetic/avatar/x/work/x.psd": "layers",
    "cosmetic/avatar/x/work/x.py": "print()",
  });
  return join(assets, "cosmetic/avatar/x");
}

const entry = (path, text, changeset) => ({ path, sha256: sha(text), changeset });
const manifestOf = (item) => JSON.parse(readFileSync(join(item, manifestName), "utf8"));

test("digest is the content sha256, or the oid of a Git LFS pointer", () => {
  assert.deepEqual(digestOf(Buffer.from("static")), { sha256: sha("static"), pointer: false });
  assert.deepEqual(digestOf(Buffer.from(pointer("static"))), { sha256: sha("static"), pointer: true });
});

test("every assets-src Git LFS rule of .gitattributes is source art", () => {
  const attributes = readFileSync(fileURLToPath(new URL("../../.gitattributes", import.meta.url)), "utf8");
  const extensions = [...attributes.matchAll(/^assets-src\/\*\*\/\*\.(\w+) filter=lfs/gm)].map((match) => match[1]);
  assert.deepEqual(extensions.filter((extension) => !sourceArtExtensions.includes(extension)), []);
});

test("write lists the source art of an item with the changeset of its check-in", () => {
  const item = avatar();
  assert.match(writeItem(item, {}).errors[0], /--changeset/);
  assert.deepEqual(writeItem(item, { changeset: 5 }), { errors: [], change: "written", count: 2 });
  assert.deepEqual(manifestOf(item), { files: [entry("export/static.png", "static", 5), entry("work/x.psd", "layers", 5)] });
  assert.equal(writeItem(item, {}).change, "unchanged");
  writeFileSync(join(item, "work/x.psd"), "layers 2");
  assert.equal(writeItem(item, { changeset: 7 }).change, "updated");
  assert.deepEqual(manifestOf(item).files, [entry("export/static.png", "static", 5), entry("work/x.psd", "layers 2", 7)]);
});

test("write keeps a listed file that is not on disk unless pruned", () => {
  const item = avatar();
  writeItem(item, { changeset: 5 });
  unlinkSync(join(item, "work/x.psd"));
  writeItem(item, {});
  assert.equal(manifestOf(item).files.length, 2);
  writeItem(item, { prune: true });
  assert.deepEqual(manifestOf(item).files, [entry("export/static.png", "static", 5)]);
});

test("write reads the oid of a Git LFS pointer without the content", () => {
  const item = avatar();
  writeFileSync(join(item, "export/static.png"), pointer("static"));
  writeItem(item, { changeset: 3 });
  assert.deepEqual(manifestOf(item).files[0], entry("export/static.png", "static", 3));
});

test("an item without source art and without a manifest gets none", () => {
  const assets = tree({ "medal/m/provenance.md": "# m\n" });
  assert.equal(writeItem(join(assets, "medal/m"), {}).change, "none");
  assert.deepEqual(verifyItem(join(assets, "medal/m")), { errors: [], skipped: true });
});

test("verify compares the files on disk with the manifest", () => {
  const item = avatar();
  writeItem(item, { changeset: 5 });
  assert.deepEqual(verifyItem(item), { errors: [], count: 2 });
  writeFileSync(join(item, "export/static.png"), pointer("static"));
  assert.deepEqual(verifyItem(item).errors, []);
  writeFileSync(join(item, "export/static.png"), "edited");
  put(item, "export/anim/0001.png", "frame");
  assert.deepEqual(verifyItem(item).errors, ["export/anim/0001.png: on disk but not in files.json, run asset-manifest write", "export/static.png: sha256 differs from files.json"]);
});

test("verify --complete needs every listed file on disk", () => {
  const item = avatar();
  writeItem(item, { changeset: 5 });
  unlinkSync(join(item, "work/x.psd"));
  assert.deepEqual(verifyItem(item).errors, []);
  assert.deepEqual(verifyItem(item, { complete: true }).errors, ["work/x.psd: listed but not on disk"]);
});

test("manifest format", () => {
  const good = entry("export/static.png", "static", 1);
  assert.deepEqual(manifestErrors({ files: [good] }), []);
  assert.deepEqual(manifestErrors({ files: [good], repository: "x" }), ["manifest must be an object with only a files array"]);
  assert.match(manifestErrors({ files: [{ ...good, sha256: "ABC" }] })[0], /64 lowercase hex/);
  assert.match(manifestErrors({ files: [{ ...good, changeset: 0 }] })[0], /whole number/);
  assert.match(manifestErrors({ files: [{ path: good.path, sha256: good.sha256 }] })[0], /exactly path, sha256, changeset/);
  assert.match(manifestErrors({ files: [{ ...good, size: 1 }] })[0], /exactly/);
  assert.match(manifestErrors({ files: [{ ...good, path: "provenance.md" }] })[0], /not a source art file/);
  assert.match(manifestErrors({ files: [{ ...good, path: "../x.png" }] })[0], /\.\. segment/);
  assert.match(manifestErrors({ files: [{ ...good, path: "/x.png" }] })[0], /relative/);
  assert.match(manifestErrors({ files: [{ ...good, path: "work/x.psd" }, good] })[0], /sorted/);
  assert.match(manifestErrors({ files: [good, good] })[0], /sorted/);
});

test("an unreadable manifest is reported", () => {
  const item = avatar();
  writeFileSync(join(item, manifestName), "{");
  assert.match(verifyItem(item).errors[0], /^files\.json: /);
  assert.match(writeItem(item, { changeset: 1 }).errors[0], /^files\.json: /);
});

test("items are the folders with a provenance.md", () => {
  const assets = tree({ "cosmetic/avatar/x/provenance.md": "", "medal/m/provenance.md": "", "README.md": "" });
  assert.deepEqual(findItems(assets), [join(assets, "cosmetic/avatar/x"), join(assets, "medal/m")]);
  assert.deepEqual(findItems(join(assets, "medal/m")), [join(assets, "medal/m")]);
});

test("listed paths join the item folder of each manifest", () => {
  const item = avatar();
  writeItem(item, { changeset: 5 });
  const assets = join(item, "../../..");
  assert.deepEqual(listedPaths(assets, ["README.md", "cosmetic/avatar/x/files.json"]), ["cosmetic/avatar/x/export/static.png", "cosmetic/avatar/x/work/x.psd"]);
});

test("pull fetches each listed file at its changeset and checks its sha256", () => {
  const item = avatar();
  writeItem(item, { changeset: 5 });
  writeFileSync(join(item, "export/static.png"), pointer("static"));
  unlinkSync(join(item, "work/x.psd"));
  const calls = [];
  const content = { "/cosmetic/avatar/x/export/static.png": "static", "/cosmetic/avatar/x/work/x.psd": "layers" };
  const run = (args) => {
    calls.push(args);
    const path = args[1].slice("serverpath:".length, args[1].indexOf("#"));
    writeFileSync(args[2].slice("--file=".length), content[path]);
  };
  assert.deepEqual(pullItem(item, { server: "org@unity", run }), { errors: [], pulled: 2 });
  assert.deepEqual(calls[0], ["cat", "serverpath:/cosmetic/avatar/x/export/static.png#cs:5@hypedomino-art@org@unity", `--file=${join(item, "export/static.png")}`]);
  assert.equal(readFileSync(join(item, "work/x.psd"), "utf8"), "layers");
  assert.deepEqual(pullItem(item, { server: "org@unity", run }), { errors: [], pulled: 0 });
});

test("pull limits to a prefix and reports content that does not match", () => {
  const item = avatar();
  writeItem(item, { changeset: 5 });
  unlinkSync(join(item, "export/static.png"));
  unlinkSync(join(item, "work/x.psd"));
  const run = (args) => writeFileSync(args[2].slice("--file=".length), "other");
  assert.deepEqual(pullItem(item, { server: "org@unity", only: "export/", run }), { errors: ["export/static.png: changeset 5 does not have the sha256 of files.json"], pulled: 1 });
  assert.equal(readManifest(join(item, "missing")), undefined);
  assert.equal(revisionSpec(item, { path: "work/x.psd", changeset: 9 }, "org@unity"), "serverpath:/cosmetic/avatar/x/work/x.psd#cs:9@hypedomino-art@org@unity");
});

test("arguments", () => {
  assert.deepEqual(parseArguments(["write", "--changeset", "12", "--prune", "assets-src"]), { command: "write", paths: ["assets-src"], prune: true, complete: false, changeset: 12 });
  assert.equal(parseArguments(["verify", "--complete", "a"]).complete, true);
  assert.throws(() => parseArguments(["write", "--changeset", "0"]), /whole number/);
  assert.throws(() => parseArguments(["write", "--force"]), /unknown option/);
});
