#pragma once

#include <stdint.h>
#include <wchar.h>
#include <windows.h>

#define PPAP_WINTAB_MAPPING_NAME L"Local\\PPAPWinTabPacketV1"
#define PPAP_WINTAB_MAGIC 0x50504157u
#define PPAP_WINTAB_VERSION 1u

#define PPAP_PK_X 0x00000080u
#define PPAP_PK_Y 0x00000100u
#define PPAP_PK_BUTTONS 0x00000040u
#define PPAP_PK_NORMAL_PRESSURE 0x00000400u

#define PPAP_WTI_INTERFACE 1u
#define PPAP_WTI_STATUS 2u
#define PPAP_WTI_DEFCONTEXT 3u
#define PPAP_WTI_DEFSYSCTX 4u
#define PPAP_WTI_DEVICES 100u
#define PPAP_DVC_NAME 1u
#define PPAP_DVC_NPRESSURE 15u

typedef HANDLE HCTX;
typedef HANDLE HMGR;
typedef BOOL(CALLBACK *WTENUMPROC)(HCTX context, LPARAM param);

typedef struct PPAPPacket {
  uint32_t magic;
  uint32_t version;
  uint32_t serial;
  int32_t x;
  int32_t y;
  uint32_t buttons;
  uint32_t pressure;
  uint32_t in_contact;
} PPAPPacket;

typedef struct PPAPWinTabPacket {
  int32_t x;
  int32_t y;
  uint32_t buttons;
  uint32_t normal_pressure;
} PPAPWinTabPacket;

typedef struct AXIS {
  LONG axMin;
  LONG axMax;
  UINT axUnits;
  DWORD axResolution;
} AXIS;

typedef struct LOGCONTEXTA {
  char lcName[40];
  UINT lcOptions;
  UINT lcStatus;
  UINT lcLocks;
  UINT lcMsgBase;
  UINT lcDevice;
  UINT lcPktRate;
  DWORD lcPktData;
  DWORD lcPktMode;
  DWORD lcMoveMask;
  DWORD lcBtnDnMask;
  DWORD lcBtnUpMask;
  LONG lcInOrgX;
  LONG lcInOrgY;
  LONG lcInOrgZ;
  LONG lcInExtX;
  LONG lcInExtY;
  LONG lcInExtZ;
  LONG lcOutOrgX;
  LONG lcOutOrgY;
  LONG lcOutOrgZ;
  LONG lcOutExtX;
  LONG lcOutExtY;
  LONG lcOutExtZ;
  DWORD lcSensX;
  DWORD lcSensY;
  DWORD lcSensZ;
  BOOL lcSysMode;
  int lcSysOrgX;
  int lcSysOrgY;
  int lcSysExtX;
  int lcSysExtY;
  DWORD lcSysSensX;
  DWORD lcSysSensY;
} LOGCONTEXTA;

typedef struct LOGCONTEXTW {
  wchar_t lcName[40];
  UINT lcOptions;
  UINT lcStatus;
  UINT lcLocks;
  UINT lcMsgBase;
  UINT lcDevice;
  UINT lcPktRate;
  DWORD lcPktData;
  DWORD lcPktMode;
  DWORD lcMoveMask;
  DWORD lcBtnDnMask;
  DWORD lcBtnUpMask;
  LONG lcInOrgX;
  LONG lcInOrgY;
  LONG lcInOrgZ;
  LONG lcInExtX;
  LONG lcInExtY;
  LONG lcInExtZ;
  LONG lcOutOrgX;
  LONG lcOutOrgY;
  LONG lcOutOrgZ;
  LONG lcOutExtX;
  LONG lcOutExtY;
  LONG lcOutExtZ;
  DWORD lcSensX;
  DWORD lcSensY;
  DWORD lcSensZ;
  BOOL lcSysMode;
  int lcSysOrgX;
  int lcSysOrgY;
  int lcSysExtX;
  int lcSysExtY;
  DWORD lcSysSensX;
  DWORD lcSysSensY;
} LOGCONTEXTW;

BOOL WINAPI DllMain(HINSTANCE instance, DWORD reason, LPVOID reserved);
UINT APIENTRY WTInfoA(UINT category, UINT index, LPVOID output);
UINT APIENTRY WTInfoW(UINT category, UINT index, LPVOID output);
HCTX APIENTRY WTOpenA(HWND window, LOGCONTEXTA *context, BOOL enable);
HCTX APIENTRY WTOpenW(HWND window, LOGCONTEXTW *context, BOOL enable);
BOOL APIENTRY WTClose(HCTX context);
BOOL APIENTRY WTEnable(HCTX context, BOOL enable);
BOOL APIENTRY WTOverlap(HCTX context, BOOL to_top);
int APIENTRY WTPacketsGet(HCTX context, int max_packets, LPVOID packets);
int APIENTRY WTPacketsPeek(HCTX context, int max_packets, LPVOID packets);
BOOL APIENTRY WTPacket(HCTX context, UINT serial, LPVOID packet);
BOOL APIENTRY WTQueuePacketsEx(HCTX context, UINT *oldest, UINT *newest);
int APIENTRY WTQueueSizeGet(HCTX context);
BOOL APIENTRY WTQueueSizeSet(HCTX context, int size);
BOOL APIENTRY WTGetA(HCTX context, LOGCONTEXTA *context_out);
BOOL APIENTRY WTGetW(HCTX context, LOGCONTEXTW *context_out);
BOOL APIENTRY WTSetA(HCTX context, LOGCONTEXTA *context_in);
BOOL APIENTRY WTSetW(HCTX context, LOGCONTEXTW *context_in);
BOOL APIENTRY WTExtGet(HCTX context, UINT extension, LPVOID data);
BOOL APIENTRY WTExtSet(HCTX context, UINT extension, LPVOID data);
HMGR APIENTRY WTMgrOpen(HWND window, UINT message);
BOOL APIENTRY WTMgrClose(HMGR manager);
BOOL APIENTRY WTMgrContextEnum(HMGR manager, WTENUMPROC callback, LPARAM lparam);
HWND APIENTRY WTMgrContextOwner(HMGR manager, HCTX context);
HCTX APIENTRY WTMgrDefContext(HMGR manager, BOOL system_context);
HCTX APIENTRY WTMgrDefContextEx(HMGR manager, UINT device, BOOL system_context);
BOOL APIENTRY WTMgrDeviceConfig(HMGR manager, UINT device, HWND owner);
