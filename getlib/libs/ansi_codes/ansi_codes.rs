#![allow(dead_code)]
#![allow(non_snake_case)]
pub const CUR_UP: &str = "\x1b[1A";
pub const CUR_DOWN: &str = "\x1b[1B";
pub const CUR_RIGHT: &str = "\x1b[1C";
pub const CUR_LEFT: &str = "\x1b[1D";

pub fn CUR_UPN(N: i32) -> String {
    format!("\x1b[{}A", N)
}
pub fn CUR_DOWNN(N: i32) -> String {
    format!("\x1b[{}B", N)
}
pub fn CUR_RIGHTN(N: i32) -> String {
    format!("\x1b[{}C", N)
}
pub fn CUR_LEFTN(N: i32) -> String {
    format!("\x1b[{}D", N)
}

// CUR_{direction}_LINE moves cursor to the beg of line N lines up/down
pub const CUR_UP_LINE: &str = "\x1b[1E";
pub const CUR_DOWN_LINE: &str = "\x1b[1F";

pub fn CUR_UP_LINEN(N: i32) -> String {
    format!("\x1b[{}E", N)
}
pub fn CUR_DOWN_LINEN(N: i32) -> String {
    format!("\x1b[{}F", N)
}

pub fn CUR_TO_COLN(N: i32) -> String {
    format!("\x1b[{}G", N)
}

// Moves cursor to home position (0, 0)
pub const CUR_TO_HOME: &str = "\x1b[H";

// Figure out difference between "H" and "f"
// Move cursor to line Y, col X
// "\x1b[{Y};{X}H"
// "\x1b[{Y};{x}f"

// TODO: Figure out difference between "\x1b[J" and "\x1b[2J"
pub const CLEAR_SCREEN_RIGHT: &str = "\x1b[0J";
pub const CLEAR_SCREEN_LEFT: &str = "\x1b[1J";
pub const CLEAR_SCREEN: &str = "\x1b[2J";

// TODO: Figure out difference between "\x1b[K" and "\x1b[2K"
pub const CLEAR_LINE_RIGHT: &str = "\x1b[0K";
pub const CLEAR_LINE_LEFT: &str = "\x1b[1K";
pub const CLEAR_LINE: &str = "\x1b[2K";

pub const RESET_GRAPHICS: &str = "\x1b[0m";
pub const SET_BOLD: &str = "\x1b[1m";
pub const SET_DIM: &str = "\x1b[2m";
pub const SET_ITALIC: &str = "\x1b[3m";
pub const SET_UNDERLINE: &str = "\x1b[4m";
pub const SET_BLINKING: &str = "\x1b[5m";
pub const SET_REVERSE: &str = "\x1b[7m";
pub const SET_INVISIBLE: &str = "\x1b[8m";
pub const SET_STRIKETHROUGH: &str = "\x1b[9m";

// FORE = foreground color; BACK = background color
pub const FORE_BLACK: i32 = 30;
pub const FORE_RED: i32 = 31;
pub const FORE_GREEN: i32 = 32;
pub const FORE_YELLOW: i32 = 33;
pub const FORE_BLUE: i32 = 34;
pub const FORE_MAGENTA: i32 = 35;
pub const FORE_CYAN: i32 = 36;
pub const FORE_WHITE: i32 = 37;
pub const BACK_BLACK: i32 = 40;
pub const BACK_RED: i32 = 41;
pub const BACK_GREEN: i32 = 42;
pub const BACK_YELLOW: i32 = 43;
pub const BACK_BLUE: i32 = 44;
pub const BACK_MAGENTA: i32 = 45;
pub const BACK_CYAN: i32 = 46;
pub const BACK_WHITE: i32 = 47;

pub fn SET_GRAPHICS(COLOR: i32) -> String {
    format!("\x1b[{}m", COLOR)
}
pub fn SET_FORE_COLOR_ID(ID: i32) -> String {
    format!("\x1b[38;5;${}m", ID)
}
pub fn SET_BACK_COLOR_ID(ID: i32) -> String {
    format!("\x1b[48;5;${}m", ID)
}
pub fn SET_FORE_COLOR_RGB(R: i32, G: i32, B: i32) -> String {
    format!("\x1b[38;2;{};{};{}m", R, G, B)
}
pub fn SET_BACK_COLOR_RGB(R: i32, G: i32, B: i32) -> String {
    format!("\x1b[48;2;{};{};{}m", R, G, B)
}

pub const CUR_INVISIBLE: &str = "\x1b[?25l";
pub const CUR_VISIBLE: &str = "\x1b[?25h";
