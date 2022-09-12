#pragma once

#include <libclipboard.h>
#include <string.h>

clipboard_c *getClipboard(void);

void ClipboardSet(char *src);
char *ClipboardGet();
