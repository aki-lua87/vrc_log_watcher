export type ThemeId = "light" | "dark" | "red";

// preview: [サイドバー, 背景, アクセント, 文字]
export const THEMES: { id: ThemeId; name: string; preview: string[] }[] = [
    { id: "light", name: "ライト", preview: ["#ffffff", "#f4f7f6", "#0d9488", "#15201e"] },
    { id: "dark", name: "ダーク", preview: ["#161a1b", "#0f1213", "#2dd4bf", "#e6ebea"] },
    { id: "red", name: "レッド", preview: ["#111113", "#09090a", "#e8112d", "#f2f2f3"] },
];

export function normalizeTheme(theme: string | undefined): ThemeId {
    return THEMES.some((t) => t.id === theme) ? (theme as ThemeId) : "light";
}

export function applyTheme(theme: ThemeId) {
    document.documentElement.dataset.theme = theme;
}
