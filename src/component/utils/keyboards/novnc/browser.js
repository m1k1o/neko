/*
 * noVNC: HTML5 VNC client
 * Copyright (C) 2019 The noVNC Authors
 * Licensed under MPL 2.0 (see LICENSE.txt)
 *
 * See README.md for usage and integration instructions.
 *
 * Browser feature support detection
 */

export function isMac() {
    return navigator && !!(/mac/i).exec(navigator.platform);
}

export function isWindows() {
    return navigator && !!(/win/i).exec(navigator.platform);
}

export function isIOS() {
    return navigator &&
           (!!(/ipad/i).exec(navigator.platform) ||
            !!(/iphone/i).exec(navigator.platform) ||
            !!(/ipod/i).exec(navigator.platform));
}

export function isSafari() {
    return navigator && (navigator.userAgent.indexOf('Safari') !== -1 &&
                         navigator.userAgent.indexOf('Chrome') === -1);
}

export function isIE() {
    return navigator && !!(/trident/i).exec(navigator.userAgent);
}

export function isEdge() {
    return navigator && !!(/edge/i).exec(navigator.userAgent);
}

export function isFirefox() {
    return navigator && !!(/firefox/i).exec(navigator.userAgent);
}

