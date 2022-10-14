#include "terminal.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <ncurses.h>

int main(int argc, char **argv) {
  /*
  printf("\x1b[?47h");
  puts("a");
  puts("a");
  puts("a");
  puts("a");
  printf("\x1b[2J"); fflush(stdout); sleep(3);
  printf("\x1b[H"); fflush(stdout); sleep(3);
  printf("\x1b[?47l"); fflush(stdout); sleep(3);
  */
  set_terminal_mode();
  puts("Trying some stuff");
  fflush(stdout);
  char ptr[2];
  fputs("\x1b[6n", stdout);
  fflush(stdout);
  do {
    int res;
    if ((res = read(STDIN_FILENO, ptr, 1)) < 1 || ptr == NULL) {
      printf("no good: %s\n", strerror(errno));
      reset_terminal_mode();
      return 1;
    }
    if (*ptr >= '0') {
      puts("Got something");
      putchar(*ptr);
    }
  } while (*ptr != 'R');
  reset_terminal_mode();
  return 0;
}
