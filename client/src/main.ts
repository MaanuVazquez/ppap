import "./style.css";

import { ScreenPreview } from "./screen";
import { PenTransport } from "./transport";
import type {
  HealthResponse,
  PenBackendInfo,
  PenBackendState,
  PenEventMessage,
  PenEventType,
  ScreenRect,
  StagePoint,
} from "./types";

const stage = getElementById<HTMLDivElement>("stage");
const overlay = getElementById<HTMLCanvasElement>("overlay");
const screen = getElementById<HTMLImageElement>("screen");
const connectionStatus = getElementById<HTMLSpanElement>("connectionStatus");
const pointerStatus = getElementById<HTMLSpanElement>("pointerStatus");
const pressureStatus = getElementById<HTMLSpanElement>("pressureStatus");
const backendSwitchButton = getElementById<HTMLButtonElement>("backendSwitchButton");
const pressureTestButton = getElementById<HTMLButtonElement>("pressureTestButton");

let activePointerId: number | null = null;
let backendState: PenBackendState | null = null;

const transport = new PenTransport({
  onStateChange: (state) => {
    connectionStatus.textContent = `Socket: ${state}`;
  },
  onError: (message) => {
    connectionStatus.textContent = message;
  },
});

const preview = new ScreenPreview({
  image: screen,
  intervalMs: 180,
  onError: (message) => {
    connectionStatus.textContent = message;
  },
});

transport.connect();
preview.start();
resizeOverlay();
void loadHealth();
void loadPenBackends();

window.addEventListener("resize", resizeOverlay);
window.addEventListener("orientationchange", resizeOverlay);

pressureTestButton.addEventListener("click", () => {
  void runPressureTest();
});

backendSwitchButton.addEventListener("click", () => {
  void switchPenBackend();
});

stage.addEventListener("pointerdown", (event) => {
  if (activePointerId !== null) {
    return;
  }

  event.preventDefault();
  activePointerId = event.pointerId;
  stage.setPointerCapture(event.pointerId);

  const point = pointFromPointerEvent(event);
  sendPenEvent("down", event, point);
});

stage.addEventListener("pointermove", (event) => {
  if (event.pointerId !== activePointerId) {
    return;
  }

  event.preventDefault();
  const coalescedEvents = typeof event.getCoalescedEvents === "function" ? event.getCoalescedEvents() : [];
  const events = coalescedEvents.length > 0 ? coalescedEvents : [event];

  for (const item of events) {
    const point = pointFromPointerEvent(item);
    sendPenEvent("move", item, point);
  }
});

stage.addEventListener("pointerup", handlePointerEnd);
stage.addEventListener("pointercancel", handlePointerEnd);

function handlePointerEnd(event: PointerEvent): void {
  if (event.pointerId !== activePointerId) {
    return;
  }

  event.preventDefault();
  const point = pointFromPointerEvent(event);
  sendPenEvent("up", event, point);
  stage.releasePointerCapture(event.pointerId);
  activePointerId = null;
}

function sendPenEvent(type: PenEventType, event: PointerEvent, point: StagePoint): void {
  const pressure = type === "up" ? 0 : pressureFromPointerEvent(event);
  const normalizedPoint = normalizedScreenPoint(point);
  const message: PenEventMessage = {
    type,
    x: normalizedPoint.x,
    y: normalizedPoint.y,
    pressure,
    pointerId: event.pointerId,
    tiltX: event.tiltX,
    tiltY: event.tiltY,
    twist: event.twist,
  };

  pointerStatus.textContent = `Pointer: ${event.pointerType}`;
  pressureStatus.textContent = `Pressure: ${pressure.toFixed(2)}`;
  transport.send(message);
}

function normalizedScreenPoint(point: StagePoint): StagePoint {
  const screenRect = visibleScreenRect();
  return {
    x: clamp((point.x - screenRect.x) / screenRect.width, 0, 1),
    y: clamp((point.y - screenRect.y) / screenRect.height, 0, 1),
  };
}

function visibleScreenRect(): ScreenRect {
  const stageWidth = Math.max(1, stage.clientWidth);
  const stageHeight = Math.max(1, stage.clientHeight);
  const naturalWidth = screen.naturalWidth || stageWidth;
  const naturalHeight = screen.naturalHeight || stageHeight;
  const stageRatio = stageWidth / stageHeight;
  const screenRatio = naturalWidth / naturalHeight;

  if (stageRatio > screenRatio) {
    const height = stageHeight;
    const width = height * screenRatio;
    return {
      x: (stageWidth - width) / 2,
      y: 0,
      width,
      height,
    };
  }

  const width = stageWidth;
  const height = width / screenRatio;
  return {
    x: 0,
    y: (stageHeight - height) / 2,
    width,
    height,
  };
}

function pointFromPointerEvent(event: PointerEvent): StagePoint {
  const rect = stage.getBoundingClientRect();
  return {
    x: clamp(event.clientX - rect.left, 0, rect.width),
    y: clamp(event.clientY - rect.top, 0, rect.height),
  };
}

function pressureFromPointerEvent(event: PointerEvent): number {
  if (event.pointerType === "pen") {
    return clamp(event.pressure, 0, 1);
  }

  return event.buttons === 0 ? 0 : 0.5;
}

function resizeOverlay(): void {
  const scale = window.devicePixelRatio || 1;
  const width = Math.max(1, Math.floor(stage.clientWidth * scale));
  const height = Math.max(1, Math.floor(stage.clientHeight * scale));

  overlay.width = width;
  overlay.height = height;
  overlay.style.width = `${stage.clientWidth}px`;
  overlay.style.height = `${stage.clientHeight}px`;
}

async function runPressureTest(): Promise<void> {
  pressureTestButton.disabled = true;
  connectionStatus.textContent = "Running pressure test";

  try {
    const response = await fetch("/api/pen/test-stroke", {
      method: "POST",
      cache: "no-store",
    });
    if (!response.ok) {
      throw new Error(`pressure test failed: ${response.status}`);
    }

    connectionStatus.textContent = "Pressure test sent";
  } catch (error) {
    connectionStatus.textContent = error instanceof Error ? error.message : "Pressure test failed";
  } finally {
    pressureTestButton.disabled = false;
  }
}

async function loadPenBackends(): Promise<void> {
  try {
    const response = await fetch("/api/pen/backend", { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`backend status failed: ${response.status}`);
    }

    backendState = (await response.json()) as PenBackendState;
    renderBackendButton();
  } catch (error) {
    backendSwitchButton.textContent = "Backend: unavailable";
    backendSwitchButton.disabled = true;
    connectionStatus.textContent = error instanceof Error ? error.message : "Backend status failed";
  }
}

async function switchPenBackend(): Promise<void> {
  if (backendState === null) {
    return;
  }

  const nextBackend = nextPenBackend(backendState);
  if (nextBackend === null) {
    connectionStatus.textContent = "No other pen backend available";
    return;
  }

  backendSwitchButton.disabled = true;
  try {
    const response = await fetch("/api/pen/backend", {
      method: "POST",
      cache: "no-store",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ backend: nextBackend.id }),
    });

    if (!response.ok) {
      const message = await response.text();
      throw new Error(message.trim() || `backend switch failed: ${response.status}`);
    }

    backendState = (await response.json()) as PenBackendState;
    connectionStatus.textContent = `Backend switched to ${activeBackendLabel(backendState)}`;
  } catch (error) {
    connectionStatus.textContent = error instanceof Error ? error.message : "Backend switch failed";
  } finally {
    renderBackendButton();
  }
}

function nextPenBackend(state: PenBackendState): PenBackendInfo | null {
  if (state.backends.length === 0) {
    return null;
  }

  const activeIndex = Math.max(
    0,
    state.backends.findIndex((backend) => backend.id === state.active),
  );
  return state.backends[(activeIndex + 1) % state.backends.length];
}

function renderBackendButton(): void {
  if (backendState === null) {
    backendSwitchButton.textContent = "Backend: unavailable";
    backendSwitchButton.disabled = true;
    return;
  }

  backendSwitchButton.textContent = `Backend: ${activeBackendLabel(backendState)}`;
  backendSwitchButton.disabled = backendState.backends.length < 2;
}

function activeBackendLabel(state: PenBackendState): string {
  const active = state.backends.find((backend) => backend.id === state.active);
  return active?.label ?? state.active;
}

async function loadHealth(): Promise<void> {
  try {
    const response = await fetch("/api/health", { cache: "no-store" });
    const health = (await response.json()) as HealthResponse;
    connectionStatus.textContent = `Pen: ${statusLabel(health.penInjection)} | Backend: ${health.activePenBackend || "none"} | Screen: ${statusLabel(health.screenCapture)}`;
  } catch {
    connectionStatus.textContent = "Health check unavailable";
  }
}

function statusLabel(value: boolean): string {
  return value ? "ready" : "unavailable";
}

function getElementById<TElement extends HTMLElement>(id: string): TElement {
  const element = document.getElementById(id);
  if (element === null) {
    throw new Error(`Missing element #${id}`);
  }

  return element as TElement;
}

function clamp(value: number, min: number, max: number): number {
  if (value < min) {
    return min;
  }
  if (value > max) {
    return max;
  }

  return value;
}
