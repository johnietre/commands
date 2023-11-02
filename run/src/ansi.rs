#![allow(dead_code)]
#![allow(non_snake_case)]
use std::fmt;
use std::ops::Add;

// Moves cursor to home position (0, 0)
pub const CUR_TO_HOME: &str = "\x1b[H";
pub const CUR_UP: &str = "\x1b[1A";
pub const CUR_DOWN: &str = "\x1b[1B";
pub const CUR_RIGHT: &str = "\x1b[1C";
pub const CUR_LEFT: &str = "\x1b[1D";
pub const CUR_SAVE: &str = "\x1b 7";
pub const CUR_RESTORE: &str = "\x1b 8";
pub const CUR_POS: &str = "\x1b[6n";

pub fn CUR_UPN(N: usize) -> String {
    format!("\x1b[{}A", N)
}
pub fn CUR_DOWNN(N: usize) -> String {
    format!("\x1b[{}B", N)
}
pub fn CUR_RIGHTN(N: usize) -> String {
    format!("\x1b[{}C", N)
}
pub fn CUR_LEFTN(N: usize) -> String {
    format!("\x1b[{}D", N)
}

// CUR_{direction}_LINE moves cursor to the beg of line N lines up/down
pub const CUR_DOWN_LINE: &str = "\x1b[1E";
pub const CUR_UP_LINE: &str = "\x1b[1F";

pub fn CUR_DOWN_LINEN(n: usize) -> String {
    format!("\x1b[{}E", n)
}
pub fn CUR_UP_LINEN(n: usize) -> String {
    format!("\x1b[{}F", n)
}

pub fn CUR_TO_COLN(n: usize) -> String {
    format!("\x1b[{}G", n)
}
pub fn CUR_TO_POS(x: usize, y: usize) -> String {
    format!("\x1b[{};{}H", y, x)
}

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

pub const RESET_ALL: &str = "\x1b[0m";

pub const SET_BOLD: &str = "\x1b[1m";
pub const SET_DIM: &str = "\x1b[2m";
pub const SET_ITALIC: &str = "\x1b[3m";
pub const SET_UNDERLINE: &str = "\x1b[4m";
pub const SET_BLINKING: &str = "\x1b[5m";
pub const SET_INVERSE: &str = "\x1b[7m";
pub const SET_INVISIBLE: &str = "\x1b[8m";
pub const SET_STRIKETHROUGH: &str = "\x1b[9m";

pub const RESET_BOLD: &str = "\x1b[22m";
pub const RESET_DIM: &str = "\x1b[22m";
pub const RESET_ITALIC: &str = "\x1b[23m";
pub const RESET_UNDERLINE: &str = "\x1b[24m";
pub const RESET_BLINKING: &str = "\x1b[25m";
pub const RESET_INVERSE: &str = "\x1b[27m";
pub const RESET_INVISIBLE: &str = "\x1b[28m";
pub const RESET_STRIKETHROUGH: &str = "\x1b[29m";

#[repr(u8)]
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum GraphicsMode {
    Reset(u8) = 0,
    Bold = 1,
    Dim = 2,
    Italic = 3,
    Underline = 4,
    Blinking = 5,
    Inverse = 7,
    Invisible = 8,
    Strikethrough = 9,
}

impl GraphicsMode {
    pub fn as_u8(self) -> u8 {
        use self::GraphicsMode::*;
        match self {
            Bold => 1,
            Dim => 2,
            Italic => 3,
            Underline => 4,
            Blinking => 5,
            Inverse => 7,
            Invisible => 8,
            Strikethrough => 9,
            Reset(code) => code,
        }
    }

    pub const fn reset_code(self) -> u8 {
        use self::GraphicsMode::*;
        match self {
            Bold | Dim => 22,
            Italic => 23,
            Underline => 24,
            Blinking => 25,
            Inverse => 27,
            Invisible => 28,
            Strikethrough => 29,
            Reset(code) => code,
        }
    }

    pub const fn reset(self) -> Self {
        Self::Reset(self.reset_code())
    }

    /*
    pub const fn const_str(self) -> &'static str {
        use self::GraphicsMode::*;
        match self {
            Bold => RESET_BOLD,
            Dim => RESET_DIM,
            Italic => RESET_ITALIC,
            Underline => RESET_UNDERLINE,
            Blinking => RESET_BLINKING,
            Inverse => RESET_INVERSE,
            Invisible => RESET_INVISIBLE,
            Strikethrough => RESET_STRIKETHROUGH,
            /*
            Reset(code) if code == Bold.reset_code() => RESET_BOLD,
            Reset(code) if code == Dim.reset_code() => RESET_DIM,
            Reset(code) if code == Italic.reset_code() => RESET_ITALIC,
            Reset(code) if code == Underline.reset_code() => RESET_UNDERLINE,
            Reset(code) if code == Blinking.reset_code() => RESET_BLINKING,
            Reset(code) if code == Inverse.reset_code() => RESET_INVERSE,
            Reset(code) if code == Invisible.reset_code() => RESET_INVISIBLE,
            Reset(code) if code == Strikethrough.reset_code() => RESET_STRIKETHROUGH,
            Reset(_) => panic!("unknown reset code"),
            */
        }
    }
    */
}

impl fmt::Display for GraphicsMode {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        //write!(f, "{}", self.const_str())
        //write!(f, "\x1b[{}m", *self as u8)
        write!(f, "\x1b[{}m", self.as_u8())
    }
}

// FORE = foreground color; BACK = background color
pub const FORE_BLACK: &str = "\x1b[30m";
pub const FORE_RED: &str = "\x1b[31m";
pub const FORE_GREEN: &str = "\x1b[32m";
pub const FORE_YELLOW: &str = "\x1b[33m";
pub const FORE_BLUE: &str = "\x1b[34m";
pub const FORE_MAGENTA: &str = "\x1b[35m";
pub const FORE_CYAN: &str = "\x1b[36m";
pub const FORE_WHITE: &str = "\x1b[37m";
pub const FORE_DEFAULT: &str = "\x1b[39m";

pub const BACK_BLACK: &str = "\x1b[40m";
pub const BACK_RED: &str = "\x1b[41m";
pub const BACK_GREEN: &str = "\x1b[42m";
pub const BACK_YELLOW: &str = "\x1b[43m";
pub const BACK_BLUE: &str = "\x1b[44m";
pub const BACK_MAGENTA: &str = "\x1b[45m";
pub const BACK_CYAN: &str = "\x1b[46m";
pub const BACK_WHITE: &str = "\x1b[47m";
pub const BACK_DEFAULT: &str = "\x1b[49m";

#[repr(u8)]
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ForeColor {
    RGB(u8, u8, u8),
    Black = 30,
    Red = 31,
    Green = 32,
    Yellow = 33,
    Blue = 34,
    Magenta = 35,
    Cyan = 36,
    White = 37,
    Default = 39,
}

impl ForeColor {
    pub fn as_back(self) -> BackColor {
        use self::ForeColor::*;
        match self {
            Black => BackColor::Black,
            Red => BackColor::Red,
            Green => BackColor::Green,
            Yellow => BackColor::Yellow,
            Blue => BackColor::Blue,
            Magenta => BackColor::Magenta,
            Cyan => BackColor::Cyan,
            White => BackColor::White,
            Default => BackColor::Default,
            RGB(r, g, b) => BackColor::RGB(r, g, b),
        }
    }

    pub fn as_u8(self) -> u8 {
        use self::ForeColor::*;
        match self {
            Black => 30,
            Red => 31,
            Green => 32,
            Yellow => 33,
            Blue => 34,
            Magenta => 35,
            Cyan => 36,
            White => 37,
            Default => 39,
            RGB(_, _, _) => 39,
        }
    }
    /*
    pub const fn const_str(self) -> &'static str {
        use self::ForeColor::*;
        match self {
            Black => FORE_BLACK,
            Red => FORE_RED,
            Green => FORE_GREEN,
            Yellow => FORE_YELLOW,
            Blue => FORE_BLUE,
            Magenta => FORE_MAGENTA,
            Cyan => FORE_CYAN,
            White => FORE_WHITE,
            Default => FORE_DEFAULT,
        }
    }
    */
}

impl fmt::Display for ForeColor {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        //write!(f, "{}", self.const_str())
        if let ForeColor::RGB(r, b, g) = self {
            write!(f, "\x1b[38;2;{};{};{}m", r, g, b)
        } else {
            write!(f, "\x1b[{}m", self.as_u8())
        }
    }
}

#[repr(u8)]
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum BackColor {
    RGB(u8, u8, u8),
    Black = 40,
    Red = 41,
    Green = 42,
    Yellow = 43,
    Blue = 44,
    Magenta = 45,
    Cyan = 46,
    White = 47,
    Default = 49,
}

impl BackColor {
    pub fn as_fore(self) -> ForeColor {
        use self::BackColor::*;
        match self {
            Black => ForeColor::Black,
            Red => ForeColor::Red,
            Green => ForeColor::Green,
            Yellow => ForeColor::Yellow,
            Blue => ForeColor::Blue,
            Magenta => ForeColor::Magenta,
            Cyan => ForeColor::Cyan,
            White => ForeColor::White,
            Default => ForeColor::Default,
            RGB(r, g, b) => ForeColor::RGB(r, g, b),
        }
    }

    pub fn as_u8(self) -> u8 {
        use self::BackColor::*;
        match self {
            Black => 40,
            Red => 41,
            Green => 42,
            Yellow => 43,
            Blue => 44,
            Magenta => 45,
            Cyan => 46,
            White => 47,
            Default => 49,
            RGB(_, _, _) => 49,
        }
    }
    /*
    pub const fn const_str(self) -> &'static str {
        use self::BackColor::*;
        match self {
            Black => BACK_BLACK,
            Red => BACK_RED,
            Green => BACK_GREEN,
            Yellow => BACK_YELLOW,
            Blue => BACK_BLUE,
            Magenta => BACK_MAGENTA,
            Cyan => BACK_CYAN,
            White => BACK_WHITE,
            Default => BACK_DEFAULT,
        }
    }
    */
}

impl fmt::Display for BackColor {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        //write!(f, "{}", self.const_str())
        if let BackColor::RGB(r, b, g) = self {
            write!(f, "\x1b[48;2;{};{};{}m", r, g, b)
        } else {
            write!(f, "\x1b[{}m", self.as_u8())
        }
    }
}

pub fn SET_GRAPHICS(graphics: GraphicsMode) -> String {
    format!("\x1b[{}m", graphics.as_u8())
}

pub fn WRAP_GRAPHICS(text: &str, graphics: GraphicsMode) -> String {
    format!(
        "\x1b[{}m{}{}",
        graphics.as_u8(),
        text,
        graphics.reset_code()
    )
}

pub fn SET_FORE_COLOR_ID(color: ForeColor) -> String {
    format!("\x1b[{}m", color.as_u8())
}

pub fn WRAP_FORE_COLOR_ID(text: &str, color: ForeColor) -> String {
    format!("\x1b[{}m{}{}", color.as_u8(), text, FORE_DEFAULT)
}

pub fn SET_BACK_COLOR_ID(color: BackColor) -> String {
    format!("\x1b[{}m", color.as_u8())
}

pub fn WRAP_BACK_COLOR_ID(text: &str, color: BackColor) -> String {
    format!("\x1b[{}m{}{}", color.as_u8(), text, BACK_DEFAULT)
}

pub fn SET_FORE_COLOR_RGB((r, g, b): (u8, u8, u8)) -> String {
    format!("\x1b[38;2;{};{};{}m", r, g, b)
}

pub fn WRAP_FORE_COLOR_RGB(text: &str, (r, g, b): (u8, u8, u8)) -> String {
    format!("\x1b[38;2;{};{};{}m{}{}", r, g, b, text, FORE_DEFAULT)
}

pub fn SET_BACK_COLOR_RGB((r, g, b): (u8, u8, u8)) -> String {
    format!("\x1b[48;2;{};{};{}m", r, g, b)
}

pub fn WRAP_BACK_COLOR_RGB(text: &str, (r, g, b): (u8, u8, u8)) -> String {
    format!("\x1b[48;2;{};{};{}m{}{}", r, g, b, text, BACK_DEFAULT)
}

#[derive(Clone, Debug, PartialEq, Eq)]
enum TextNode {
    Text(String),
    Fore(ForeColor),
    Back(BackColor),
    Graphics(GraphicsMode),
}

impl fmt::Display for TextNode {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            TextNode::Text(text) => write!(f, "{}", text),
            TextNode::Fore(color) => write!(f, "{}", color),
            TextNode::Back(color) => write!(f, "{}", color),
            TextNode::Graphics(graphics) => write!(f, "{}", graphics),
        }
    }
}

#[derive(Default)]
pub struct Text {
    nodes: Vec<TextNode>,
}

impl Text {
    pub fn new() -> Self {
        Default::default()
    }

    pub fn text(&mut self, text: impl ToString) -> &mut Self {
        self.nodes.push(TextNode::Text(text.to_string()));
        self
    }

    pub fn graphics(&mut self, graphics: GraphicsMode) -> &mut Self {
        self.nodes.push(TextNode::Graphics(graphics));
        self
    }

    pub fn wrap_graphics(&mut self, text: impl ToString, graphics: GraphicsMode) -> &mut Self {
        self.nodes.extend_from_slice(&[
            TextNode::Graphics(graphics),
            TextNode::Text(text.to_string()),
            TextNode::Graphics(graphics.reset()),
        ]);
        self
    }

    pub fn reset_graphics(&mut self, graphics: GraphicsMode) -> &mut Self {
        self.nodes.push(TextNode::Graphics(graphics.reset()));
        self
    }

    pub fn fore(&mut self, color: ForeColor) -> &mut Self {
        self.nodes.push(TextNode::Fore(color));
        self
    }

    pub fn wrap_fore(&mut self, text: impl ToString, color: ForeColor) -> &mut Self {
        self.nodes.extend_from_slice(&[
            TextNode::Fore(color),
            TextNode::Text(text.to_string()),
            TextNode::Fore(ForeColor::Default),
        ]);
        self
    }

    pub fn reset_fore(&mut self) -> &mut Self {
        self.nodes.push(TextNode::Fore(ForeColor::Default));
        self
    }

    pub fn back(&mut self, color: BackColor) -> &mut Self {
        self.nodes.push(TextNode::Back(color));
        self
    }

    pub fn wrap_back(&mut self, text: impl ToString, color: BackColor) -> &mut Self {
        self.nodes.extend_from_slice(&[
            TextNode::Back(color),
            TextNode::Text(text.to_string()),
            TextNode::Back(BackColor::Default),
        ]);
        self
    }

    pub fn reset_back(&mut self) -> &mut Self {
        self.nodes.push(TextNode::Back(BackColor::Default));
        self
    }

    pub fn reset(&mut self) {
        self.nodes.clear();
    }
}

impl Add for Text {
    type Output = Text;

    fn add(mut self, rhs: Text) -> Self::Output {
        self.nodes.extend(rhs.nodes);
        self
    }
}

impl Add<GraphicsMode> for Text {
    type Output = Text;

    fn add(mut self, rhs: GraphicsMode) -> Self::Output {
        self.graphics(rhs);
        self
    }
}

impl Add<ForeColor> for Text {
    type Output = Text;

    fn add(mut self, rhs: ForeColor) -> Self::Output {
        self.fore(rhs);
        self
    }
}

impl Add<BackColor> for Text {
    type Output = Text;

    fn add(mut self, rhs: BackColor) -> Self::Output {
        self.back(rhs);
        self
    }
}

impl fmt::Display for Text {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        self.nodes.iter().try_for_each(|node| write!(f, "{}", node))
    }
}

#[derive(Default)]
pub struct TextBuilder {
    text: Text,
}

impl TextBuilder {
    pub fn new() -> Self {
        Default::default()
    }

    pub fn text(mut self, text: impl ToString) -> Self {
        self.text.nodes.push(TextNode::Text(text.to_string()));
        self
    }

    pub fn graphics(mut self, graphics: GraphicsMode) -> Self {
        self.text.nodes.push(TextNode::Graphics(graphics));
        self
    }

    pub fn wrap_graphics(mut self, text: impl ToString, graphics: GraphicsMode) -> Self {
        self.text.nodes.extend_from_slice(&[
            TextNode::Graphics(graphics),
            TextNode::Text(text.to_string()),
            TextNode::Graphics(graphics.reset()),
        ]);
        self
    }

    pub fn reset_graphics(mut self, graphics: GraphicsMode) -> Self {
        self.text.nodes.push(TextNode::Graphics(graphics.reset()));
        self
    }

    pub fn fore(mut self, color: ForeColor) -> Self {
        self.text.nodes.push(TextNode::Fore(color));
        self
    }

    pub fn wrap_fore(mut self, text: impl ToString, color: ForeColor) -> Self {
        self.text.nodes.extend_from_slice(&[
            TextNode::Fore(color),
            TextNode::Text(text.to_string()),
            TextNode::Fore(ForeColor::Default),
        ]);
        self
    }

    pub fn reset_fore(mut self) -> Self {
        self.text.nodes.push(TextNode::Fore(ForeColor::Default));
        self
    }

    pub fn back(mut self, color: BackColor) -> Self {
        self.text.nodes.push(TextNode::Back(color));
        self
    }

    pub fn wrap_back(mut self, text: impl ToString, color: BackColor) -> Self {
        self.text.nodes.extend_from_slice(&[
            TextNode::Back(color),
            TextNode::Text(text.to_string()),
            TextNode::Back(BackColor::Default),
        ]);
        self
    }

    pub fn reset_back(mut self) -> Self {
        self.text.nodes.push(TextNode::Back(BackColor::Default));
        self
    }

    pub fn build(self) -> Text {
        self.text
    }
}

pub const CUR_INVISIBLE: &str = "\x1b[?25l";
pub const CUR_VISIBLE: &str = "\x1b[?25h";
pub const SCREEN_RESTORE: &str = "\x1b[?47l";
pub const SCREEN_SAVE: &str = "\x1b[?47h";
