
/*
 * Copyright 2002, SuSE Linux AG, Author: Egbert Eich
 */

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

/* All drivers should typically include these */
#include "xf86.h"
#include "xf86_OSproc.h"

/* All drivers initialising the SW cursor need this */
#include "mipointer.h"

/* All drivers using the mi colormap manipulation need this */
#include "micmap.h"

/* identifying atom needed by magnifiers */
#include <X11/Xatom.h>
#include "property.h"

#include "xf86cmap.h"

#include "xf86fbman.h"

#include "fb.h"

#include "picturestr.h"

#include "xf86Crtc.h"

/*
 * Driver data structures.
 */
#include "dummy.h"

/* These need to be checked */
#include <X11/X.h>
#include <X11/Xproto.h>
#include "scrnintstr.h"
#include "servermd.h"

/* Mandatory functions */
static const OptionInfoRec *	DUMMYAvailableOptions(int chipid, int busid);
static void     DUMMYIdentify(int flags);
static Bool     DUMMYProbe(DriverPtr drv, int flags);
static Bool     DUMMYPreInit(ScrnInfoPtr pScrn, int flags);
static Bool     DUMMYScreenInit(SCREEN_INIT_ARGS_DECL);
static Bool     DUMMYEnterVT(VT_FUNC_ARGS_DECL);
static void     DUMMYLeaveVT(VT_FUNC_ARGS_DECL);
static Bool     DUMMYCloseScreen(CLOSE_SCREEN_ARGS_DECL);
static Bool     DUMMYCreateWindow(WindowPtr pWin);
static void     DUMMYFreeScreen(FREE_SCREEN_ARGS_DECL);
static ModeStatus DUMMYValidMode(SCRN_ARG_TYPE arg, DisplayModePtr mode,
                                 Bool verbose, int flags);
static Bool	DUMMYSaveScreen(ScreenPtr pScreen, int mode);

/* Internally used functions */
static Bool	dummyDriverFunc(ScrnInfoPtr pScrn, xorgDriverFuncOp op,
				pointer ptr);


/* static void     DUMMYDisplayPowerManagementSet(ScrnInfoPtr pScrn, */
/* 				int PowerManagementMode, int flags); */

#define DUMMY_VERSION 4000
#define DUMMY_NAME "DUMMY"
#define DUMMY_DRIVER_NAME "dummy"

#define DUMMY_MAJOR_VERSION PACKAGE_VERSION_MAJOR
#define DUMMY_MINOR_VERSION PACKAGE_VERSION_MINOR
#define DUMMY_PATCHLEVEL PACKAGE_VERSION_PATCHLEVEL

#define DUMMY_MAX_WIDTH 32767
#define DUMMY_MAX_HEIGHT 32767

/*
 * This is intentionally screen-independent.  It indicates the binding
 * choice made in the first PreInit.
 */
static int pix24bpp = 0;

Atom width_mm_atom = 0;
#define WIDTH_MM_NAME  "WIDTH_MM"
Atom height_mm_atom = 0;
#define HEIGHT_MM_NAME "HEIGHT_MM"

/*
 * This contains the functions needed by the server after loading the driver
 * module.  It must be supplied, and gets passed back by the SetupProc
 * function in the dynamic case.  In the static case, a reference to this
 * is compiled in, and this requires that the name of this DriverRec be
 * an upper-case version of the driver name.
 */

_X_EXPORT DriverRec DUMMY = {
    DUMMY_VERSION,
    DUMMY_DRIVER_NAME,
    DUMMYIdentify,
    DUMMYProbe,
    DUMMYAvailableOptions,
    NULL,
    0,
    dummyDriverFunc
};

static SymTabRec DUMMYChipsets[] = {
    { DUMMY_CHIP,   "dummy" },
    { -1,		 NULL }
};

typedef enum {
    OPTION_SW_CURSOR
} DUMMYOpts;

static const OptionInfoRec DUMMYOptions[] = {
    { OPTION_SW_CURSOR,	"SWcursor",	OPTV_BOOLEAN,	{0}, FALSE },
    { -1,                  NULL,           OPTV_NONE,	{0}, FALSE }
};

#ifdef XFree86LOADER

static MODULESETUPPROTO(dummySetup);

static XF86ModuleVersionInfo dummyVersRec =
{
	"dummy",
	MODULEVENDORSTRING,
	MODINFOSTRING1,
	MODINFOSTRING2,
	XORG_VERSION_CURRENT,
	DUMMY_MAJOR_VERSION, DUMMY_MINOR_VERSION, DUMMY_PATCHLEVEL,
	ABI_CLASS_VIDEODRV,
	ABI_VIDEODRV_VERSION,
	MOD_CLASS_VIDEODRV,
	{0,0,0,0}
};


/************************
 * XRANDR support begin *
 ************************/

static Bool dummy_config_resize(ScrnInfoPtr pScrn, int cw, int ch);
static Bool DUMMYAdjustScreenPixmap(ScrnInfoPtr pScrn, int width, int height);

static const xf86CrtcConfigFuncsRec DUMMYCrtcConfigFuncs = {
    .resize = dummy_config_resize
};


static void
dummy_crtc_dpms(xf86CrtcPtr crtc, int mode)
{
}

static Bool
dummy_crtc_lock (xf86CrtcPtr crtc)
{
    return FALSE;
}

static Bool
dummy_crtc_mode_fixup (xf86CrtcPtr crtc, DisplayModePtr mode,
                              DisplayModePtr adjusted_mode)
{
    return TRUE;
}

static void
dummy_crtc_stub (xf86CrtcPtr crtc)
{
}

static void
dummy_crtc_gamma_set (xf86CrtcPtr crtc, CARD16 *red,
                             CARD16 *green, CARD16 *blue, int size)
{
}

static void *
dummy_crtc_shadow_allocate (xf86CrtcPtr crtc, int width, int height)
{
    return NULL;
}

static void
dummy_crtc_mode_set (xf86CrtcPtr crtc, DisplayModePtr mode,
                            DisplayModePtr adjusted_mode, int x, int y)
{
}

static const xf86CrtcFuncsRec DUMMYCrtcFuncs = {
    .dpms = dummy_crtc_dpms,
    .save = NULL, /* These two are never called by the server. */
    .restore = NULL,
    .lock = dummy_crtc_lock,
    .unlock = NULL, /* This will not be invoked if lock returns FALSE. */
    .mode_fixup = dummy_crtc_mode_fixup,
    .prepare = dummy_crtc_stub,
    .mode_set = dummy_crtc_mode_set,
    .commit = dummy_crtc_stub,
    .gamma_set = dummy_crtc_gamma_set,
    .shadow_allocate = dummy_crtc_shadow_allocate,
    .shadow_create = NULL, /* These two should not be invoked if allocate
                              returns NULL. */
    .shadow_destroy = NULL,
    .set_cursor_colors = NULL,
    .set_cursor_position = NULL,
    .show_cursor = NULL,
    .hide_cursor = NULL,
    .load_cursor_argb = NULL,
    .destroy = dummy_crtc_stub
};

static void
dummy_output_stub (xf86OutputPtr output)
{
}

static void
dummy_output_dpms (xf86OutputPtr output, int mode)
{
}

static int
dummy_output_mode_valid (xf86OutputPtr output, DisplayModePtr mode)
{
    return MODE_OK;
}

static Bool
dummy_output_mode_fixup (xf86OutputPtr output, DisplayModePtr mode,
        DisplayModePtr adjusted_mode)
{
    return TRUE;
}

static void
dummy_output_mode_set (xf86OutputPtr output, DisplayModePtr mode,
        DisplayModePtr adjusted_mode)
{
    DUMMYPtr dPtr = DUMMYPTR(output->scrn);
    int index = (int64_t)output->driver_private;

    /* set to connected at first mode set */
    dPtr->connected_outputs |= 1 << index;
}

/* The first virtual monitor is always connected. Others only after setting its
 * mode */
static xf86OutputStatus
dummy_output_detect (xf86OutputPtr output)
{
    DUMMYPtr dPtr = DUMMYPTR(output->scrn);
    int index = (int64_t)output->driver_private;

    if (dPtr->connected_outputs & (1 << index))
        return XF86OutputStatusConnected;
    else
        return XF86OutputStatusDisconnected;
}

static DisplayModePtr
dummy_output_get_modes (xf86OutputPtr output)
{
    DisplayModePtr pModes = NULL, pMode, pModeSrc;

    /* copy modes from config */
    for (pModeSrc = output->scrn->modes; pModeSrc; pModeSrc = pModeSrc->next)
    {
            pMode = xnfcalloc(1, sizeof(DisplayModeRec));
            memcpy(pMode, pModeSrc, sizeof(DisplayModeRec));
            pMode->next = NULL;
            pMode->prev = NULL;
            pMode->name = strdup(pModeSrc->name);
            pModes = xf86ModesAdd(pModes, pMode);
            if (pModeSrc->next == output->scrn->modes)
                break;
    }
    return pModes;
}

void dummy_output_register_prop(xf86OutputPtr output, Atom prop, uint64_t value)
{
    INT32 dims_range[2] = { 0, 65535 };
    int err;

    err = RRConfigureOutputProperty(output->randr_output, prop, FALSE,
            TRUE, FALSE, 2, dims_range);
    if (err != 0)
        xf86DrvMsg(output->scrn->scrnIndex, X_ERROR,
                                       "RRConfigureOutputProperty error, %d\n", err);

    err = RRChangeOutputProperty(output->randr_output, prop, XA_INTEGER,
            32, PropModeReplace, 1, &value, FALSE, FALSE);
    if (err != 0)
        xf86DrvMsg(output->scrn->scrnIndex, X_ERROR,
                "RRChangeOutputProperty error, %d\n", err);
}

void dummy_output_create_resources(xf86OutputPtr output)
{
    if (!ValidAtom(width_mm_atom))
        width_mm_atom = MakeAtom(WIDTH_MM_NAME, strlen(WIDTH_MM_NAME), 1);
    if (!ValidAtom(height_mm_atom))
        height_mm_atom = MakeAtom(HEIGHT_MM_NAME, strlen(HEIGHT_MM_NAME), 1);

    dummy_output_register_prop(output, width_mm_atom, 0);
    dummy_output_register_prop(output, height_mm_atom, 0);
}

static Bool dummy_output_set_property(xf86OutputPtr output, Atom property,
        RRPropertyValuePtr value)
{

    if (property == width_mm_atom || property == height_mm_atom) {
        INT32 val;

        if (value->type != XA_INTEGER || value->format != 32 ||
                value->size != 1)
        {
            return FALSE;
        }

        val = *(INT32 *)value->data;
        if (property == width_mm_atom)
            output->mm_width = val;
        else if (property == height_mm_atom)
            output->mm_height = val;
        return TRUE;
    }
    return TRUE;
}


static const xf86OutputFuncsRec DUMMYOutputFuncs = {
    .create_resources = dummy_output_create_resources,
    .dpms = dummy_output_dpms,
    .save = NULL, /* These two are never called by the server. */
    .restore = NULL,
    .mode_valid = dummy_output_mode_valid,
    .mode_fixup = dummy_output_mode_fixup,
    .prepare = dummy_output_stub,
    .commit = dummy_output_stub,
    .mode_set = dummy_output_mode_set,
    .detect = dummy_output_detect,
    .get_modes = dummy_output_get_modes,
#ifdef RANDR_12_INTERFACE
    .set_property = dummy_output_set_property,
#endif
    .destroy = dummy_output_stub
};

static Bool
dummy_config_resize(ScrnInfoPtr pScrn, int cw, int ch)
{
    if (!pScrn->vtSema) {
        xf86DrvMsg(pScrn->scrnIndex, X_ERROR,
                "We do not own the active VT, exiting.\n");
        return TRUE;
    }
    return DUMMYAdjustScreenPixmap(pScrn, cw, ch);
}

Bool DUMMYAdjustScreenPixmap(ScrnInfoPtr pScrn, int width, int height)
{
    ScreenPtr pScreen = pScrn->pScreen;
    PixmapPtr pPixmap = pScreen->GetScreenPixmap(pScreen);
    DUMMYPtr dPtr = DUMMYPTR(pScrn);
    uint64_t cbLine = (width * xf86GetBppFromDepth(pScrn, pScrn->depth) / 8 + 3) & ~3;
    int displayWidth = cbLine * 8 / xf86GetBppFromDepth(pScrn, pScrn->depth);

    if (   width == pScrn->virtualX
            && height == pScrn->virtualY
            && displayWidth == pScrn->displayWidth)
        return TRUE;
    if (!pPixmap) {
        xf86DrvMsg(pScrn->scrnIndex, X_ERROR,
                "Failed to get the screen pixmap.\n");
        return FALSE;
    }
    if (cbLine > UINT32_MAX || cbLine * height >= pScrn->videoRam * 1024)
    {
        xf86DrvMsg(pScrn->scrnIndex, X_ERROR,
                "Unable to set up a virtual screen size of %dx%d with %d Kb of video memory available.  Please increase the video memory size.\n",
                width, height, pScrn->videoRam);
        return FALSE;
    }
    pScreen->ModifyPixmapHeader(pPixmap, width, height,
            pScrn->depth, xf86GetBppFromDepth(pScrn, pScrn->depth), cbLine,
            pPixmap->devPrivate.ptr);
    pScrn->virtualX = width;
    pScrn->virtualY = height;
    pScrn->displayWidth = displayWidth;

    return TRUE;
}

/**********************
 * XRANDR support end *
 **********************/

/*
 * This is the module init data.
 * Its name has to be the driver name followed by ModuleData
 */
_X_EXPORT XF86ModuleData dummyModuleData = { &dummyVersRec, dummySetup, NULL };

static pointer
dummySetup(pointer module, pointer opts, int *errmaj, int *errmin)
{
    static Bool setupDone = FALSE;

    if (!setupDone) {
	setupDone = TRUE;
        xf86AddDriver(&DUMMY, module, HaveDriverFuncs);

	/*
	 * Modules that this driver always requires can be loaded here
	 * by calling LoadSubModule().
	 */

	/*
	 * The return value must be non-NULL on success even though there
	 * is no TearDownProc.
	 */
	return (pointer)1;
    } else {
	if (errmaj) *errmaj = LDR_ONCEONLY;
	return NULL;
    }
}

#endif /* XFree86LOADER */

static Bool
DUMMYGetRec(ScrnInfoPtr pScrn)
{
    /*
     * Allocate a DUMMYRec, and hook it into pScrn->driverPrivate.
     * pScrn->driverPrivate is initialised to NULL, so we can check if
     * the allocation has already been done.
     */
    if (pScrn->driverPrivate != NULL)
	return TRUE;

    pScrn->driverPrivate = xnfcalloc(sizeof(DUMMYRec), 1);

    if (pScrn->driverPrivate == NULL)
	return FALSE;
    return TRUE;
}

static void
DUMMYFreeRec(ScrnInfoPtr pScrn)
{
    if (pScrn->driverPrivate == NULL)
	return;
    free(pScrn->driverPrivate);
    pScrn->driverPrivate = NULL;
}

static const OptionInfoRec *
DUMMYAvailableOptions(int chipid, int busid)
{
    return DUMMYOptions;
}

/* Mandatory */
static void
DUMMYIdentify(int flags)
{
    xf86PrintChipsets(DUMMY_NAME, "Driver for Dummy chipsets",
			DUMMYChipsets);
}

/* Mandatory */
static Bool
DUMMYProbe(DriverPtr drv, int flags)
{
    Bool foundScreen = FALSE;
    int numDevSections, numUsed;
    GDevPtr *devSections;
    int i;

    if (flags & PROBE_DETECT)
	return FALSE;
    /*
     * Find the config file Device sections that match this
     * driver, and return if there are none.
     */
    if ((numDevSections = xf86MatchDevice(DUMMY_DRIVER_NAME,
					  &devSections)) <= 0) {
	return FALSE;
    }

    numUsed = numDevSections;

    if (numUsed > 0) {

	for (i = 0; i < numUsed; i++) {
	    ScrnInfoPtr pScrn = NULL;
	    int entityIndex = 
		xf86ClaimNoSlot(drv,DUMMY_CHIP,devSections[i],TRUE);
	    /* Allocate a ScrnInfoRec and claim the slot */
	    if ((pScrn = xf86AllocateScreen(drv,0 ))) {
		   xf86AddEntityToScreen(pScrn,entityIndex);
		    pScrn->driverVersion = DUMMY_VERSION;
		    pScrn->driverName    = DUMMY_DRIVER_NAME;
		    pScrn->name          = DUMMY_NAME;
		    pScrn->Probe         = DUMMYProbe;
		    pScrn->PreInit       = DUMMYPreInit;
		    pScrn->ScreenInit    = DUMMYScreenInit;
		    pScrn->SwitchMode    = DUMMYSwitchMode;
		    pScrn->AdjustFrame   = DUMMYAdjustFrame;
		    pScrn->EnterVT       = DUMMYEnterVT;
		    pScrn->LeaveVT       = DUMMYLeaveVT;
		    pScrn->FreeScreen    = DUMMYFreeScreen;
		    pScrn->ValidMode     = DUMMYValidMode;

		    foundScreen = TRUE;
	    }
	}
    }    

    free(devSections);

    return foundScreen;
}

# define RETURN \
    { DUMMYFreeRec(pScrn);\
			    return FALSE;\
					     }

/* Mandatory */
Bool
DUMMYPreInit(ScrnInfoPtr pScrn, int flags)
{
    ClockRangePtr clockRanges;
    int i;
    DUMMYPtr dPtr;
    int maxClock = 300000;
    GDevPtr device = xf86GetEntityInfo(pScrn->entityList[0])->device;

    if (flags & PROBE_DETECT) 
	return TRUE;
    
    /* Allocate the DummyRec driverPrivate */
    if (!DUMMYGetRec(pScrn)) {
	return FALSE;
    }
    
    dPtr = DUMMYPTR(pScrn);

    pScrn->chipset = (char *)xf86TokenToString(DUMMYChipsets,
					       DUMMY_CHIP);

    xf86DrvMsg(pScrn->scrnIndex, X_INFO, "Chipset is a DUMMY\n");
    
    pScrn->monitor = pScrn->confScreen->monitor;

    if (!xf86SetDepthBpp(pScrn, 0, 0, 0,  Support24bppFb | Support32bppFb))
	return FALSE;
    else {
	/* Check that the returned depth is one we support */
	switch (pScrn->depth) {
	case 8:
	case 15:
	case 16:
	case 24:
	case 30:
	    break;
	default:
	    xf86DrvMsg(pScrn->scrnIndex, X_ERROR,
		       "Given depth (%d) is not supported by this driver\n",
		       pScrn->depth);
	    return FALSE;
	}
    }

    xf86PrintDepthBpp(pScrn);
    if (pScrn->depth == 8)
	pScrn->rgbBits = 8;

    /* Get the depth24 pixmap format */
    if (pScrn->depth == 24 && pix24bpp == 0)
	pix24bpp = xf86GetBppFromDepth(pScrn, 24);

    /*
     * This must happen after pScrn->display has been set because
     * xf86SetWeight references it.
     */
    if (pScrn->depth > 8) {
	/* The defaults are OK for us */
	rgb zeros = {0, 0, 0};

	if (!xf86SetWeight(pScrn, zeros, zeros)) {
	    return FALSE;
	} else {
	    /* XXX check that weight returned is supported */
	    ;
	}
    }

    if (!xf86SetDefaultVisual(pScrn, -1)) 
	return FALSE;

    if (pScrn->depth > 1) {
	Gamma zeros = {0.0, 0.0, 0.0};

	if (!xf86SetGamma(pScrn, zeros))
	    return FALSE;
    }

    xf86CollectOptions(pScrn, device->options);
    /* Process the options */
    if (!(dPtr->Options = malloc(sizeof(DUMMYOptions))))
	return FALSE;
    memcpy(dPtr->Options, DUMMYOptions, sizeof(DUMMYOptions));

    xf86ProcessOptions(pScrn->scrnIndex, pScrn->options, dPtr->Options);

    xf86GetOptValBool(dPtr->Options, OPTION_SW_CURSOR,&dPtr->swCursor);

    if (device->videoRam != 0) {
	pScrn->videoRam = device->videoRam;
	xf86DrvMsg(pScrn->scrnIndex, X_CONFIG, "VideoRAM: %d kByte\n",
		   pScrn->videoRam);
    } else {
	pScrn->videoRam = 4096;
	xf86DrvMsg(pScrn->scrnIndex, X_PROBED, "VideoRAM: %d kByte\n",
		   pScrn->videoRam);
    }
    
    if (device->dacSpeeds[0] != 0) {
	maxClock = device->dacSpeeds[0];
	xf86DrvMsg(pScrn->scrnIndex, X_CONFIG, "Max Clock: %d kHz\n",
		   maxClock);
    } else {
	xf86DrvMsg(pScrn->scrnIndex, X_PROBED, "Max Clock: %d kHz\n",
		   maxClock);
    }

    pScrn->progClock = TRUE;
    /*
     * Setup the ClockRanges, which describe what clock ranges are available,
     * and what sort of modes they can be used for.
     */
    clockRanges = (ClockRangePtr)xnfcalloc(sizeof(ClockRange), 1);
    clockRanges->next = NULL;
    clockRanges->ClockMulFactor = 1;
    clockRanges->minClock = 11000;   /* guessed §§§ */
    clockRanges->maxClock = maxClock;
    clockRanges->clockIndex = -1;		/* programmable */
    clockRanges->interlaceAllowed = TRUE; 
    clockRanges->doubleScanAllowed = TRUE;

    /* Subtract memory for HW cursor */


    {
	int apertureSize = (pScrn->videoRam * 1024);
	i = xf86ValidateModes(pScrn, pScrn->monitor->Modes,
			      pScrn->display->modes, clockRanges,
			      NULL, 256, DUMMY_MAX_WIDTH,
			      (8 * pScrn->bitsPerPixel),
			      128, DUMMY_MAX_HEIGHT, pScrn->display->virtualX,
			      pScrn->display->virtualY, apertureSize,
			      LOOKUP_BEST_REFRESH);

       if (i == -1)
           RETURN;
    }

    /* Prune the modes marked as invalid */
    xf86PruneDriverModes(pScrn);

    if (i == 0 || pScrn->modes == NULL) {
	xf86DrvMsg(pScrn->scrnIndex, X_ERROR, "No valid modes found\n");
	RETURN;
    }

    /*
     * Set the CRTC parameters for all of the modes based on the type
     * of mode, and the chipset's interlace requirements.
     *
     * Calling this is required if the mode->Crtc* values are used by the
     * driver and if the driver doesn't provide code to set them.  They
     * are not pre-initialised at all.
     */
    xf86SetCrtcForModes(pScrn, 0); 
 
    /* Set the current mode to the first in the list */
    pScrn->currentMode = pScrn->modes;

    /* Print the list of modes being used */
    xf86PrintModes(pScrn);

    /* If monitor resolution is set on the command line, use it */
    xf86SetDpi(pScrn, 0, 0);

    if (xf86LoadSubModule(pScrn, "fb") == NULL) {
	RETURN;
    }

    if (!dPtr->swCursor) {
	if (!xf86LoadSubModule(pScrn, "ramdac"))
	    RETURN;
    }
    
    /* We have no contiguous physical fb in physical memory */
    pScrn->memPhysBase = 0;
    pScrn->fbOffset = 0;

    return TRUE;
}
#undef RETURN

/* Mandatory */
static Bool
DUMMYEnterVT(VT_FUNC_ARGS_DECL)
{
    return TRUE;
}

/* Mandatory */
static void
DUMMYLeaveVT(VT_FUNC_ARGS_DECL)
{
}

static void
DUMMYLoadPalette(
   ScrnInfoPtr pScrn,
   int numColors,
   int *indices,
   LOCO *colors,
   VisualPtr pVisual
){
   int i, index, shift, Gshift;
   DUMMYPtr dPtr = DUMMYPTR(pScrn);

   switch(pScrn->depth) {
   case 15:	
	shift = Gshift = 1;
	break;
   case 16:
	shift = 0; 
        Gshift = 0;
	break;
   default:
	shift = Gshift = 0;
	break;
   }

   for(i = 0; i < numColors; i++) {
       index = indices[i];
       dPtr->colors[index].red = colors[index].red << shift;
       dPtr->colors[index].green = colors[index].green << Gshift;
       dPtr->colors[index].blue = colors[index].blue << shift;
   } 

}

static ScrnInfoPtr DUMMYScrn; /* static-globalize it */

/* Mandatory */
static Bool
DUMMYScreenInit(SCREEN_INIT_ARGS_DECL)
{
    ScrnInfoPtr pScrn;
    DUMMYPtr dPtr;
    int ret;
    VisualPtr visual;
    void *pixels;

    /*
     * we need to get the ScrnInfoRec for this screen, so let's allocate
     * one first thing
     */
    pScrn = xf86ScreenToScrn(pScreen);
    dPtr = DUMMYPTR(pScrn);
    DUMMYScrn = pScrn;


    if (!(pixels = malloc(pScrn->videoRam * 1024)))
	return FALSE;

    /*
     * Reset visual list.
     */
    miClearVisualTypes();
    
    /* Setup the visuals we support. */
    
    if (!miSetVisualTypes(pScrn->depth,
                          miGetDefaultVisualMask(pScrn->depth),
                          pScrn->rgbBits, pScrn->defaultVisual)) {
        free(pixels);
        return FALSE;
    }

    if (!miSetPixmapDepths ()) {
        free(pixels);
        return FALSE;
    }

    /*
     * Call the framebuffer layer's ScreenInit function, and fill in other
     * pScreen fields.
     */
    ret = fbScreenInit(pScreen, pixels,
			    pScrn->virtualX, pScrn->virtualY,
			    pScrn->xDpi, pScrn->yDpi,
			    pScrn->displayWidth, pScrn->bitsPerPixel);
    if (!ret)
	return FALSE;

    if (pScrn->depth > 8) {
        /* Fixup RGB ordering */
        visual = pScreen->visuals + pScreen->numVisuals;
        while (--visual >= pScreen->visuals) {
	    if ((visual->class | DynamicClass) == DirectColor) {
		visual->offsetRed = pScrn->offset.red;
		visual->offsetGreen = pScrn->offset.green;
		visual->offsetBlue = pScrn->offset.blue;
		visual->redMask = pScrn->mask.red;
		visual->greenMask = pScrn->mask.green;
		visual->blueMask = pScrn->mask.blue;
	    }
	}
    }
    
    /* must be after RGB ordering fixed */
    fbPictureInit(pScreen, 0, 0);

    xf86SetBlackWhitePixels(pScreen);

    /* initialize XRANDR */
    xf86CrtcConfigInit(pScrn, &DUMMYCrtcConfigFuncs);
    /* FIXME */
    dPtr->num_screens = DUMMY_MAX_SCREENS;

    for (int i=0; i < dPtr->num_screens; i++) {
        char szOutput[256];

        dPtr->paCrtcs[i] = xf86CrtcCreate(pScrn, &DUMMYCrtcFuncs);
        dPtr->paCrtcs[i]->driver_private = (void *)(uintptr_t)i;

        /* Set up our virtual outputs. */
        snprintf(szOutput, sizeof(szOutput), "DUMMY%u", i);
        dPtr->paOutputs[i] = xf86OutputCreate(pScrn, &DUMMYOutputFuncs,
                szOutput);


        xf86OutputUseScreenMonitor(dPtr->paOutputs[i], FALSE);
        dPtr->paOutputs[i]->possible_crtcs = 1 << i;
        dPtr->paOutputs[i]->possible_clones = 0;
        dPtr->paOutputs[i]->driver_private = (void *)(uintptr_t)i;
        xf86DrvMsg(pScrn->scrnIndex, X_INFO, "Created crtc (%p) and output %s (%p)\n",
                (void *)dPtr->paCrtcs[i], szOutput,
                (void *)dPtr->paOutputs[i]);

    }

    /* bitmask */
    dPtr->connected_outputs = 1;

    xf86CrtcSetSizeRange(pScrn, 64, 64, DUMMY_MAX_WIDTH, DUMMY_MAX_HEIGHT);


    /* Now create our initial CRTC/output configuration. */
    if (!xf86InitialConfiguration(pScrn, TRUE)) {
        xf86DrvMsg(pScrn->scrnIndex, X_ERROR, "Initial CRTC configuration failed!\n");
        return (FALSE);
    }

    /* Initialise randr 1.2 mode-setting functions and set first mode.
     * Note that the mode won't be usable until the server has resized the
     * framebuffer to something reasonable. */
    if (!xf86CrtcScreenInit(pScreen)) {
        return FALSE;
    }
    if (!xf86SetDesiredModes(pScrn)) {
        return FALSE;
    }

    /* XRANDR initialization end */
    if (dPtr->swCursor)
	xf86DrvMsg(pScrn->scrnIndex, X_CONFIG, "Using Software Cursor.\n");

    {

	 
	BoxRec AvailFBArea;
	int lines = pScrn->videoRam * 1024 /
	    (pScrn->displayWidth * (pScrn->bitsPerPixel >> 3));
	AvailFBArea.x1 = 0;
	AvailFBArea.y1 = 0;
	AvailFBArea.x2 = pScrn->displayWidth;
	AvailFBArea.y2 = lines;
	xf86InitFBManager(pScreen, &AvailFBArea); 
	
	xf86DrvMsg(pScrn->scrnIndex, X_INFO, 
		   "Using %i scanlines of offscreen memory \n"
		   , lines - pScrn->virtualY);
    }

    xf86SetBackingStore(pScreen);
    xf86SetSilkenMouse(pScreen);
	
    /* Initialise cursor functions */
    miDCInitialize (pScreen, xf86GetPointerScreenFuncs());


    if (!dPtr->swCursor) {
      /* HW cursor functions */
      if (!DUMMYCursorInit(pScreen)) {
	  xf86DrvMsg(pScrn->scrnIndex, X_ERROR,
		     "Hardware cursor initialization failed\n");
	  return FALSE;
      }
    }
    
    /* Initialise default colourmap */
    if(!miCreateDefColormap(pScreen))
	return FALSE;

    if (!xf86HandleColormaps(pScreen, 1024, pScrn->rgbBits,
                         DUMMYLoadPalette, NULL, 
                         CMAP_PALETTED_TRUECOLOR 
			     | CMAP_RELOAD_ON_MODE_SWITCH))
	return FALSE;

    pScreen->SaveScreen = DUMMYSaveScreen;

    
    /* Wrap the current CloseScreen function */
    dPtr->CloseScreen = pScreen->CloseScreen;
    pScreen->CloseScreen = DUMMYCloseScreen;

    /* Wrap the current CreateWindow function */
    dPtr->CreateWindow = pScreen->CreateWindow;
    pScreen->CreateWindow = DUMMYCreateWindow;

    /* Report any unused options (only for the first generation) */
    if (serverGeneration == 1) {
	xf86ShowUnusedOptions(pScrn->scrnIndex, pScrn->options);
    }

    return TRUE;
}

/* Mandatory */
Bool
DUMMYSwitchMode(SWITCH_MODE_ARGS_DECL)
{
    return TRUE;
}

/* Mandatory */
void
DUMMYAdjustFrame(ADJUST_FRAME_ARGS_DECL)
{
}

/* Mandatory */
static Bool
DUMMYCloseScreen(CLOSE_SCREEN_ARGS_DECL)
{
    ScrnInfoPtr pScrn = xf86ScreenToScrn(pScreen);
    DUMMYPtr dPtr = DUMMYPTR(pScrn);

    free(pScreen->GetScreenPixmap(pScreen)->devPrivate.ptr);

    if (dPtr->CursorInfo)
	xf86DestroyCursorInfoRec(dPtr->CursorInfo);

    pScrn->vtSema = FALSE;
    pScreen->CloseScreen = dPtr->CloseScreen;
    return (*pScreen->CloseScreen)(CLOSE_SCREEN_ARGS);
}

/* Optional */
static void
DUMMYFreeScreen(FREE_SCREEN_ARGS_DECL)
{
    SCRN_INFO_PTR(arg);
    DUMMYFreeRec(pScrn);
}

static Bool
DUMMYSaveScreen(ScreenPtr pScreen, int mode)
{
    return TRUE;
}

/* Optional */
static ModeStatus
DUMMYValidMode(SCRN_ARG_TYPE arg, DisplayModePtr mode, Bool verbose, int flags)
{
    return(MODE_OK);
}

Atom VFB_PROP  = 0;
#define  VFB_PROP_NAME  "VFB_IDENT"

static Bool
DUMMYCreateWindow(WindowPtr pWin)
{
    ScreenPtr pScreen = pWin->drawable.pScreen;
    DUMMYPtr dPtr = DUMMYPTR(DUMMYScrn);
    WindowPtr pWinRoot;
    int ret;

    pScreen->CreateWindow = dPtr->CreateWindow;
    ret = pScreen->CreateWindow(pWin);
    dPtr->CreateWindow = pScreen->CreateWindow;
    pScreen->CreateWindow = DUMMYCreateWindow;

    if(ret != TRUE)
	return(ret);
	
    if(dPtr->prop == FALSE) {
#if GET_ABI_MAJOR(ABI_VIDEODRV_VERSION) < 8
        pWinRoot = WindowTable[DUMMYScrn->pScreen->myNum];
#else
        pWinRoot = DUMMYScrn->pScreen->root;
#endif
        if (! ValidAtom(VFB_PROP))
            VFB_PROP = MakeAtom(VFB_PROP_NAME, strlen(VFB_PROP_NAME), 1);

        ret = dixChangeWindowProperty(serverClient, pWinRoot, VFB_PROP,
                                      XA_STRING, 8, PropModeReplace,
                                      (int)4, (pointer)"TRUE", FALSE);
	if( ret != Success)
		ErrorF("Could not set VFB root window property");
        dPtr->prop = TRUE;

	return TRUE;
    }
    return TRUE;
}

#ifndef HW_SKIP_CONSOLE
#define HW_SKIP_CONSOLE 4
#endif

static Bool
dummyDriverFunc(ScrnInfoPtr pScrn, xorgDriverFuncOp op, pointer ptr)
{
    CARD32 *flag;
    
    switch (op) {
	case GET_REQUIRED_HW_INTERFACES:
	    flag = (CARD32*)ptr;
	    (*flag) = HW_SKIP_CONSOLE;
	    return TRUE;
	default:
	    return FALSE;
    }
}
