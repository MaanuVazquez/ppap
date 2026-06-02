import "./style.css";

import { ScreenPreview } from "./screen";
import { PenTransport } from "./transport";
import type { HealthResponse, PenEventMessage, PenEventType, ScreenRect, StagePoint } from "./types";

const stage = getElementById<HTMLDivElement>("stage");
const overlay = getElementById<HTMLCanvasElement>("overlay");
const screen = getElementById<HTMLImageElement>("screen");
const connectionStatus = getElementById<HTMLSpanElement>("connectionStatus");
const pointerStatus = getElementById<HTMLSpanElement>("pointerStatus");
const pressureStatus = getElementById<HTMLSpanElement>("pressureStatus");

const context = getCanvasContext(overlay);

let activePointerId: number | null = null;
let previousPoint: StagePoint | null = null;

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

window.addEventListener("resize", resizeOverlay);
window.addEventListener("orientationchange", resizeOverlay);

stage.addEventListener("pointerdown", (event) => {
  if (activePointerId !== null) {
    return;
  }

  event.preventDefault();
  activePointerId = event.pointerId;
  stage.setPointerCapture(event.pointerId);

  const point = pointFromPointerEvent(event);
  previousPoint = point;
  sendPenEvent("down", event, point);
  drawPoint(point, pressureFromPointerEvent(event));
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
    drawStroke(point, pressureFromPointerEvent(item));
    previousPoint = point;
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
  previousPoint = null;
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

function drawPoint(point: StagePoint, pressure: number): void {
  context.fillStyle = "rgba(110, 168, 254, 0.7)";
  context.beginPath();
  context.arc(point.x, point.y, Math.max(2, pressure * 12), 0, Math.PI * 2);
  context.fill();
}

function drawStroke(point: StagePoint, pressure: number): void {
  if (previousPoint === null) {
    drawPoint(point, pressure);
    return;
  }

  context.strokeStyle = "rgba(110, 168, 254, 0.7)";
  context.lineCap = "round";
  context.lineJoin = "round";
  context.lineWidth = Math.max(2, pressure * 18);
  context.beginPath();
  context.moveTo(previousPoint.x, previousPoint.y);
  context.lineTo(point.x, point.y);
  context.stroke();
}

function resizeOverlay(): void {
  const scale = window.devicePixelRatio || 1;
  const width = Math.max(1, Math.floor(stage.clientWidth * scale));
  const height = Math.max(1, Math.floor(stage.clientHeight * scale));

  overlay.width = width;
  overlay.height = height;
  overlay.style.width = `${stage.clientWidth}px`;
  overlay.style.height = `${stage.clientHeight}px`;
  context.setTransform(scale, 0, 0, scale, 0, 0);
  context.clearRect(0, 0, stage.clientWidth, stage.clientHeight);
}

async function loadHealth(): Promise<void> {
  try {
    const response = await fetch("/api/health", { cache: "no-store" });
    const health = (await response.json()) as HealthResponse;
    connectionStatus.textContent = `Pen: ${statusLabel(health.penInjection)} | Screen: ${statusLabel(health.screenCapture)}`;
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

function getCanvasContext(canvas: HTMLCanvasElement): CanvasRenderingContext2D {
  const canvasContext = canvas.getContext("2d");
  if (canvasContext === null) {
    throw new Error("2D canvas context is unavailable");
  }

  return canvasContext;
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
