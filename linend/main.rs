use std::io::{prelude::*, BufReader};
use std::fs;

fn main() {
    let args = std::env::args().skip(1).collect::<Vec<_>>();
    if args.len() != 2 {
        die("usage: linend <FILE> <LF|CRLF>");
    }
    let [fname, end] = args.as_slice() else { unreachable!() };
    fs::metadata(&fname).unwrap_or_else(|e| die(e));
    let end = match end.to_lowercase().as_str() {
        "lf" => String::from("\n"),
        "crlf" => String::from("\r\n"),
        _ => die("invalid line ending, only LF or CRLF accepted"),
    };
    let mut i = 0;
    let tmp_fname = loop {
        if i == 100 {
            die("too many attempts to create temp file");
        }
        let tmp = format!("{}.tmp.{}", fname, i);
        if fs::metadata(&tmp).is_err() {
            break tmp;
        }
        i += 1;
    };
    let infile = fs::File::open(&fname).unwrap_or_else(|e| die(format!("error opening file: {}", e)));
    let mut outfile = fs::File::create(&tmp_fname).unwrap_or_else(|e| die(format!("error creating temp file: {}", e)));
    for line in BufReader::new(infile).lines() {
        let line = match line {
            Ok(line) => line,
            Err(e) => {
                let _ = fs::remove_file(&tmp_fname);
                die(format!("error reading from file: {}", e))
            }
        };
        if let Err(e) = write!(outfile, "{}{}", line, end) {
            let _ = fs::remove_file(&tmp_fname);
            die(format!("error writing to temp file: {}", e));
        }
    }
    if let Err(e) = fs::rename(&tmp_fname, fname) {
        let _ = fs::remove_file(&tmp_fname);
        die(format!("error renaming temp file: {}", e));
    }
}

fn die(err: impl std::fmt::Display) -> ! {
    eprintln!("{}", err);
    std::process::exit(1)
}
