use std::fs::File;
use std::io::{self, prelude::*, BufReader};

mod args;
use args::*;

fn main() {
    let mut args = ArgParser::new();
    args
        .string_arg("o", "", None)
        .bool_arg("", "files", Some(false));
    args.parse(std::env::args().skip(1));

    if args.floating().len() == 0 {
        // Read from stdin until EOF
        return
    } else if *args.files.get_flag("files").unwrap().bool_value().unwrap() {
        //
        return
    }
    //
}

struct Document<W: Write> {
    in_para: bool,
    in_list: bool,
    in_bold: bool,
    in_italic: bool,
    in_strike: bool,
    in_highlight: bool,
    in_sub: bool,
    in_super: bool,
    indent: usize,
    writer: W,
}

impl<W: Write> Document<W> {
    fn new(writer: W) -> Self {
        Self {
            in_para: false,
            in_list: false,
            in_bold: false,
            in_italic: false,
            in_strike: false,
            in_highlight: false,
            in_sub: false,
            in_super: false,
            indent: 0,
            writer,
        }
    }
    
    // TODO: Cache prev line to check for "==" or "--" as the next line
    fn add_line(&mut self, md_line: &str) -> io::Result<()> {
        if md_line.trim().len() == 0 {
            self.in_para = false;
            return Ok(());
        }

        // Check for horizontal rule
        if trimmed.len() >= 3 && "*-_".contains(first) && chars.all(|c| c == first) {
            return write!(self.writer, "<hr>").map(|_| ());
        }

        let mut html_line = String::new();
        let line = if let Some((first, rest)) = md_line.split_once(|c| c.is_whitespace()) {
            //
            // TODO: Parse beginning of line
            rest
        } else {
            &md_line
        };
        let mut rest = String::with_capacity(line.len());
        let mut chars = line.chars();
        let mut prev = chars.next().unwrap();
        for c in line.chars() {
            match c {
                '*' => (),
                '=' => (),
                '~' => (),
                '^' => (),
            }
        }
    }

    // Return how many bytes were written?
    fn add_line(&mut self, md_line: String) -> io::Result<()> {
        if md_line.trim().len() == 0 {
            self.in_para = false;
            return Ok(());
        }

        let trimmed = md_line.trim();
        if trimmed == "" {
            self.in_para = false;
            return Ok(());
        }
        let mut chars = trimmed.chars();
        let first = chars.next().unwrap();

        let mut html_line = String::new();
        let mut toks = md_line.split_whitespace();
        let Some(first) = toks.next() else {
            // TODO: Add blank line?
        };
        // Check for heading
        if first.startswith("#") {
            if first.len() <= 6 && first.bytes().all(|b| b == b'#') {
                return write!(self.writer, "<h>"
            }
            let _ = self.writer.write(line.as_bytes())?;
            return Ok(())
        }
        match first {
            "#" => 
        }
    }

    fn writer(&mut self) -> &mut W {
        &mut self.writer
    }

    fn flush(&mut self) -> io::Result<()> {
        self.writer.flush()
    }
}
