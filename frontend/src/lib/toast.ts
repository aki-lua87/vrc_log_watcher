import { writable } from "svelte/store";

// 画面下に一瞬だけ出す短いお知らせ
export const toastMessage = writable("");

let timer: number | undefined;

export function toast(message: string) {
    toastMessage.set(message);
    clearTimeout(timer);
    timer = window.setTimeout(() => toastMessage.set(""), 1800);
}
