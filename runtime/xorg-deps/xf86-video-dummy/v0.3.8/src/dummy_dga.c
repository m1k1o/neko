#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include "xf86.h"
#include "xf86_OSproc.h"
#include "dgaproc.h"
#include "dummy.h"

static Bool DUMMY_OpenFramebuffer(ScrnInfoPtr, char **, unsigned char **, 
					int *, int *, int *);
static Bool DUMMY_SetMode(ScrnInfoPtr, DGAModePtr);
static int  DUMMY_GetViewport(ScrnInfoPtr);
static void DUMMY_SetViewport(ScrnInfoPtr, int, int, int);

static
DGAFunctionRec DUMMYDGAFuncs = {
   DUMMY_OpenFramebuffer,
   NULL,
   DUMMY_SetMode,
   DUMMY_SetViewport,
   DUMMY_GetViewport,
   NULL,
   NULL,
   NULL,
#if 0
   DUMMY_BlitTransRect
#else
   NULL
#endif
};

Bool
DUMMYDGAInit(ScreenPtr pScreen)
{   
   ScrnInfoPtr pScrn = xf86ScreenToScrn(pScreen);
   DUMMYPtr pDUMMY = DUMMYPTR(pScrn);
   DGAModePtr modes = NULL, newmodes = NULL, currentMode;
   DisplayModePtr pMode, firstMode;
   int Bpp = pScrn->bitsPerPixel >> 3;
   int num = 0, imlines, pixlines;

   imlines =  (pScrn->videoRam * 1024) /
      (pScrn->displayWidth * (pScrn->bitsPerPixel >> 3));

   pixlines =   imlines;

   pMode = firstMode = pScrn->modes;

   while(pMode) {

	newmodes = realloc(modes, (num + 1) * sizeof(DGAModeRec));

	if(!newmodes) {
	   free(modes);
	   return FALSE;
	}
	modes = newmodes;

	currentMode = modes + num;
	num++;

	currentMode->mode = pMode;
	currentMode->flags = DGA_CONCURRENT_ACCESS | DGA_PIXMAP_AVAILABLE;
	if(pMode->Flags & V_DBLSCAN)
	   currentMode->flags |= DGA_DOUBLESCAN;
	if(pMode->Flags & V_INTERLACE)
	   currentMode->flags |= DGA_INTERLACED;
	currentMode->byteOrder = pScrn->imageByteOrder;
	currentMode->depth = pScrn->depth;
	currentMode->bitsPerPixel = pScrn->bitsPerPixel;
	currentMode->red_mask = pScrn->mask.red;
	currentMode->green_mask = pScrn->mask.green;
	currentMode->blue_mask = pScrn->mask.blue;
	currentMode->visualClass = (Bpp == 1) ? PseudoColor : TrueColor;
	currentMode->viewportWidth = pMode->HDisplay;
	currentMode->viewportHeight = pMode->VDisplay;
	currentMode->xViewportStep = 1;
	currentMode->yViewportStep = 1;
	currentMode->viewportFlags = DGA_FLIP_RETRACE;
	currentMode->offset = 0;
	currentMode->address = (unsigned char *)pDUMMY->FBBase;

	currentMode->bytesPerScanline = 
			((pScrn->displayWidth * Bpp) + 3) & ~3L;
	currentMode->imageWidth = pScrn->displayWidth;
	currentMode->imageHeight =  imlines;
	currentMode->pixmapWidth = currentMode->imageWidth;
	currentMode->pixmapHeight = pixlines;
	currentMode->maxViewportX = currentMode->imageWidth - 
					currentMode->viewportWidth;
	currentMode->maxViewportY = currentMode->imageHeight -
					currentMode->viewportHeight;

	pMode = pMode->next;
	if(pMode == firstMode)
	   break;
   }

   pDUMMY->numDGAModes = num;
   pDUMMY->DGAModes = modes;

   return DGAInit(pScreen, &DUMMYDGAFuncs, modes, num);  
}

static DisplayModePtr DUMMYSavedDGAModes[MAXSCREENS];

static Bool
DUMMY_SetMode(
   ScrnInfoPtr pScrn,
   DGAModePtr pMode
){
   int index = pScrn->pScreen->myNum;
   DUMMYPtr pDUMMY = DUMMYPTR(pScrn);

   if(!pMode) { /* restore the original mode */
 	if(pDUMMY->DGAactive) {	
	    pScrn->currentMode = DUMMYSavedDGAModes[index];
            DUMMYSwitchMode(SWITCH_MODE_ARGS(pScrn, pScrn->currentMode));
	    DUMMYAdjustFrame(ADJUST_FRAME_ARGS(pScrn, 0, 0));
 	    pDUMMY->DGAactive = FALSE;
	}
   } else {
	if(!pDUMMY->DGAactive) {  /* save the old parameters */
	    DUMMYSavedDGAModes[index] = pScrn->currentMode;
	    pDUMMY->DGAactive = TRUE;
	}

        DUMMYSwitchMode(SWITCH_MODE_ARGS(pScrn, pMode->mode));
   }
   
   return TRUE;
}

static int  
DUMMY_GetViewport(
  ScrnInfoPtr pScrn
){
    DUMMYPtr pDUMMY = DUMMYPTR(pScrn);

    return pDUMMY->DGAViewportStatus;
}

static void 
DUMMY_SetViewport(
   ScrnInfoPtr pScrn, 
   int x, int y, 
   int flags
){
   DUMMYPtr pDUMMY = DUMMYPTR(pScrn);

   DUMMYAdjustFrame(ADJUST_FRAME_ARGS(pScrn, x, y));
   pDUMMY->DGAViewportStatus = 0;  
}


static Bool 
DUMMY_OpenFramebuffer(
   ScrnInfoPtr pScrn, 
   char **name,
   unsigned char **mem,
   int *size,
   int *offset,
   int *flags
){
    DUMMYPtr pDUMMY = DUMMYPTR(pScrn);

    *name = NULL; 		/* no special device */
    *mem = (unsigned char*)pDUMMY->FBBase;
    *size = pScrn->videoRam * 1024;
    *offset = 0;
    *flags = DGA_NEED_ROOT;

    return TRUE;
}
