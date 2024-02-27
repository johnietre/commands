// TODO: Handle UTF8 characters

#![allow(dead_code)]
use crate::ansi::{self, *};
use crate::deferrer::*;
use std::fs::File;
use std::io::{prelude::*, stdin, stdout, StdoutLock};
use std::ops::{Add, Sub};
use std::os::fd::{AsRawFd, RawFd};

const BUF_SIZE: usize = 12;

pub fn run_app(_: &crate::args::Args) -> i32 {
    // TODO: Get terminal dims
    let mut app = App::new(80, 160);
    unsafe {
        app.run();
    }
    app.status.unwrap_or(0)
}

struct App {
    term: Term,
    stdin_fd: RawFd,
    buf: [u8; BUF_SIZE],
    status: Option<i32>,
    log_file: Option<File>,
}

impl App {
    fn new(width: usize, height: usize) -> Self {
        Self {
            term: Term::new(width, height),
            stdin_fd: stdin().as_raw_fd(),
            buf: [0u8; BUF_SIZE],
            status: None,
            log_file: cfg!(feature = "debug").then(|| {
                File::options()
                    .append(true)
                    .create(true)
                    .open("app.log")
                    .unwrap()
            }),
        }
    }

    fn log(&mut self, msg: &str) {
        if cfg!(feature = "debug") {
            let _ = write!(self.log_file.as_mut().unwrap(), "{}\n", msg);
        }
    }

    unsafe fn run(&mut self) {
        let mut deferrer = Deferrer::new();
        if !set_term(self.stdin_fd, &mut deferrer) {
            self.status = Some(1);
            return;
        }
        deferrer.push_back(Box::new(|| {
            print!("{}{}", ansi::CLEAR_SCREEN, ansi::CUR_TO_HOME);
            println!("DONE");
            let _ = stdout().flush();
        }));
        let _ = write!(self.term, "{}{}", ansi::CLEAR_SCREEN, ansi::CUR_TO_HOME);
        let _ = self.term.flush();

        while self.is_running() {
            let n = libc::read(self.stdin_fd, self.buf.as_mut_ptr().cast(), BUF_SIZE as _);
            if n < 0 {
                eprintln!("error reading from stdin, quitting");
                self.status = Some(1);
                break;
            } else if n == 0 {
                continue;
            }
            self.parse_input(n as usize);
            // Parse the input
            /*
            if n == -1 {
                eprintln!("error reading from stdin, quitting");
                return;
            } else if n == 0 {
                continue;
            } else if n == 1 {
                match self.buf[0] {
                    b'q' => break,
                    b'n' => self.editing_mode(),
                    b'd' => self.delete_line_mode(),
                    _ => continue,
                }
            } else if n == 3 {
                match self.buf[0] {
                    0x1b => {
                        match self.buf[1..] {
                            [0x5b, 0x41] => self.term.move_pos_up(),
                            [0x5b, 0x42] => self.term.move_pos_down(),
                            [0x5b, 0x43] => self.term.move_pos_right(),
                            [0x5b, 0x44] => self.term.move_pos_left(),
                            _ => continue,
                        }
                        let _ = self.term.write(&[7]);
                        let _ = self.term.flush();
                    }
                    _ => continue,
                }
            }
            */
        }
    }

    unsafe fn parse_input(&mut self, n: usize) {
        match Input::parse(&self.buf[..n]) {
            Input::Char('q') => self.status = Some(0),
            Input::Char('n') => self.editing_mode(),
            Input::Char('d') => self.delete_line_mode(),
            Input::ArrowUp => {
                self.term.move_pos_up();
                let _ = self.term.flush();
            }
            Input::ArrowDown => {
                self.term.move_pos_down();
                let _ = self.term.flush();
            }
            Input::ArrowRight => {
                self.term.move_pos_right();
                let _ = self.term.flush();
            }
            Input::ArrowLeft => {
                self.term.move_pos_left();
                let _ = self.term.flush();
            }
            Input::Backspace => (),
            Input::Del => (),
            Input::Char(_) | Input::Unknown => (),
        }
    }

    #[inline(always)]
    pub fn is_running(&self) -> bool {
        self.status.is_none()
    }

    unsafe fn editing_mode(&mut self) {
        self.term.save_pos();
        self.term.pos = Pos::new(1, self.term.height);
        self.term.to_pos();
        let _ = self.term.flush();

        let mut line = String::new();
        while self.is_running() {
            let n = libc::read(self.stdin_fd, self.buf.as_mut_ptr().cast(), 3);
            if n < 0 {
                eprintln!("error reading from stdin, quitting");
                return;
            } else if n == 0 {
                continue;
            }
            let n = n as usize;
            let should_flush = match Input::parse(&self.buf[..n]) {
                Input::Char(c) if !c.is_control() => {
                    line.push(c);
                    self.term.add_to_bottom(c);
                    true
                }
                Input::Char('\x0a') => {
                    self.term.add_line(line);
                    break;
                }
                Input::Char('\x1b') => break,
                Input::Backspace => {
                    self.term.bottom_backspace();
                    true
                }
                Input::Del => {
                    self.term.bottom_del();
                    true
                }
                Input::ArrowUp => {
                    self.term.bottom_to_beg();
                    true
                }
                Input::ArrowDown => {
                    self.term.bottom_to_end();
                    true
                }
                Input::ArrowRight => {
                    self.term.bottom_right();
                    true
                }
                Input::ArrowLeft => {
                    self.term.bottom_left();
                    true
                }
                _ => false,
            };
            if should_flush {
                let _ = self.term.flush();
            }
        }
        self.term.clear_bottom();
        self.term.restore_pos();
        let _ = self.term.flush();
    }

    unsafe fn delete_line_mode(&mut self) {
        loop {
            let n = libc::read(self.stdin_fd, self.buf.as_mut_ptr().cast(), 3);
            if n == -1 {
                eprintln!("error reading from stdin, quitting");
                return;
            } else if n == 0 {
                continue;
            } else if n == 1 {
                match self.buf[0] {
                    b'd' => {
                        self.term.remove_line_at_pos();
                        let _ = self.term.flush();
                    }
                    _ => break,
                }
            }
        }
    }
}

struct Term {
    lines: Vec<Line>,
    lines_start: usize,
    bottom_line: Line,

    height: usize,
    width: usize,
    pos: Pos,
    prev_pos: Pos,
    //max_pos: Pos,
    stdout_buf: Vec<u8>,
    stdout: StdoutLock<'static>,

    log_file: Option<File>,
}

impl Term {
    fn new(width: usize, height: usize) -> Self {
        Self {
            lines: Vec::new(),
            lines_start: 0,
            bottom_line: Line::default(),
            width,
            height,
            pos: Pos::new(1, 1),
            prev_pos: Pos::new(1, 1),
            stdout_buf: Vec::with_capacity(1024),
            stdout: stdout().lock(),
            log_file: cfg!(feature = "debug").then(|| {
                File::options()
                    .append(true)
                    .create(true)
                    .open("term.log")
                    .unwrap()
            }),
        }
    }

    fn log(&mut self, msg: &str) {
        if cfg!(feature = "debug") {
            let _ = write!(self.log_file.as_mut().unwrap(), "{}\n", msg);
        }
    }

    fn move_pos_up(&mut self) {
        if self.pos.y == 1 {
            return;
        }
        self.pos.y -= 1;
        let _ = self.write(ansi::CUR_UP.as_bytes());
    }

    fn move_pos_down(&mut self) {
        if self.pos.y == self.lines.len().min(self.height - 1) {
            return;
        }
        self.pos.y += 1;
        let _ = self.write(ansi::CUR_DOWN.as_bytes());
    }

    fn move_pos_left(&mut self) {
        /*
        if self.pos.x != 1 {
            return;
        }
        self.pos.x -= 1;
        let _ = self.write(ansi::CUR_LEFT.as_bytes());
        */
        let line = &mut self.lines[self.pos.y - 1];
        // Check to see if the line is at the beginning, if not, shift left
        if line.start != 0 {
            line.start -= 1;
            let line = line.fit(self.width);
            let _ = self.write(line.as_bytes());
            let _ = self.write(CUR_TO_POS(self.pos.x, self.pos.y).as_bytes());
            //self.restore_pos();
        }
    }

    fn move_pos_right(&mut self) {
        /*
        if self.pos.x == self.lines[self.pos.y - 1].len().min(self.width) {
            return;
        }
        self.pos.x += 1;
        let _ = self.write(ansi::CUR_RIGHT.as_bytes());
        */
        let line = &mut self.lines[self.pos.y - 1];
        // Check to see if the line is at the end, if not, shift right
        if line.has_more(self.width) {
            line.start += 1;
            let line = line.fit(self.width);
            let _ = self.write(line.as_bytes());
            let _ = self.write(CUR_TO_POS(self.pos.x, self.pos.y).as_bytes());
            //self.restore_pos();
        }
    }

    fn save_pos(&mut self) {
        self.prev_pos = self.pos;
        self.log(&format!("saved {:?}", self.pos));
    }

    fn restore_pos(&mut self) {
        self.pos = self.prev_pos;
        // TODO: Don't do here and require call to to_pos?
        self.to_pos();
        self.log(&format!("restored {:?}", self.prev_pos));
        //let _ = self.write(ansi::CUR_TO_POS(self.pos.x, self.pos.y).as_bytes());
    }

    fn to_pos(&mut self) {
        let _ = self.write(ansi::CUR_TO_POS(self.pos.x, self.pos.y).as_bytes());
    }

    // Adds lines to buffer
    fn render_lines(&mut self) {
        self.render_lines_from(0);
    }

    // NOTE: lineno is 0-indexed
    fn render_lines_from(&mut self, lineno: usize) {
        let _ = self.write(ansi::CUR_TO_POS(0, lineno + 1).as_bytes());
        for lineno in lineno..self.lines.len().min(self.height - 1) {
            let _ = self.write(ansi::CLEAR_LINE.as_bytes());
            let line = &self.lines[lineno + self.lines_start];
            let _ = self.write(line.fit(self.width).as_bytes());
            let _ = self.write(ansi::CUR_DOWN_LINE.as_bytes());
        }
    }

    fn add_line(&mut self, s: impl ToString) {
        let line = Line::new(s.to_string());
        self.lines.push(line);
        if self.lines.len() <= self.height - 1 {
            let _ = write!(
                self,
                "{}{}",
                ansi::CUR_TO_POS(1, self.lines.len()),
                self.lines[self.lines.len() - 1].fit(self.width),
            );
            return;
        }
        self.lines_start += 1;
        self.render_lines();
    }

    fn remove_line_at_pos(&mut self) {
        //if self.pos.x > self.lines.len() || self.pos.x == 0 {
        if self.pos.x > self.lines.len() || self.pos.x == 1 {
            return;
        }
        let lineno = self.pos.x - 1;
        self.lines.remove(lineno);
        self.render_lines_from(lineno);
    }

    fn add_to_bottom(&mut self, c: char) {
        // TODO: Figure out where to rerender from
        self.bottom_line.contents.insert(self.pos.x - 1, c);
        // Check whether the cursor is at the end of the line or not
        if self.pos.x == self.width + 1 {
            // Check to see if the bottom line fits within the width
            if self.bottom_line.len() - self.bottom_line.start > self.width {
                self.bottom_line.start += 1;
                let _ = write!(
                    self.stdout_buf,
                    "{}{}",
                    CUR_TO_COLN(1),
                    self.bottom_line.fit(self.width),
                );
            } else {
                let _ = write!(
                    self.stdout_buf,
                    "{}{}",
                    CUR_TO_COLN(self.bottom_line.len()),
                    c,
                );
            }
        } else {
            let x1 = self.pos.x - 1;
            let _ = write!(
                self.stdout_buf,
                "{}{}",
                self.bottom_line.truncs(x1, self.width - x1),
                CUR_TO_COLN(self.pos.x + 1),
            );
        }
        self.pos.x += 1;
        //self.restore_pos();
    }

    fn pop_from_bottom(&mut self) {
        self.bottom_line.contents.pop();

        if self.bottom_line.start > 0 {
            self.bottom_line.start -= 1;
        }
        if self.bottom_line.len() == self.width {
            let _ = write!(
                self.stdout_buf,
                "{}{}",
                CUR_TO_POS(1, self.height),
                self.bottom_line.contents,
            );
        } else if self.bottom_line.len() - self.bottom_line.start == self.width {
            let _ = write!(
                self.stdout_buf,
                "{}{}",
                CUR_TO_POS(0, self.height),
                self.bottom_line.substr(self.width),
            );
        } else {
            let _ = self.write(CUR_TO_POS(self.bottom_line.len() + 1, self.height).as_bytes());
            self.stdout_buf.push(b' ');
        }
        //self.restore_pos();
    }

    fn clear_bottom(&mut self) {
        self.bottom_line = Line::new("");
        let _ = write!(
            self.stdout_buf,
            "{}{}",
            ansi::CUR_TO_POS(0, self.height),
            ansi::CLEAR_LINE,
        );
        self.restore_pos();
    }

    // Moves the cursor
    fn bottom_to_beg(&mut self) {
        if self.bottom_line.start != 0 {
            self.bottom_line.start = 0;
            let _ = write!(
                self.stdout_buf,
                "{}{}{}",
                CUR_TO_POS(1, self.height),
                self.bottom_line.trunc(self.width),
                CUR_TO_COLN(1),
            );
        }
    }

    fn bottom_to_end(&mut self) {
        if self.bottom_line.start != 0 {
            self.bottom_line.start = self.bottom_line.len().checked_sub(self.width).unwrap_or(0);
            let _ = write!(
                self.stdout_buf,
                "{}{}{}",
                CUR_TO_POS(1, self.height),
                self.bottom_line.fit(self.width),
                CUR_TO_COLN(self.width),
            );
        }
    }

    fn bottom_right(&mut self) {
        if self.pos.x == self.width + 1 {
            self.bottom_line.start += 1;
        }
    }

    fn bottom_left(&mut self) {
        if self.bottom_line.start == 0 {
            return;
        }
        if self.pos.x != 1 {
            self.pos.x -= 1;
            let _ = self.write(CUR_LEFT.as_bytes());
            return;
        }
        self.bottom_line.start -= 1;
        let _ = write!(
            self.stdout_buf,
            "{}{}",
            self.bottom_line.fit(self.width),
            CUR_TO_COLN(1),
        );
    }

    fn bottom_backspace(&mut self) {
        /*
        if self.bottom_line.len() == 0 {
            return;
        }
        if self.pos.x == 1 {
            self.bottom_line.remove(self.bottom_line.start);
            let _ = write!(
                self.stdout_buf,
                "{}{}",
                self.bottom_line.fit(self.width).as_bytes(),
                CUR_TO_COLN(1),
            );
            return;
        }
        self.pos.x -= 1;
        // Minus 2 since positions are 1-based and the cursor will be after the character
        self.bottom_line.remove(self.pos.x - 2);
        let _ = write!(
            self.stdout_buf,
            "{}{}",
            CUR_TO_COLN(self.pos.x),
            self.bottom_line,
        );
        */
    }

    fn bottom_del(&mut self) {
        /*
        if self.bottom_line.len() == 0 || self.pos.x == self.width + 1 {
            return;
        }
        // Minus 2 since positions are 1-based and the cursor will be after the character
        self.bottom_line.remove(self.pos.x - 2);
        */
    }

    fn main_menu_bar(&self) -> String {
        format!("h:Help",)
    }

    fn help_menu(&self) -> String {
        /*
        format!(
            "q:Quit"
            "h:Help"
            "r:Run"
            "R:Restart"
            "c:CTRL-C"
            "k:Kill",
        )
        */
        String::new()
    }
}

impl Write for Term {
    fn write(&mut self, buf: &[u8]) -> std::io::Result<usize> {
        self.stdout_buf.extend_from_slice(buf);
        Ok(buf.len())
    }

    fn flush(&mut self) -> std::io::Result<()> {
        self.stdout.write(&self.stdout_buf)?;
        self.stdout_buf.clear();
        self.stdout.flush()
    }
}

impl ToString for Term {
    fn to_string(&self) -> String {
        if self.height < 2 {
            return format!("{}{}", 1, 2,);
        }
        String::new()
    }
}

#[derive(Clone, Copy, Default, Debug, PartialEq, Eq)]
struct Pos {
    x: usize,
    y: usize,
}

impl Pos {
    fn new(x: usize, y: usize) -> Self {
        Pos { x, y }
    }
}

impl Add for Pos {
    type Output = Pos;

    fn add(self, rhs: Pos) -> Pos {
        Self {
            x: self.x + rhs.x,
            y: self.y + rhs.y,
        }
    }
}

impl Sub for Pos {
    type Output = Pos;

    fn sub(self, rhs: Pos) -> Pos {
        Self {
            x: self.x - rhs.x,
            y: self.y - rhs.y,
        }
    }
}

#[derive(Default, Debug)]
struct Line {
    contents: String,
    start: usize,
}

impl Line {
    fn new(s: impl ToString) -> Self {
        Self {
            contents: s.to_string(),
            start: 0,
        }
    }

    // Returns whether the line has more when starting at self.start and has a length of width
    fn has_more(&self, width: usize) -> bool {
        self.lens() > width
    }

    #[inline(always)]
    fn len(&self) -> usize {
        self.contents.len()
    }

    #[inline(always)]
    // Returns the length of the line when starting from self.start
    fn lens(&self) -> usize {
        self.contents.len().checked_sub(self.start).unwrap_or(0)
    }

    fn substr(&self, width: usize) -> &str {
        &self.contents[self.start..self.start + width]
        //&self.contents[self.start + 1..self.start + width]
    }

    fn fit(&self, width: usize) -> String {
        self.fits(self.start, width)
    }

    fn fits(&self, start: usize, width: usize) -> String {
        if start == 0 {
            if self.len() > width {
                self.trunc(width)
            } else {
                self.contents.clone()
            }
        } else {
            /*
            if self.len() - self.start <= width - 1 {
                format!("${}", &self.contents[start..])
            }
            */
            let (s1, w1) = (start + 1, width - 1);
            let mut text = TextBuilder::new()
                .wrap_back(&self.contents[start..s1], BackColor::Yellow)
                .build();
            // Check if the end of the string (from the start) is within the width
            if self.len() - self.start <= width {
                text.text(&self.contents[s1..]);
            } else {
                let sw1 = start + w1;
                text.text(&self.contents[s1..sw1])
                    .wrap_back(&self.contents[sw1..sw1 + 1], BackColor::Yellow);
            }
            text.to_string()
            /*
            if self.len() - self.start < width {
                format!("${}", &self.contents[start..])
            } else {
                format!("${}$", &self.contents[start + 1..start + width - 1])
            }
            */
        }
    }

    // Fits the contents into width starting from the beginning of the string
    fn trunc(&self, width: usize) -> String {
        if self.len() > width {
            let w1 = width - 1;
            Text::new()
                .text(&self.contents[..w1])
                .wrap_back(&self.contents[w1..width], BackColor::Yellow)
                .to_string()
        } else {
            self.contents.clone()
        }
    }

    // Fits the contents into width starting from the given beginning index
    fn truncs(&self, start: usize, width: usize) -> String {
        let contents = &self.contents[start..];
        if contents.len() > width {
            let w1 = width - 1;
            Text::new()
                .text(&contents[..w1])
                .wrap_back(&contents[w1..width], BackColor::Yellow)
                .to_string()
        } else {
            contents.to_string()
        }
    }

    #[inline(always)]
    fn remove(&mut self, idx: usize) {
        self.contents.remove(idx);
    }
}

// Returns true if successful
unsafe fn set_term(stdin_fd: RawFd, deferrer: &mut Deferrer) -> bool {
    // Get the old terminal and create the new
    let mut old_term = std::mem::MaybeUninit::<libc::termios>::zeroed().assume_init();
    if libc::tcgetattr(stdin_fd, (&mut old_term) as *mut _) != 0 {
        eprintln!("error getting terminal info");
        //std::process::exit(1);
        return false;
    }
    // Set the terminal
    let new_term = libc::termios {
        c_lflag: old_term.c_lflag & (!libc::ICANON & !libc::ECHO),
        c_cc: {
            let mut vals = old_term.c_cc;
            vals[libc::VMIN] = 0; // Minimum number of btyes required to be read
            vals[libc::VTIME] = 1; // Amount of time to wait before read returns
            vals
        },
        ..old_term
    };
    if libc::tcsetattr(stdin_fd, libc::TCSANOW, (&new_term) as *const _) != 0 {
        eprintln!("error setting terminal info");
        //std::process::exit(1);
        return false;
    }
    deferrer.push_back(Box::new(move || {
        if libc::tcsetattr(stdin_fd, libc::TCSANOW, (&old_term) as *const _) != 0 {
            eprintln!("error resetting terminal info");
        }
    }));
    true
}

#[allow(dead_code)]
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
enum Input {
    Unknown,
    Char(char),
    Backspace,
    ArrowUp,
    ArrowDown,
    ArrowRight,
    ArrowLeft,
    Del,
}

impl Input {
    fn parse(buf: &[u8]) -> Self {
        use Input::*;
        match buf {
            [127] => Backspace,
            [b] => Char(*b as _),
            [0x1b, 0x5b, 0x41] => ArrowUp,
            [0x1b, 0x5b, 0x42] => ArrowDown,
            [0x1b, 0x5b, 0x43] => ArrowRight,
            [0x1b, 0x5b, 0x44] => ArrowLeft,
            [0x1b, 0x5b, 0x33, 0x7e] => Del,
            _ => Unknown,
        }
        /*
        match buf.len() {
            1 => Char(buf[0] as _),
            3 => match buf {
                [0x1b, 0x5b, 0x41] => ArrowUp,
                [0x1b, 0x5b, 0x42] => ArrowDown,
                [0x1b, 0x5b, 0x43] => ArrowRight,
                [0x1b, 0x5b, 0x44] => ArrowLeft,
                _ => Unknown,
            }
            4 => match buf {
                [0x1b, 0x5b, 0x33, 0x7e] => Del,
                _ => Unknown,
            }
            _ => Unknown,
        }
        */
    }
}
