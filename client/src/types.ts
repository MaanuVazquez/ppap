export type PenEventType = "down" | "move" | "up";

export interface PenEventMessage {
  type: PenEventType;
  x: number;
  y: number;
  pressure: number;
  pointerId: number;
  tiltX: number;
  tiltY: number;
  twist: number;
}

export interface HealthResponse {
  penInjection: boolean;
  activePenBackend: PenBackendId | "";
  screenCapture: boolean;
}

export type PenBackendId = "windowsInk" | "winTab";

export interface PenBackendInfo {
  id: PenBackendId;
  label: string;
  available: boolean;
  reason?: string;
}

export interface PenBackendState {
  active: PenBackendId;
  backends: PenBackendInfo[];
}

export interface StagePoint {
  x: number;
  y: number;
}

export interface ScreenRect {
  x: number;
  y: number;
  width: number;
  height: number;
}
