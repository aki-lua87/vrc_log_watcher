// 通知先（Setting.Type）の表示情報。並び順は通知先カードの表示順
export const DESTINATIONS: { type: string; label: string; icon: string; desc: string }[] = [
    { type: "SendXSOverlay", label: "XSOverlay", icon: "monitor", desc: "VRの中に通知を表示します" },
    { type: "SendDiscordWebHook", label: "Discord", icon: "msg", desc: "Webhookでチャンネルに投稿します" },
    { type: "LogOnly", label: "アプリ内に表示", icon: "eye", desc: "このアプリのログにだけ出します" },
    { type: "OutputTextFile", label: "テキストファイル", icon: "file", desc: "1件ずつファイルに書き出します" },
    { type: "WebRequest", label: "Webリクエスト", icon: "send", desc: "指定したURLへ送ります（上級者向け）" },
    { type: "Disable", label: "停止", icon: "ban", desc: "このルールを一時的に使いません" },
];

export function destinationOf(type: string) {
    return DESTINATIONS.find((d) => d.type === type) ?? DESTINATIONS[DESTINATIONS.length - 1];
}

export const TEMPLATE_VARS: { name: string; desc: string; example: string }[] = [
    { name: "{time}", desc: "ログに記録された時刻", example: "2026.10.08 20:41:31" },
    { name: "{title}", desc: "ルール名", example: "ワールド入室 (名前)" },
    { name: "{description}", desc: "ルールの説明", example: "ワールド入室時のワールド名を取得" },
    { name: "{matched_text}", desc: "ログから取り出した値", example: "Midnight Rooftop" },
    { name: "{log_file}", desc: "読み取り中のログファイル名", example: "output_log_2026-10-08_20-41-08.txt" },
];
