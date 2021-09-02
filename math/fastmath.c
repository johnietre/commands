#include "ansi_codes.h"
#include "terminal.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

int a, q;

int getNum() {
  int n = 0;
  while (1) {
    char c = getchar();
    if (c >= '0' && c <= '9' && !a && !c) {
      n = (n * 10) + (c - '0');
      putchar(c);
    } else if (c == 'q') {
      if (!n && !a) {
        q = 1;
        putchar(c);
      }
    } else if (c == 'a') {
      if (!n && !q) {
        a = 1;
        putchar(c);
      }
    } else if (c == 127) {
      if (a) {
        a = 0;
      } else if (q) {
        q = 0;
      } else {
        n /= 10;
      }
      printf("%s %s", CUR_LEFT, CUR_LEFT);
    } else if (c == '\n') {
      putchar(c);
      return n;
    }
  }
}

int main(int argc, char **argv) {
  set_terminal_mode();
  srand(time(0));
  int n1 = rand(), n2 = rand();
  int n = getNum();
  printf("%d\t%d\t%d\n", a, q, n);
  reset_terminal_mode();
  return 0;
}
