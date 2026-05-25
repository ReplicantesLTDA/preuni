# Quickstart: App Design Refresh — Core Flow

**Feature**: `011-app-design-refresh` | **Date**: 2026-05-25

How to preview and validate the UI/UX refactor for this feature.

---

## 1. Run the Web Preview (primary reference)

```bash
# From repo root
make run-web
```

Open `http://localhost:8088`.

## 2. Core Flow Smoke Test

Verify the end-to-end journey feels cohesive:

- **Welcome**: mascot placeholder is present and looks intentional
- **Login/Register**: layout is warm and consistent with welcome
- **Onboarding / escolher matéria**: track selection remains clear and readable
- **Main navigation**: bottom tabs match the wireframe destinations and the current tab is obvious

## 3. Navigation Audit (wireframe alignment)

From the main app:

- [ ] You can reach **Trilha**, **Redação**, **Amigos**, **Liga**, **Perfil** from the bottom bar
- [ ] Each destination has a clear title and uses the same scaffold
- [ ] **Ajustes** is reachable from **Perfil** (gear action) and has a structured layout

## 4. Placeholder Quality Check

Placeholders are allowed, but must not feel broken:

- [ ] Mascot region uses the defined placeholder component (not a missing image)
- [ ] Amigos/Liga screens have a friendly empty state with clear next steps
- [ ] Any locked/upcoming state explains what exists today vs what comes later

## 5. Accessibility Quick Checks

- [ ] All nav icons have meaningful labels (screen reader / accessibility)
- [ ] Buttons and interactive list items have large tap targets
- [ ] No critical action relies on color alone

## 6. Minimum Viewport Check

In DevTools set width to ~360px:

- [ ] No horizontal scrolling
- [ ] Titles and metrics don’t clip
- [ ] Bottom nav labels remain legible

## 7. Stop

```bash
make stop-web
```