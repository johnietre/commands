/* TODO:
 * Carefully read thru parser and determine what needs to be fixed/finished
 * Handle for multiple "-"s separated by a space
 * Handle for multiple "*" or "/" when separated by a space (not allowed)
 * Handle invalid operations during initial parsing (?)
 * Possibly handle factorial differently
 * Finish calc function
 */

/*
 * Operators:
 * "+": addition
 * "-"(unary): negative
 * "-"(binary): subtraction
 * "*": multiplication
 * "/": division
 * "//": floor division
 * "^", "**": power
 * "%"(unary): percent (?)
 * "%"(binary): modulo
 * "!": factorial

 * Other:
 * Accept functions like trig, log, ln
 * Accept things like combinations (ex: 2C3)
 * Accept absolue value (||)
 * Accept "root" as function to find roots (?)
 */

#include "ansi_codes.h" // SET_GRAPHICS, BACK_RED, RESET_GRAPHICS
#include <ctype.h>      // isdigit, isspace
#include <math.h>       // math functions
#include <stdio.h>      // NULL, fprintf, fputc, stderr
#include <stdlib.h>     // exit, malloc, free
#include <string.h>     // memcpy, strlen

double pi;
double e;

void die(const char *msg, int code, const char *expr, int start, int stop);
void check(const char *expr);
double evaluate(const char *expr);
double calc(double x, double y, const char *op);
double factorial(double x);

int main(int argc, char **argv) {
  if (argc != 2)
    die("Usage: math [expression]", 1, NULL, 0, 0);

  char *expr = argv[1];
  check(expr);

  // Assign values to mathematical constants
  pi = 4 * atan(1);
  e = exp(1);

  return 0;
}

void die(const char *msg, int code, const char *expr, int start, int end) {
  if (msg != NULL)
    fprintf(stderr, "%s\n", msg);
  if (expr == NULL)
    exit(code);
  for (int i = 0; expr[i] != '\0'; i++) {
    if (i == start)
      fprintf(stderr, SET_GRAPHICS(BACK_RED));
    else if (i == end)
      fprintf(stderr, RESET_GRAPHICS);
    fputc(expr[i], stderr);
  }
  fprintf(stderr, "%s\n", RESET_GRAPHICS);
  exit(code);
}

void check(const char *expr) {
  // "prev" holds the character from the previous iteration
  char prev;
  // "parens" is used to make sure parentheses are matching
  // "open_parens" keeps track of where the set's opening parenthesis is
  // "decimal" keeps track of the last decimal in a number (not necessary?)
  // "two" tells whether there are already to of an operator (ex: -- or //)
  int parens = 0, open_parens = -1, decimal = -1, two = 0;
  for (int i = 0, l = strlen(expr); i < l; i++) {
    char c = expr[i];
    if (isdigit(c)) {
      prev = c;
      continue;
    } else if (isspace(c)) {
      continue;
    }
    if (prev == '.')
      die("Invalid expression", 1, expr, i - 1, i);
    if (c == '.') {
      if (decimal != -1)
        die("Invalid expression", 1, expr, i, i + 1);
      decimal = i;
      prev = c;
      continue; // continue because we don't want to reset "decimal" variable
    } else if (c == '*' || c == '/') {
      // These can have 2 in a row but must not be separated by a space
      if (expr[i - 1] == c) {
        if (two)
          die("Invalid expression", 1, expr, i, i + 1);
        two = 1;
      } else {
        two = 0;
      }
      prev = c;
      decimal = -1;
      continue;
    } else if (c == '+' || c == '^' || c == '-' || c == '%') {
      if (prev == c)
        die("Invalid expression", 1, expr, i - 1, i + 1);
    } else if (c == '(') {
      if (open_parens == -1)
        open_parens = i;
      parens++;
    } else if (c == ')') {
      if (parens == 0)
        die("Mismatch parentheses", 1, expr, i, i + 1);
      else if (parens == 1)
        open_parens = -1;
      parens--;
    } else {
      die("Invalid character", 1, expr, i, i + 1);
    }
    prev = c;
    decimal = -1;
  }
  if (parens)
    die("Mismatch parentheses", 1, expr, open_parens, open_parens + 1);
}

double evaluate(const char *expr) {
  double res = 0;
  return 0.0;
}

double calc(double x, double y, const char *op) {
  char c = op[0];
  if (c == '+')
    return x + y;
  if (c == '-')
    return x - y;
  if (c == '*') {
    if (!strcmp(op, "**"))
      return pow(x, y);
    return x * y;
  }
  if (c == '/') {
    if (!strcmp(op, "//"))
      return floor(x);
    return x / y;
  }
  if (c == '^')
    return pow(x, y);
  if (c == '!')
    return factorial(x);
  if (c == '%')
    return (int) x % (int) y;
  if (!strcmp(op, "ln"))
    return log(x);
  if (!strcmp(op, "log")) {
    // x is the argument and y is the optional base
    if (y <= 0)
      return log10(x);
    return log10(x) / log10(y);
  }
  if (!strcmp(op, "sin"))
    return sin(x);
  if (!strcmp(op, "cos"))
    return cos(x);
  if (!strcmp(op, "tan"))
    return tan(x);
}

double factorial(double x) {
  if (x < 2)
    return 1;
  double res = 1;
  for (int i = (int)x; i > 1; i--)
    res *= (double)i;
  return res;
}
