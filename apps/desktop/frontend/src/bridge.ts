export type Document = { name: string; path: string; size: number };
export type SaveResult = { cancelled: boolean; path?: string };
type Desktop = {
  SelectFile(): Promise<Document | null>;
  UseDroppedFile(paths: string[]): Promise<Document>;
  ClearFile(): Promise<void>;
  RemovePassword(password: string): Promise<SaveResult>;
};
declare global {
  interface Window {
    go?: { main: { Desktop: Desktop } };
    runtime?: {
      OnFileDrop(callback: (x: number, y: number, paths: string[]) => void, useDropTarget: boolean): void;
      OnFileDropOff(): void;
    };
  }
}
export const isDesktop = () => Boolean(window.go?.main.Desktop);
export function desktop(): Desktop {
  const app = window.go?.main.Desktop;
  if (!app) throw new Error('Open the desktop app to select and process PDFs. This browser preview only shows the interface.');
  return app;
}
export function onFileDrop(callback: (paths: string[]) => void): () => void {
  window.runtime?.OnFileDrop((_x, _y, paths) => callback(paths), true);
  return () => window.runtime?.OnFileDropOff();
}
