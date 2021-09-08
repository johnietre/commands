CUR_UP = "\x1b[1A"
CUR_DOWN = "\x1b[1B"
CUR_RIGHT = "\x1b[1C"
CUR_LEFT = "\x1b[1D"
CUR_UPN = lambda N: "\x1b[" + str(N) + "A"
CUR_DOWNN = lambda N: "\x1b[" + str(N) + "B"
CUR_RIGHTN = lambda N: "\x1b[" + str(N) + "C"
CUR_LEFTN = lambda N: "\x1b[" + str(N) + "D"

# CUR_{direction}_LINE moves cursor to the beg of line N lines up/down
CUR_UP_LINE = "\x1b[1E"
CUR_DOWN_LINE = "\x1b[1F"
CUR_UP_LINEN = lambda N: "\x1b[" + str(N) + "E"
CUR_DOWN_LINEN = lambda N: "\x1b[" + str(N) + "F"
CUR_TO_COLN = lambda N: "\x1b[" + str(N) + "G"

# Moves cursor to home position (0, 0)
CUR_TO_HOME = "\x1b[H"

# Figure out difference between "H" and "f"
# Move cursor to line Y, col X
# "\x1b[{Y};{X}H"
# "\x1b[{Y};{x}f"

# TODO: Figure out difference between "\x1b[J" and "\x1b[2J"
CLEAR_SCREEN_RIGHT = "\x1b[0J"
CLEAR_SCREEN_LEFT = "\x1b[1J"
CLEAR_SCREEN = "\x1b[2J"

# TODO: Figure out difference between "\x1b[K" and "\x1b[2K"
CLEAR_LINE_RIGHT = "\x1b[0K"
CLEAR_LINE_LEFT = "\x1b[1K"
CLEAR_LINE = "\x1b[2K"

RESET_GRAPHICS = "\x1b[0m"
SET_BOLD = "\x1b[1m"
SET_DIM = "\x1b[2m"
SET_ITALIC = "\x1b[3m"
SET_UNDERLINE = "\x1b[4m"
SET_BLINKING = "\x1b[5m"
SET_REVERSE = "\x1b[7m"
SET_INVISIBLE = "\x1b[8m"
SET_STRIKETHROUGH = "\x1b[9m"

# FORE = foreground color; BACK = background color
FORE_BLACK = 30
FORE_RED = 31
FORE_GREEN = 32
FORE_YELLOW = 33
FORE_BLUE = 34
FORE_MAGENTA = 35
FORE_CYAN = 36
FORE_WHITE = 37
BACK_BLACK = 40
BACK_RED = 41
BACK_GREEN = 42
BACK_YELLOW = 43
BACK_BLUE = 44
BACK_MAGENTA = 45
BACK_CYAN = 46
BACK_WHITE = 47

SET_GRAPHICS = lambda COLOR: "\x1b[" + str(COLOR) + "m"
SET_FORE_COLOR_ID = lambda ID: "\x1b[38;5;$" + str(ID) + "m"
SET_BACK_COLOR_ID = lambda ID: "\x1b[48;5;$" + str(ID) + "m"
SET_FORE_COLOR_RGB = lambda R,G,B: "\x1b[38;2;" + str(R) + ";" + str(G) + ";" + str(B) + "m"
SET_BACK_COLOR_RGB = lambda R,G,B: "\x1b[48;2;" + str(R) + ";" + str(G) + ";" + str(B) + "m"

CUR_INVISIBLE = "\x1b[?25l"
CUR_VISIBLE = "\x1b[?25h"
