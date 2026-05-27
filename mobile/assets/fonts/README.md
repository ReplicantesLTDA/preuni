# Fonts

The app loads hand-drawn fonts at the root layout via `expo-font.useFonts`. Drop the following TTF files in this directory (download from Google Fonts):

| File | Family token (`tokens.font.*`) |
|------|-------------------------------|
| `CaveatBrush-Regular.ttf` | `display` |
| `PatrickHand-Regular.ttf` | `heading` |
| `ArchitectsDaughter-Regular.ttf` | `body` |
| `Kalam-Regular.ttf` | `numeric` |
| `Kalam-Bold.ttf` | `numeric` (bold) |
| `Caveat-Regular.ttf` | `accent` |
| `Caveat-Bold.ttf` | `accent` (bold) |

> The placeholder bytes in `*.ttf.placeholder` files exist so the build does not break before assets are wired. Replace with real font binaries before first device build.

Sources (Apache / OFL licensed):

- Caveat Brush: https://fonts.google.com/specimen/Caveat+Brush
- Patrick Hand: https://fonts.google.com/specimen/Patrick+Hand
- Architects Daughter: https://fonts.google.com/specimen/Architects+Daughter
- Kalam: https://fonts.google.com/specimen/Kalam
- Caveat: https://fonts.google.com/specimen/Caveat
