use std::io::{self, Write};
pub fn die(msg: &str, code: i32) {
    let stderr = io::stderr();
    let mut handle = stderr.lock();
    handle.write_fmt(format_args!("{}\n", msg))
        .expect("Error dying");
    std::process::exit(code);
}

pub fn check(expr: &String) {
    let (mut parens, mut open_parens, mut decimal, mut two) = (0, -1, -1, false);
    let mut prev = '\0';
    let expr_bytes = expr.as_bytes();
    let l = expr.len();
    for i in 0..l {
        let c = expr_bytes[i] as char;
        if c.is_ascii_digit() {
            prev = c;
            continue;
        } else if c.is_ascii_whitespace() {
            continue;
        }
        if prev == '.' {
            die("Invalid expression", 1);
        }
        if c == '.' {
            if decimal != -1 {
                die("Invalid expression", 1);
            }
            decimal = i as i32;
            prev = c;
            continue;
        } else if c == '*' || c == '/' {
            if expr_bytes[i - 1 as usize] as char == c {
                if two {
                    die("Invalid expression", 1);
                }
                two = true
            } else {
                two = false;
            }
            prev = c;
            decimal = -1;
            continue;
        } else if c == '+' || c == '^' || c == '-' || c == '%' {
            if prev == c {
                die("Invalid expression", 1);
            }
        } else if c == '(' {
            if open_parens == -1 {
                open_parens = i as i32;
            }
            parens += 1;
        } else if c == ')' {
            if parens == 0 {
                die("Mismatch parentheses", 1);
            } else if parens == 1 {
                open_parens = -1;
            }
            parens -= 1;
        } else {
            die("Invalid character", 1);
        }
        prev = c;
        decimal = -1;
    }
    if parens != 0 {
        die("Mismatch parentheses", 1);
    }
}