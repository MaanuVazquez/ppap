export interface ScreenPreviewOptions {
  image: HTMLImageElement;
  intervalMs: number;
  onError: (message: string) => void;
}

export class ScreenPreview {
  private readonly image: HTMLImageElement;
  private readonly intervalMs: number;
  private readonly onError: (message: string) => void;
  private timeoutId: number | null = null;
  private abortController: AbortController | null = null;
  private objectUrl: string | null = null;
  private stopped = true;

  public constructor(options: ScreenPreviewOptions) {
    this.image = options.image;
    this.intervalMs = options.intervalMs;
    this.onError = options.onError;
  }

  public start(): void {
    this.stop();
    this.stopped = false;
    this.scheduleNextRefresh(0);
  }

  public stop(): void {
    this.stopped = true;
    if (this.timeoutId !== null) {
      window.clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
    if (this.abortController !== null) {
      this.abortController.abort();
      this.abortController = null;
    }
    if (this.objectUrl !== null) {
      URL.revokeObjectURL(this.objectUrl);
      this.objectUrl = null;
    }
  }

  private scheduleNextRefresh(delayMs: number): void {
    if (this.stopped) {
      return;
    }

    this.timeoutId = window.setTimeout(() => {
      this.timeoutId = null;
      void this.refresh().finally(() => {
        this.scheduleNextRefresh(this.intervalMs);
      });
    }, delayMs);
  }

  private async refresh(): Promise<void> {
    const abortController = new AbortController();
    this.abortController = abortController;

    try {
      const response = await fetch(`/api/screen.jpg?t=${Date.now()}`, {
        cache: "no-store",
        signal: abortController.signal,
      });
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
      if (isAbortError(error)) {
        return;
      }

      const message = error instanceof Error ? error.message : "screen preview failed";
      this.onError(message);
    } finally {
      if (this.abortController === abortController) {
        this.abortController = null;
      }
    }
  }
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}
