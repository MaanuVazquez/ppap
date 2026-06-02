#include "wintab_proxy.h"

#include <string.h>

static HANDLE g_mapping = NULL;
static PPAPPacket *g_packet = NULL;
static uint32_t g_last_serial = 0;
static int g_context_token = 1;

static void ppap_open_mapping(void) {
  if (g_packet != NULL) {
    return;
  }

  g_mapping = OpenFileMappingW(FILE_MAP_READ, FALSE, PPAP_WINTAB_MAPPING_NAME);
  if (g_mapping == NULL) {
    return;
  }

  g_packet = (PPAPPacket *)MapViewOfFile(g_mapping, FILE_MAP_READ, 0, 0, sizeof(PPAPPacket));
  if (g_packet == NULL) {
    CloseHandle(g_mapping);
    g_mapping = NULL;
  }
}

static void ppap_close_mapping(void) {
  if (g_packet != NULL) {
    UnmapViewOfFile(g_packet);
    g_packet = NULL;
  }
  if (g_mapping != NULL) {
    CloseHandle(g_mapping);
    g_mapping = NULL;
  }
}

static UINT copy_string_a(LPVOID output, const char *value) {
  if (output != NULL) {
    strcpy((char *)output, value);
  }
  return (UINT)strlen(value) + 1;
}

static UINT copy_string_w(LPVOID output, const wchar_t *value) {
  size_t length = wcslen(value) + 1;
  if (output != NULL) {
    memcpy(output, value, length * sizeof(wchar_t));
  }
  return (UINT)(length * sizeof(wchar_t));
}

static void fill_context_a(LOGCONTEXTA *context) {
  memset(context, 0, sizeof(*context));
  strcpy(context->lcName, "PPAP WinTab Context");
  context->lcPktRate = 240;
  context->lcPktData = PPAP_PK_X | PPAP_PK_Y | PPAP_PK_BUTTONS | PPAP_PK_NORMAL_PRESSURE;
  context->lcMoveMask = context->lcPktData;
  context->lcBtnDnMask = 1;
  context->lcBtnUpMask = 1;
  context->lcInOrgX = 0;
  context->lcInOrgY = 0;
  context->lcInExtX = 65535;
  context->lcInExtY = 65535;
  context->lcOutOrgX = 0;
  context->lcOutOrgY = 0;
  context->lcOutExtX = 65535;
  context->lcOutExtY = 65535;
  context->lcSysMode = TRUE;
  context->lcSysExtX = GetSystemMetrics(SM_CXVIRTUALSCREEN);
  context->lcSysExtY = GetSystemMetrics(SM_CYVIRTUALSCREEN);
}

static void fill_context_w(LOGCONTEXTW *context) {
  memset(context, 0, sizeof(*context));
  wcscpy(context->lcName, L"PPAP WinTab Context");
  context->lcPktRate = 240;
  context->lcPktData = PPAP_PK_X | PPAP_PK_Y | PPAP_PK_BUTTONS | PPAP_PK_NORMAL_PRESSURE;
  context->lcMoveMask = context->lcPktData;
  context->lcBtnDnMask = 1;
  context->lcBtnUpMask = 1;
  context->lcInOrgX = 0;
  context->lcInOrgY = 0;
  context->lcInExtX = 65535;
  context->lcInExtY = 65535;
  context->lcOutOrgX = 0;
  context->lcOutOrgY = 0;
  context->lcOutExtX = 65535;
  context->lcOutExtY = 65535;
  context->lcSysMode = TRUE;
  context->lcSysExtX = GetSystemMetrics(SM_CXVIRTUALSCREEN);
  context->lcSysExtY = GetSystemMetrics(SM_CYVIRTUALSCREEN);
}

static BOOL read_latest_packet(PPAPWinTabPacket *out, BOOL consume) {
  ppap_open_mapping();
  if (g_packet == NULL || g_packet->magic != PPAP_WINTAB_MAGIC || g_packet->version != PPAP_WINTAB_VERSION) {
    return FALSE;
  }
  if (consume && g_packet->serial == g_last_serial) {
    return FALSE;
  }

  out->x = g_packet->x;
  out->y = g_packet->y;
  out->buttons = g_packet->in_contact != 0 ? 1u : 0u;
  out->normal_pressure = g_packet->pressure;

  if (consume) {
    g_last_serial = g_packet->serial;
  }
  return TRUE;
}

BOOL WINAPI DllMain(HINSTANCE instance, DWORD reason, LPVOID reserved) {
  (void)instance;
  (void)reserved;
  if (reason == DLL_PROCESS_DETACH) {
    ppap_close_mapping();
  }
  return TRUE;
}

UINT APIENTRY WTInfoA(UINT category, UINT index, LPVOID output) {
  if (category == PPAP_WTI_INTERFACE && index == 0) {
    return copy_string_a(output, "PPAP WinTab Proxy");
  }
  if ((category == PPAP_WTI_DEFCONTEXT || category == PPAP_WTI_DEFSYSCTX) && index == 0) {
    if (output != NULL) {
      fill_context_a((LOGCONTEXTA *)output);
    }
    return sizeof(LOGCONTEXTA);
  }
  if (category == PPAP_WTI_DEVICES && index == PPAP_DVC_NAME) {
    return copy_string_a(output, "PPAP Apple Pencil");
  }
  if (category == PPAP_WTI_DEVICES && index == PPAP_DVC_NPRESSURE) {
    if (output != NULL) {
      AXIS *axis = (AXIS *)output;
      axis->axMin = 0;
      axis->axMax = 1024;
      axis->axUnits = 0;
      axis->axResolution = 1024;
    }
    return sizeof(AXIS);
  }
  if (category == PPAP_WTI_STATUS && output != NULL) {
    *(UINT *)output = 0;
    return sizeof(UINT);
  }
  return 0;
}

UINT APIENTRY WTInfoW(UINT category, UINT index, LPVOID output) {
  if (category == PPAP_WTI_INTERFACE && index == 0) {
    return copy_string_w(output, L"PPAP WinTab Proxy");
  }
  if ((category == PPAP_WTI_DEFCONTEXT || category == PPAP_WTI_DEFSYSCTX) && index == 0) {
    if (output != NULL) {
      fill_context_w((LOGCONTEXTW *)output);
    }
    return sizeof(LOGCONTEXTW);
  }
  if (category == PPAP_WTI_DEVICES && index == PPAP_DVC_NAME) {
    return copy_string_w(output, L"PPAP Apple Pencil");
  }
  return WTInfoA(category, index, output);
}

HCTX APIENTRY WTOpenA(HWND window, LOGCONTEXTA *context, BOOL enable) {
  (void)window;
  (void)context;
  (void)enable;
  ppap_open_mapping();
  return (HCTX)(uintptr_t)InterlockedIncrement((LONG *)&g_context_token);
}

HCTX APIENTRY WTOpenW(HWND window, LOGCONTEXTW *context, BOOL enable) {
  (void)window;
  (void)context;
  (void)enable;
  ppap_open_mapping();
  return (HCTX)(uintptr_t)InterlockedIncrement((LONG *)&g_context_token);
}

BOOL APIENTRY WTClose(HCTX context) {
  (void)context;
  return TRUE;
}

BOOL APIENTRY WTEnable(HCTX context, BOOL enable) {
  (void)context;
  (void)enable;
  return TRUE;
}

BOOL APIENTRY WTOverlap(HCTX context, BOOL to_top) {
  (void)context;
  (void)to_top;
  return TRUE;
}

int APIENTRY WTPacketsGet(HCTX context, int max_packets, LPVOID packets) {
  (void)context;
  if (max_packets <= 0 || packets == NULL) {
    return 0;
  }
  return read_latest_packet((PPAPWinTabPacket *)packets, TRUE) ? 1 : 0;
}

int APIENTRY WTPacketsPeek(HCTX context, int max_packets, LPVOID packets) {
  (void)context;
  if (max_packets <= 0 || packets == NULL) {
    return 0;
  }
  return read_latest_packet((PPAPWinTabPacket *)packets, FALSE) ? 1 : 0;
}

BOOL APIENTRY WTPacket(HCTX context, UINT serial, LPVOID packet) {
  (void)context;
  (void)serial;
  if (packet == NULL) {
    return FALSE;
  }
  return read_latest_packet((PPAPWinTabPacket *)packet, FALSE);
}

BOOL APIENTRY WTQueuePacketsEx(HCTX context, UINT *oldest, UINT *newest) {
  (void)context;
  ppap_open_mapping();
  if (g_packet == NULL || g_packet->magic != PPAP_WINTAB_MAGIC) {
    return FALSE;
  }
  if (oldest != NULL) {
    *oldest = g_packet->serial;
  }
  if (newest != NULL) {
    *newest = g_packet->serial;
  }
  return TRUE;
}

int APIENTRY WTQueueSizeGet(HCTX context) {
  (void)context;
  return 1;
}

BOOL APIENTRY WTQueueSizeSet(HCTX context, int size) {
  (void)context;
  (void)size;
  return TRUE;
}

BOOL APIENTRY WTGetA(HCTX context, LOGCONTEXTA *context_out) {
  (void)context;
  if (context_out != NULL) {
    fill_context_a(context_out);
  }
  return TRUE;
}

BOOL APIENTRY WTGetW(HCTX context, LOGCONTEXTW *context_out) {
  (void)context;
  if (context_out != NULL) {
    fill_context_w(context_out);
  }
  return TRUE;
}

BOOL APIENTRY WTSetA(HCTX context, LOGCONTEXTA *context_in) {
  (void)context;
  (void)context_in;
  return TRUE;
}

BOOL APIENTRY WTSetW(HCTX context, LOGCONTEXTW *context_in) {
  (void)context;
  (void)context_in;
  return TRUE;
}

BOOL APIENTRY WTExtGet(HCTX context, UINT extension, LPVOID data) {
  (void)context;
  (void)extension;
  (void)data;
  return FALSE;
}

BOOL APIENTRY WTExtSet(HCTX context, UINT extension, LPVOID data) {
  (void)context;
  (void)extension;
  (void)data;
  return FALSE;
}

HMGR APIENTRY WTMgrOpen(HWND window, UINT message) {
  (void)window;
  (void)message;
  return (HMGR)1;
}

BOOL APIENTRY WTMgrClose(HMGR manager) {
  (void)manager;
  return TRUE;
}

BOOL APIENTRY WTMgrContextEnum(HMGR manager, WTENUMPROC callback, LPARAM lparam) {
  (void)manager;
  (void)callback;
  (void)lparam;
  return TRUE;
}

HWND APIENTRY WTMgrContextOwner(HMGR manager, HCTX context) {
  (void)manager;
  (void)context;
  return NULL;
}

HCTX APIENTRY WTMgrDefContext(HMGR manager, BOOL system_context) {
  (void)manager;
  (void)system_context;
  return (HCTX)1;
}

HCTX APIENTRY WTMgrDefContextEx(HMGR manager, UINT device, BOOL system_context) {
  (void)manager;
  (void)device;
  (void)system_context;
  return (HCTX)1;
}

BOOL APIENTRY WTMgrDeviceConfig(HMGR manager, UINT device, HWND owner) {
  (void)manager;
  (void)device;
  (void)owner;
  return TRUE;
}
