#ifndef ANSI_CODES_H
#define ANSI_CODES_H

#define CUR_UP "\x1b[1A"
#define CUR_DOWN "\x1b[1B"
#define CUR_RIGHT "\x1b[1C"
#define CUR_LEFT "\x1b[1D"

#define CUR_UPN(N) "\x1b[" #N "A"
#define CUR_DOWNN(N) "\x1b[" #N "B"
#define CUR_RIGHTN(N) "\x1b[" #N "C"
#define CUR_LEFTN(N) "\x1b[" #N "D"

// CUR_{direction}_LINE moves cursor to the beg of line N lines up/down
#define CUR_UP_LINE "\x1b[1E"
#define CUR_DOWN_LINE "\x1b[1F"

#define CUR_UP_LINEN(N) "\x1b[" #N "E"
#define CUR_DOWN_LINEN(N) "\x1b[" #N "F"

#define CUR_TO_COLN(N) "\x1b[" #N "G"

// Moves cursor to home position (0, 0)
#define CUR_TO_HOME "\x1b[H"

#if 0
// Figure out difference between "H" and "f"
// Move cursor to line Y, col X
// "\x1b[{Y};{X}H"
// "\x1b[{Y};{x}f"
#endif

// TODO: Figure out difference between "\x1b[J" and "\x1b[2J"
#define CLEAR_SCREEN_RIGHT "\x1b[0J"
#define CLEAR_SCREEN_LEFT "\x1b[1J"
#define CLEAR_SCREEN "\x1b[2J"

// TODO: Figure out difference between "\x1b[K" and "\x1b[2K"
#define CLEAR_LINE_RIGHT "\x1b[0K"
#define CLEAR_LINE_LEFT "\x1b[1K"
#define CLEAR_LINE "\x1b[2K"

#define RESET_GRAPHICS "\x1b[0m"
#define SET_BOLD "\x1b[1m"
#define SET_DIM "\x1b[2m"
#define SET_ITALIC "\x1b[3m"
#define SET_UNDERLINE "\x1b[4m"
#define SET_BLINKING "\x1b[5m"
#define SET_REVERSE "\x1b[7m"
#define SET_INVISIBLE "\x1b[8m"
#define SET_STRIKETHROUGH "\x1b[9m"

// FORE = foreground color; BACK = background color
#define FORE_BLACK 30
#define FORE_RED 31
#define FORE_GREEN 32
#define FORE_YELLOW 33
#define FORE_BLUE 34
#define FORE_MAGENTA 35
#define FORE_CYAN 36
#define FORE_WHITE 37
#define BACK_BLACK 40
#define BACK_RED 41
#define BACK_GREEN 42
#define BACK_YELLOW 43
#define BACK_BLUE 44
#define BACK_MAGENTA 45
#define BACK_CYAN 46
#define BACK_WHITE 47

#define SET_GRAPHICS(COLOR) "\x1b[" #COLOR "m"
#define SET_FORE_COLOR_ID(ID) "\x1b[38;5;$" #ID "m"
#define SET_BACK_COLOR_ID(ID) "\x1b[48;5;$" #ID "m"
#define SET_FORE_COLOR_RGB(R,G,B) "\x1b[38;2;" #R ";" #G ";" #B "m"
#define SET_BACK_COLOR_RGB(R,G,B) "\x1b[48;2;" #R ":" #G ";" #B "m"

#define CUR_INVISIBLE "\x1b[?25l"
#define CUR_VISIBLE "\x1b[?25h"

#endif

