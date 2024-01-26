#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

/* All drivers should typically include these */
#include "xf86.h"
#include "xf86_OSproc.h"

#include "xf86Cursor.h"
#include "cursorstr.h"
/* Driver specific headers */
#include "dummy.h"

static void
dummyShowCursor(ScrnInfoPtr pScrn)
{
    DUMMYPtr dPtr = DUMMYPTR(pScrn);

    /* turn cursor on */
    dPtr->DummyHWCursorShown = TRUE;    
}

static void
dummyHideCursor(ScrnInfoPtr pScrn)
{
    DUMMYPtr dPtr = DUMMYPTR(pScrn);

    /*
     * turn cursor off 
     *
     */
    dPtr->DummyHWCursorShown = FALSE;
}

#define MAX_CURS 64

static void
dummySetCursorPosition(ScrnInfoPtr pScrn, int x, int y)
{
    DUMMYPtr dPtr = DUMMYPTR(pScrn);

/*     unsigned char *_dest = ((unsigned char *)dPtr->FBBase + */
/* 			    pScrn->videoRam * 1024 - 1024); */
    dPtr->cursorX = x;
    dPtr->cursorY = y;
}

static void
dummySetCursorColors(ScrnInfoPtr pScrn, int bg, int fg)
{
    DUMMYPtr dPtr = DUMMYPTR(pScrn);
    
    dPtr->cursorFG = fg;
    dPtr->cursorBG = bg;
}

static void
dummyLoadCursorImage(ScrnInfoPtr pScrn, unsigned char *src)
{
}

static Bool
dummyUseHWCursor(ScreenPtr pScr, CursorPtr pCurs)
{
    DUMMYPtr dPtr = DUMMYPTR(xf86ScreenToScrn(pScr));
    return(!dPtr->swCursor);
}

#if 0
static unsigned char*
dummyRealizeCursor(xf86CursorInfoPtr infoPtr, CursorPtr pCurs)
{
    return NULL;
}
#endif

Bool
DUMMYCursorInit(ScreenPtr pScreen)
{
    DUMMYPtr dPtr = DUMMYPTR(xf86ScreenToScrn(pScreen));

    xf86CursorInfoPtr infoPtr;
    infoPtr = xf86CreateCursorInfoRec();
    if(!infoPtr) return FALSE;

    dPtr->CursorInfo = infoPtr;

    infoPtr->MaxHeight = 64;
    infoPtr->MaxWidth = 64;
    infoPtr->Flags = HARDWARE_CURSOR_TRUECOLOR_AT_8BPP;

    infoPtr->SetCursorColors = dummySetCursorColors;
    infoPtr->SetCursorPosition = dummySetCursorPosition;
    infoPtr->LoadCursorImage = dummyLoadCursorImage;
    infoPtr->HideCursor = dummyHideCursor;
    infoPtr->ShowCursor = dummyShowCursor;
    infoPtr->UseHWCursor = dummyUseHWCursor;
/*     infoPtr->RealizeCursor = dummyRealizeCursor; */
    
    return(xf86InitCursor(pScreen, infoPtr));
}



