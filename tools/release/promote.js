// Перевести проверенную тестовую версию клиентам: channels/main.json = channels/test.json
// node promote.js [версия]  — без версии берётся текущая тестовая
const fs = require("fs"), path = require("path");
const REPO_DIR = process.env.RELEASES_DIR || path.join(__dirname, "..", "..");
const test = JSON.parse(fs.readFileSync(path.join(REPO_DIR, "channels", "test.json"), "utf8"));
const want = process.argv[2];
if (want && want !== test.version) { console.error(`В тесте сейчас ${test.version}, а не ${want}`); process.exit(1); }
fs.writeFileSync(path.join(REPO_DIR, "channels", "main.json"), JSON.stringify(Object.assign({}, test, { promotedAt: new Date().toISOString() }), null, 2));
// новый клиентский установщик ждёт в tools/installer/next — переносим вместе с версией
const next = path.join(REPO_DIR, "tools", "installer", "next", "kok-edit-setup.exe");
if (fs.existsSync(next)) { fs.copyFileSync(next, path.join(REPO_DIR, "installer", "kok-edit-setup.exe")); fs.unlinkSync(next); console.log("Клиентский установщик обновлён."); }
console.log(`Клиентам: ${test.version}. Осталось git commit и push.`);
