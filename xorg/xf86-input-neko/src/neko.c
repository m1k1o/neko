/*
 * (c) 2017 Martin Kepplinger <martink@posteo.de>
 * (c) 2007 Clement Chauplannaz, Thales e-Transactions <chauplac@gmail.com>
 * (c) 2006 Sascha Hauer, Pengutronix <s.hauer@pengutronix.de>
 * (c) 2023 Miroslav Sedivy
 *
 * derived from the xf86-input-void driver
 * Copyright 1999 by Frederic Lepied, France. <Lepied@XFree86.org>
 *
 * Permission to use, copy, modify, distribute, and sell this software and its
 * documentation for any purpose is  hereby granted without fee, provided that
 * the  above copyright   notice appear  in   all  copies and  that both  that
 * copyright  notice   and   this  permission   notice  appear  in  supporting
 * documentation, and that   the  name of  Frederic   Lepied not  be  used  in
 * advertising or publicity pertaining to distribution of the software without
 * specific,  written      prior  permission.     Frederic  Lepied   makes  no
 * representations about the suitability of this software for any purpose.  It
 * is provided "as is" without express or implied warranty.
 *
 * FREDERIC  LEPIED DISCLAIMS ALL   WARRANTIES WITH REGARD  TO  THIS SOFTWARE,
 * INCLUDING ALL IMPLIED   WARRANTIES OF MERCHANTABILITY  AND   FITNESS, IN NO
 * EVENT  SHALL FREDERIC  LEPIED BE   LIABLE   FOR ANY  SPECIAL, INDIRECT   OR
 * CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE,
 * DATA  OR PROFITS, WHETHER  IN  AN ACTION OF  CONTRACT,  NEGLIGENCE OR OTHER
 * TORTIOUS  ACTION, ARISING    OUT OF OR   IN  CONNECTION  WITH THE USE    OR
 * PERFORMANCE OF THIS SOFTWARE.
 *
 * SPDX-License-Identifier: MIT
 * License-Filename: COPYING
 */

/* neko input driver */
// https://www.x.org/releases/X11R7.7/doc/xorg-server/Xinput.html

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#define DEF_SOCKET_NAME "/tmp/xf86-input-neko.sock"
#define BUFFER_SIZE 12

#include <stdio.h>
#include <stdio.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <misc.h>
#include <xf86.h>
#if !defined(DGUX)
#include <xisb.h>
#endif
#include <xf86_OSproc.h>
#include <xf86Xinput.h>
#include <exevents.h> /* Needed for InitValuator/Proximity stuff */
#include <X11/keysym.h>
#include <mipointer.h>
#include <xserver-properties.h>
#include <pthread.h>

#define MAX_USED_VALUATORS 3 /* x, y, pressure */
#define TOUCH_MAX_SLOTS 10 /* max number of simultaneous touches */

struct neko_message
{
    uint16_t type;
    uint32_t touchId;
    int32_t x;
    int32_t y;
    uint8_t pressure;
};

struct neko_priv
{
    pthread_t thread;
    /* config */
    int height;
    int width;
    int pmax;
    ValuatorMask *valuators;
    uint16_t slots;
    /* socket */
    struct sockaddr_un addr;
    int listen_socket;
    char *socket_name;
};

// from binary representation to struct
static void
UnpackNekoMessage(struct neko_message *msg, unsigned char *buffer)
{
    msg->type = buffer[0]; // TODO: use full 16bit type
    msg->touchId = buffer[1] | (buffer[2] << 8); // TODO: use full 32bit touchId
    msg->x = buffer[3] | (buffer[4] << 8) | (buffer[5] << 16) | (buffer[6] << 24);
    msg->y = buffer[7] | (buffer[8] << 8) | (buffer[9] << 16) | (buffer[10] << 24);
    msg->pressure = buffer[11];
}

static void
ReadInput(InputInfoPtr pInfo)
{
    struct neko_priv *priv = (struct neko_priv *) (pInfo->private);
    struct neko_message msg;
    int ret;

    int data_socket;
    unsigned char buffer[BUFFER_SIZE];

    for (;;)
    {
        /* Wait for incoming connection. */
        data_socket = accept(priv->listen_socket, NULL, NULL);

        /* Handle error conditions. */
        if (data_socket == -1)
        {
            xf86IDrvMsg(pInfo, X_ERROR, "unable to accept connection\n");
            break;
        }

        xf86IDrvMsg(pInfo, X_INFO, "accepted connection\n");

        for(;;)
        {
            /* Wait for next data packet. */
            ret = read(data_socket, buffer, BUFFER_SIZE);

            /* Handle error conditions. */
            if (ret == -1)
            {
                xf86IDrvMsg(pInfo, X_ERROR, "unable to read data\n");
                break;
            }

            /* Connection closed by client. */
            if (ret == 0)
            {
                xf86IDrvMsg(pInfo, X_INFO, "connection closed\n");
                break;
            }

            /* Ensure message is long enough. */
            if (ret != BUFFER_SIZE)
            {
                xf86IDrvMsg(pInfo, X_ERROR, "invalid message size\n");
                break;
            }

            UnpackNekoMessage(&msg, buffer);

            ValuatorMask *m = priv->valuators;
            valuator_mask_zero(m);

            // do not send valuators if x and y are -1
            if (msg.x != -1 && msg.y != -1)
            {
                valuator_mask_set_double(m, 0, msg.x);
                valuator_mask_set_double(m, 1, msg.y);
                valuator_mask_set_double(m, 2, msg.pressure);
            }

            // TODO: extend to other types, such as keyboard and mouse
            xf86PostTouchEvent(pInfo->dev, msg.touchId, msg.type, 0, m);
        }

        /* Close socket. */
        close(data_socket);

        xf86IDrvMsg(pInfo, X_INFO, "closed connection\n");
    }
}

static void
PointerCtrl(__attribute__ ((unused)) DeviceIntPtr device,
            __attribute__ ((unused)) PtrCtrl *ctrl)
{
}

static int
InitTouch(InputInfoPtr pInfo)
{
    // custom private data
    struct neko_priv *priv = pInfo->private;

	const int nbtns = 11;
	const int naxes = 3;

    unsigned char map[nbtns + 1];
    Atom btn_labels[nbtns];
    Atom axis_labels[naxes];

    // init button map
    memset(map, 0, sizeof(map));
    for (int i = 0; i < nbtns; i++)
    {
        map[i + 1] = i + 1;
    }

    // init btn_labels
    memset(btn_labels, 0, ARRAY_SIZE(btn_labels) * sizeof(Atom));
    btn_labels[0] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_LEFT);
    btn_labels[1] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_MIDDLE);
    btn_labels[2] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_RIGHT);
    btn_labels[3] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_WHEEL_UP);
    btn_labels[4] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_WHEEL_DOWN);
    btn_labels[5] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_HWHEEL_LEFT);
    btn_labels[6] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_HWHEEL_RIGHT);
    btn_labels[7] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_SIDE);
    btn_labels[8] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_EXTRA);
    btn_labels[9] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_FORWARD);
    btn_labels[10] = XIGetKnownProperty(BTN_LABEL_PROP_BTN_BACK);

    // init axis labels
    memset(axis_labels, 0, ARRAY_SIZE(axis_labels) * sizeof(Atom));
    axis_labels[0] = XIGetKnownProperty(AXIS_LABEL_PROP_ABS_MT_POSITION_X);
    axis_labels[1] = XIGetKnownProperty(AXIS_LABEL_PROP_ABS_MT_POSITION_Y);
    axis_labels[2] = XIGetKnownProperty(AXIS_LABEL_PROP_ABS_MT_PRESSURE);

    /* initialize mouse emulation valuators */
    if (InitPointerDeviceStruct((DevicePtr)pInfo->dev,
            map,
            nbtns, btn_labels,
            PointerCtrl,
            GetMotionHistorySize(),
            naxes, axis_labels) == FALSE)
    {
        xf86IDrvMsg(pInfo, X_ERROR,
            "unable to allocate PointerDeviceStruct\n");
        return !Success;
    }

    /* 
        This function is provided to initialize an XAxisInfoRec, and should be
        called for core and extension devices that have valuators. The space
        for the XAxisInfoRec is allocated by the InitValuatorClassDeviceStruct
        function, but is not initialized.
        
        InitValuatorAxisStruct should be called once for each axis of motion
        reported by the device. Each invocation should be passed the axis
        number (starting with 0), the minimum value for that axis, the maximum
        value for that axis, and the resolution of the device in counts per meter.
        If the device reports relative motion, 0 should be reported as the
        minimum and maximum values.

        InitValuatorAxisStruct(dev, axnum, minval, maxval, resolution)
            DeviceIntPtr dev;
            int axnum;
            int minval;
            int maxval;
            int resolution;
    */
    xf86InitValuatorAxisStruct(pInfo->dev, 0,
        XIGetKnownProperty(AXIS_LABEL_PROP_ABS_MT_POSITION_X),
        0,                /* min val */
        priv->width - 1,  /* max val */
        priv->width,      /* resolution */
        0,                /* min_res */
        priv->width,      /* max_res */
        Absolute);

    xf86InitValuatorAxisStruct(pInfo->dev, 1,
        XIGetKnownProperty(AXIS_LABEL_PROP_ABS_MT_POSITION_Y),
        0,                /* min val */
        priv->height - 1, /* max val */
        priv->height,     /* resolution */
        0,                /* min_res */
        priv->height,     /* max_res */
        Absolute);

    xf86InitValuatorAxisStruct(pInfo->dev, 2,
        XIGetKnownProperty(AXIS_LABEL_PROP_ABS_MT_PRESSURE),
        0,                /* min val */
        priv->pmax,       /* max val */
        priv->pmax + 1,   /* resolution */
        0,                /* min_res */
        priv->pmax + 1,   /* max_res */
        Absolute);

    /*
        The mode field is either XIDirectTouch for direct−input touch devices
        such as touchscreens or XIDependentTouch for indirect input devices such
        as touchpads. For XIDirectTouch devices, touch events are sent to window
        at the position the touch occured. For XIDependentTouch devices, touch
        events are sent to the window at the position of the device's sprite.

        The num_touches field defines the maximum number of simultaneous touches
        the device supports. A num_touches of 0 means the maximum number of
        simultaneous touches is undefined or unspecified. This field should be
        used as a guide only, devices will lie about their capabilities.
    */
    if (InitTouchClassDeviceStruct(pInfo->dev,
            priv->slots,
            XIDirectTouch,
            naxes) == FALSE)
    {
        xf86IDrvMsg(pInfo, X_ERROR,
            "unable to allocate TouchClassDeviceStruct\n");
        return !Success;
    }

    return Success;
}

static int
DeviceControl(DeviceIntPtr device, int what)
{
    // device pInfo
    InputInfoPtr pInfo = device->public.devicePrivate;
    // custom private data
    struct neko_priv *priv = pInfo->private;

    switch (what) {
    case DEVICE_INIT:
        device->public.on = FALSE;

        if (InitTouch(pInfo) != Success)
        {
            xf86IDrvMsg(pInfo, X_ERROR, "unable to init touch\n");
            return !Success;
        }
        break;

    case DEVICE_ON:
        xf86IDrvMsg(pInfo, X_INFO, "DEVICE ON\n");
        device->public.on = TRUE;

        if (priv->thread == 0)
        {
            /* start thread */
            pthread_create(&priv->thread, NULL, (void *)ReadInput, pInfo);
        }
        break;

    case DEVICE_OFF:
    case DEVICE_CLOSE:
        xf86IDrvMsg(pInfo, X_INFO, "DEVICE OFF\n");
        device->public.on = FALSE;
        break;
    }

    return Success;
}

static int
PreInit(__attribute__ ((unused)) InputDriverPtr drv,
        InputInfoPtr pInfo,
        __attribute__ ((unused)) int flags)
{
    struct neko_priv *priv;
    int ret;

    priv = calloc(1, sizeof (struct neko_priv));
    if (!priv)
    {
        xf86IDrvMsg(pInfo, X_ERROR, "%s: out of memory\n", __FUNCTION__);
        return BadValue;
    }

    pInfo->type_name      = (char*)XI_TOUCHSCREEN;
    pInfo->device_control = DeviceControl;
    pInfo->read_input     = NULL;
    pInfo->control_proc   = NULL;
    pInfo->switch_mode    = NULL; /* Only support Absolute mode */
    pInfo->private        = priv;
    pInfo->fd             = -1;

    /* get socket name from config */
    priv->socket_name = xf86SetStrOption(pInfo->options, "SocketName", DEF_SOCKET_NAME);

    /*
    * In case the program exited inadvertently on the last run,
    * remove the socket.
    */

    unlink(priv->socket_name);

    /* Create local socket. */
    priv->listen_socket = socket(AF_UNIX, SOCK_STREAM, 0);
    if (priv->listen_socket == -1)
    {
        xf86IDrvMsg(pInfo, X_ERROR, "unable to create socket\n");
        return BadValue;
    }

    /*
    * For portability clear the whole structure, since some
    * implementations have additional (nonstandard) fields in
    * the structure.
    */

    memset(&priv->addr, 0, sizeof(struct sockaddr_un));

    /* Bind socket to socket name. */

    priv->addr.sun_family = AF_UNIX;
    strncpy(priv->addr.sun_path, priv->socket_name, sizeof(priv->addr.sun_path) - 1);

    ret = bind(priv->listen_socket, (const struct sockaddr *) &priv->addr, sizeof(struct sockaddr_un));
    if (ret == -1)
    {
        xf86IDrvMsg(pInfo, X_ERROR, "unable to bind socket\n");
        return BadValue;
    }

    /*
    * Prepare for accepting connections. The backlog size is set
    * to 5. So while one request is being processed other requests
    * can be waiting.
    */

    ret = listen(priv->listen_socket, 5);
    if (ret == -1)
    {
        xf86IDrvMsg(pInfo, X_ERROR, "unable to listen on socket\n");
        return BadValue;
    }

    /* process generic options */
    xf86CollectInputOptions(pInfo, NULL);
    xf86ProcessCommonOptions(pInfo, pInfo->options);

    /* create valuators */
    priv->valuators = valuator_mask_new(MAX_USED_VALUATORS);
    if (!priv->valuators)
    {
        xf86IDrvMsg(pInfo, X_ERROR, "%s: out of memory\n", __FUNCTION__);
        return BadValue;
    }

    priv->slots = TOUCH_MAX_SLOTS;
    priv->width = 0xffff;
    priv->height = 0xffff;
    priv->pmax = 255;
    priv->thread = 0;

    /* Return the configured device */
    return Success;
}

static void
UnInit(__attribute__ ((unused)) InputDriverPtr drv,
       InputInfoPtr pInfo,
       __attribute__ ((unused)) int flags)
{
    struct neko_priv *priv = (struct neko_priv *)(pInfo->private);

    /* close socket */
    close(priv->listen_socket);
    /* remove socket file */
    unlink(priv->socket_name);

    if (priv->thread)
    {
        /* cancel thread */
        pthread_cancel(priv->thread);
        /* wait for thread to finish */
        pthread_join(priv->thread, NULL);
        /* ensure thread is not cancelled again */
        priv->thread = 0;
    }

    /* free valuators */
    valuator_mask_free(&priv->valuators);

    free(pInfo->private);
    pInfo->private = NULL;
    xf86DeleteInput(pInfo, 0);
}

/**
 * X module information and plug / unplug routines
 */

_X_EXPORT InputDriverRec NEKO =
{
    .driverVersion = 1,
    .driverName    = "neko",
	.Identify      = NULL,
    .PreInit       = PreInit,
    .UnInit        = UnInit,
    .module        = NULL
};

static pointer
Plug(pointer module,
     __attribute__ ((unused)) pointer options,
     __attribute__ ((unused)) int *errmaj,
     __attribute__ ((unused)) int *errmin)
{
    xf86AddInputDriver(&NEKO, module, 0);
    return module;
}

static void
Unplug(__attribute__ ((unused)) pointer module)
{
}

static XF86ModuleVersionInfo versionRec =
{
	.modname      = "neko",
	.vendor       = MODULEVENDORSTRING,
	._modinfo1_   = MODINFOSTRING1,
	._modinfo2_   = MODINFOSTRING2,
	.xf86version  = XORG_VERSION_CURRENT,
	.majorversion = PACKAGE_VERSION_MAJOR,
	.minorversion = PACKAGE_VERSION_MINOR,
	.patchlevel   = PACKAGE_VERSION_PATCHLEVEL,
	.abiclass     = ABI_CLASS_XINPUT,
	.abiversion   = ABI_XINPUT_VERSION,
	.moduleclass  = MOD_CLASS_XINPUT,
    .checksum     = {0, 0, 0, 0} /* signature, to be patched into the file by a tool */
};

_X_EXPORT XF86ModuleData nekoModuleData =
{
    .vers     = &versionRec,
    .setup    = Plug,
    .teardown = Unplug
};
