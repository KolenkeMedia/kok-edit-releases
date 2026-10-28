// Сборка версии kök Edit для GitHub: части core / ffmpeg / whisper, release.json с подписью, указатель канала.
// node build-release.js <папка kok-edit> <test|main> "что нового 1" "что нового 2" ...
// Ключ подписи: переменная EDIT_SIGN_KEY (то же значение, что в Railway).
const fs = require("fs"), path = require("path"), crypto = require("crypto"), { execFileSync } = require("child_process");
const REPO = "KolenkeMedia/kok-edit-releases";
const [src, channel, ...notes] = process.argv.slice(2);
if (!src || !["test", "main"].includes(channel)) { console.error("node build-release.js <папка kok-edit> <test|main> [заметки…]"); process.exit(1); }
const KEY = process.env.EDIT_SIGN_KEY;
if (!KEY) { console.error("Нет EDIT_SIGN_KEY"); process.exit(1); }
const manifest = fs.readFileSync(path.join(src, "CSXS", "manifest.xml"), "utf8");
const version = manifest.match(/ExtensionBundleVersion="([^"]+)"/)[1];
const tag = "v" + version;
// Всё лежит прямо в репозитории (GitHub Releases из этой среды недоступны): files/, versions/, channels/
const REPO_DIR = process.env.RELEASES_DIR || path.join(__dirname, "..", "..");
const RAW = `https://raw.githubusercontent.com/${REPO}/main/`;
const OUT = path.join(__dirname, "out", tag);
const STATE = path.join(__dirname, "state.json"); // ключ части → где уже лежит (чтобы не заливать ffmpeg каждый раз)
const state = fs.existsSync(STATE) ? JSON.parse(fs.readFileSync(STATE, "utf8")) : {};
fs.rmSync(OUT, { recursive: true, force: true });
fs.mkdirSync(OUT, { recursive: true });

const walk = (d, base = d) => fs.readdirSync(d, { withFileTypes: true }).flatMap((e) => (e.isDirectory() ? walk(path.join(d, e.name), base) : [path.relative(base, path.join(d, e.name)).split(path.sep).join("/")]));
const all = walk(src).filter((f) => f !== "components.json");
const PARTS = {
  ffmpeg: all.filter((f) => /^bin\/(ffmpeg\.exe|FFMPEG_LICENSE\.txt)$/.test(f)),
  whisper: all.filter((f) => f.startsWith("bin/whisper/")),
};
PARTS.core = all.filter((f) => !PARTS.ffmpeg.includes(f) && !PARTS.whisper.includes(f));

const comps = {}, upload = [];
for (const [name, files] of Object.entries(PARTS)) {
  if (!files.length) continue;
  const h = crypto.createHash("sha256");
  for (const f of files.sort()) h.update(f + "\0").update(fs.readFileSync(path.join(src, f)));
  const key = h.digest("hex");
  const known = state[name + ":" + key];
  if (known && name !== "core") { comps[name] = Object.assign({ key }, known); console.log(`${name}: не менялся — ${known.url}`); continue; }
  // стабильный zip: одинаковое время у файлов
  const stage = path.join(OUT, "_" + name);
  for (const f of files) { const d = path.join(stage, f); fs.mkdirSync(path.dirname(d), { recursive: true }); fs.copyFileSync(path.join(src, f), d); fs.utimesSync(d, new Date("2026-01-01T00:00:00Z"), new Date("2026-01-01T00:00:00Z")); }
  const file = `${name}-${name === "core" ? version : key.slice(0, 10)}.zip`;
  execFileSync("zip", ["-q", "-X", "-9", "-r", path.join(OUT, file), ...files.sort()], { cwd: stage });
  fs.rmSync(stage, { recursive: true, force: true });
  const buf = fs.readFileSync(path.join(OUT, file));
  fs.mkdirSync(path.join(REPO_DIR, "files"), { recursive: true });
  fs.copyFileSync(path.join(OUT, file), path.join(REPO_DIR, "files", file));
  comps[name] = { key, url: RAW + "files/" + file, sha256: crypto.createHash("sha256").update(buf).digest("hex"), size: buf.length, files: files.length };
  upload.push(file);
  console.log(`${name}: ${file} ${(buf.length / 1048576).toFixed(1)} МБ`);
}
// изменился манифест (размер панели и т.п.) — Premiere нужно перезапустить
const prevManifest = state.lastManifest || "";
const restart = !!prevManifest && prevManifest.replace(/Version="[^"]+"/g, "") !== manifest.replace(/Version="[^"]+"/g, "");
const release = { version, channel, date: new Date().toISOString(), notes, restart, components: comps };
const data = JSON.stringify(release, null, 2);
const priv = crypto.createPrivateKey({ key: Buffer.from(KEY, "base64"), format: "der", type: "pkcs8" });
fs.writeFileSync(path.join(OUT, "release.json"), data);
fs.writeFileSync(path.join(OUT, "release.sig"), crypto.sign(null, Buffer.from(data), priv).toString("base64"));
upload.push("release.json", "release.sig");
const vdir = path.join(REPO_DIR, "versions", tag);
fs.mkdirSync(vdir, { recursive: true });
fs.copyFileSync(path.join(OUT, "release.json"), path.join(vdir, "release.json"));
fs.copyFileSync(path.join(OUT, "release.sig"), path.join(vdir, "release.sig"));
const base = RAW + "versions/" + tag + "/";
const ptr = JSON.stringify({ version, tag, release: base + "release.json", sig: base + "release.sig", date: release.date }, null, 2);
fs.mkdirSync(path.join(REPO_DIR, "channels"), { recursive: true });
fs.writeFileSync(path.join(REPO_DIR, "channels", channel + ".json"), ptr);
fs.writeFileSync(path.join(OUT, "upload.txt"), upload.join("\n"));
// запомнить, где лежат части
for (const [name, c] of Object.entries(comps)) state[name + ":" + c.key] = { url: c.url, sha256: c.sha256, size: c.size };
state.lastManifest = manifest;
fs.writeFileSync(STATE, JSON.stringify(state, null, 2));
console.log(`Готово: версия ${version} → канал ${channel}. Файлы в ${REPO_DIR} — осталось git commit и push.`);
