/// <reference types="svelte" />
/// <reference types="vite/client" />

// Wailsのランタイム型定義
interface Window {
  runtime: {
    EventsOn: (eventName: string, callback: (data: any) => void) => void;
    EventsOff: (eventName: string) => void;
    EventsOnce: (eventName: string, callback: (data: any) => void) => void;
    EventsEmit: (eventName: string, data?: any) => void;
  };
}
