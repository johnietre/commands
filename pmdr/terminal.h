#ifndef TERMINAL_H
#define TERMINAL_H

#include <stdio.h>
#include <unistd.h>
#include <termios.h>

static int old = 1;
static struct termios old_tio, new_tio;

int set_terminal_mode() {
  if (!old)
    return 1;
	tcgetattr(STDIN_FILENO, &old_tio); // Get terminal settings for stdin
	new_tio=old_tio; // Keep old settings and restore them at the end
	new_tio.c_lflag &=(~ICANON & ~ECHO); // diable canonical and local echo
  old = 0;
  return !tcsetattr(STDIN_FILENO, TCSANOW, &new_tio);
}

int reset_terminal_mode() {
  if (old)
    return 1;
  old = 1;
  return !tcsetattr(STDIN_FILENO, TCSANOW, &old_tio);
}

#endif
