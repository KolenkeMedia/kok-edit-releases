# Как выпускать версии kök Edit

1. Собрать папку `kok-edit` новой версии (версия — в `CSXS/manifest.xml`).
2. Тестовая среда:
   ```
   EDIT_SIGN_KEY=… node tools/release/build-release.js путь/к/kok-edit test "что нового" "ещё пункт"
   git add -A && git commit -m "kök Edit X.Y.Z — тест" && git push
   ```
   У кого стоит тестовый установщик, через ≤6 часов (или по «Проверить обновления») появится «Обновить».
3. Проверили — отдать клиентам:
   ```
   node tools/release/promote.js X.Y.Z
   git add -A && git commit -m "kök Edit X.Y.Z — клиентам" && git push
   ```
4. Установщики: `tools/installer/build.sh` → `installer/kok-edit-setup.exe` и `-test.exe`. Пересобирать нужно только если менялся сам установщик.

`EDIT_SIGN_KEY` — тот же секретный ключ, что в Railway. В репозиторий не класть.
