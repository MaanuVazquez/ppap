export interface ScreenPreviewOptions {
  image: HTMLImageElement;
  intervalMs: number;
  onError: (message: string) => void;
}

export class ScreenPreview {
  private readonly image: HTMLImageElement;
  private readonly intervalMs: number;
  private readonly onError: (message: string) => void;
  private timer: number | null = null;
  private objectUrl: string | null = null;

  public constructor(options: ScreenPreviewOptions) {
    this.image = options.image;
    this.intervalMs = options.intervalMs;
    this.onError = options.onError;
  }

  public start(): void {
    this.stop();
    void this.refresh();
    this.timer = window.setInterval(() => {
      void this.refresh();
    }, this.intervalMs);
  }

  public stop(): void {
    if (this.timer !== null) {
      window.clearInterval(this.timer);
      this.timer = null;
    }
    if (this.objectUrl !== null) {
      URL.revokeObjectURL(this.objectUrl);
      this.objectUrl = null;
    }
  }

  private async refresh(): Promise<void> {
    try {
      const response = await fetch(`/api/screen.jpg?t=${Date.now()}`, { cache: "no-store" });
      if (!response.ok) {
        throw new Error(`screen preview failed: ${response.status}`);
      }

      const blob = await response.blob();
      const nextUrl = URL.createObjectURL(blob);
      const previousUrl = this.objectUrl;
      this.objectUrl = nextUrl;
      this.image.src = nextUrl;

      if (previousUrl !== null) {
        URL.revokeObjectURL(previousUrl);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "screen preview failed";
      this.onError(message);
    }
  }
}
