/// <reference types="svelte" />
/// <reference types="vite/client" />

// Wailsのランタイム型定義（受け取るデータの型は呼び出し側のコールバックから決まる）
interface Window {
    runtime: {
        EventsOn: <T = unknown>(eventName: string, callback: (data: T) => void) => void;
        EventsOff: (eventName: string) => void;
        EventsOnce: <T = unknown>(eventName: string, callback: (data: T) => void) => void;
        EventsEmit: (eventName: string, data?: unknown) => void;
    };
}

// svelte-dnd-action のイベント型
// メソッド記法にして、ハンドラ側で items の型（models.Setting など）を指定できるようにする
declare namespace svelte.JSX {
    interface HTMLAttributes<T> {
        onconsider?(event: CustomEvent<import("svelte-dnd-action").DndEvent<unknown>>): void;
        onfinalize?(event: CustomEvent<import("svelte-dnd-action").DndEvent<unknown>>): void;
    }
}
