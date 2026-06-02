import type { PenEventMessage } from "./types";

export type ConnectionState = "connecting" | "open" | "closed" | "error";

export interface PenTransportOptions {
  onStateChange: (state: ConnectionState) => void;
  onError: (message: string) => void;
}

export class PenTransport {
  private socket: WebSocket | null = null;
  private readonly options: PenTransportOptions;

  public constructor(options: PenTransportOptions) {
    this.options = options;
  }

  public connect(): void {
    this.close();

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = `${protocol}//${window.location.host}/api/pen`;
    this.options.onStateChange("connecting");

    const socket = new WebSocket(url);
    this.socket = socket;

    socket.addEventListener("open", () => {
      this.options.onStateChange("open");
    });

    socket.addEventListener("close", () => {
      if (this.socket === socket) {
        this.options.onStateChange("closed");
      }
    });

    socket.addEventListener("error", () => {
      this.options.onStateChange("error");
      this.options.onError("WebSocket connection failed");
    });

    socket.addEventListener("message", (event: MessageEvent<string>) => {
      this.options.onError(event.data);
    });
  }

  public send(event: PenEventMessage): void {
    if (this.socket?.readyState !== WebSocket.OPEN) {
      return;
    }

    this.socket.send(JSON.stringify(event));
  }

  public close(): void {
    if (this.socket !== null) {
      this.socket.close();
      this.socket = null;
    }
  }
}
