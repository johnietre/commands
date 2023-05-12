// TODO: Implement functions (sqrt, trig, etc).
// TODO: Fix: will panic on (2)(pi)(2^2)
#![feature(linked_list_cursors)]
use std::collections::LinkedList;
use std::env::args;
use std::f64::consts::{E, PI};
use std::io::{prelude::*, stdin};

fn main() {
    let expr = args().skip(1).collect::<Vec<_>>().join("");
    if expr != "" {
        match Parser::<f64>::eval(&expr) {
            Ok(res) => println!("{}", res),
            Err(e) => eprintln!("{}", e),
        }
        return
    }
    for line in stdin().lock().lines() {
        let line = match line {
            Ok(mut line) => {
                line.make_ascii_lowercase();
                line
            }
            Err(e) => {
                eprintln!("error reading line: {}", e);
                return;
            }
        };
        let line = line.trim();
        if line == "help" {
            print_help();
        } else if line != "" {
            match Parser::<f64>::eval(line) {
                Ok(res) => println!("{}", res),
                Err(e) => eprintln!("{}", e),
            }
        }
    }
}

#[derive(Default)]
struct Parser<T>(std::marker::PhantomData<T>);

impl<T> Parser<T> {}

impl Parser<f64> {
    fn eval(expr: &str) -> Result<f64, Box<dyn std::error::Error>> {
        use Token::*;
        use Operator::*;
        let mut op_stack: LinkedList<Operator> = LinkedList::new();
        let mut out_queue: LinkedList<Token<f64>> = LinkedList::new();
        let mut start = usize::MAX;
        let mut parens = 0u32;
        let mut prev = 0u8;
        let mut neg = false;
        for (i, b) in expr.bytes().enumerate() {
            match b {
                b'0'..=b'9' => {
                    if start == usize::MAX {
                        start = i;
                    }
                }
                // TODO: Do better?
                b'!' | b'+' | b'-' | b'*' | b'/' | b'^' => {
                    if b"\0(+-*/^".contains(&prev) {
                        if b == b'-' && !neg {
                            neg = true;
                            prev = b;
                            continue;
                        }
                        return Err(
                            format!(
                                r#"bad expression at position {}: "{}""#, i + 1, b as char,
                            ).into(),
                        );
                    }
                    if start != usize::MAX {
                        let v = match parse_num(&expr[start..i]) {
                            Some(v) => v,
                            None => return Err(
                                format!(
                                    r#"bad number at position {}: "{}""#, i + 1, b as char,
                                ).into(),
                            ),
                        };
                        out_queue.push_back(Value(v));
                        if neg {
                            let op = Neg;
                            while op_stack.front().map_or(false, |o| op.should_pop(o)) {
                                out_queue.push_back(Op(op_stack.pop_front().unwrap()));
                            }
                            op_stack.push_front(op);
                            neg = false;
                        }
                        start = usize::MAX;
                    }
                    let op = Operator::from(b);
                    // TODO: Use if?
                    while op_stack.front().map_or(false, |o| op.should_pop(o)) {
                        out_queue.push_back(Op(op_stack.pop_front().unwrap()));
                    }
                    op_stack.push_front(op);
                }
                b'(' => {
                    if start != usize::MAX {
                        let v = match parse_num(&expr[start..i]) {
                            Some(v) => v,
                            None => return Err(
                                format!(
                                    r#"bad number at position {}: "{}""#, i + 1, b as char,
                                ).into(),
                            ),
                        };
                        start = usize::MAX;
                        out_queue.push_back(Value(v));
                        if neg {
                            let op = Neg;
                            while op_stack.front().map_or(false, |o| op.should_pop(o)) {
                                out_queue.push_back(Op(op_stack.pop_front().unwrap()));
                            }
                            op_stack.push_front(op);
                            neg = false;
                        }
                        // Add implicit multiplication
                        let op = Mul;
                        // TODO: Use if?
                        while op_stack.front().map_or(false, |o| op.should_pop(o)) {
                            out_queue.push_back(Op(op_stack.pop_front().unwrap()));
                        }
                        op_stack.push_front(op);
                    }
                    parens += 1;
                    op_stack.push_front(Paren);
                }
                b')' => {
                    if parens == 0 {
                        return Err(
                            format!("unmatched parenthesis at position: {}", i + 1).into(),
                        );
                    }
                    parens -= 1;
                    if start == usize::MAX {
                        if prev == b'(' {
                            return Err(
                                format!("don't put empty parentheses").into(),
                            );
                        } else if prev != b')' {
                            return Err(
                                format!(
                                    r#"bad expression at position {}: "{}""#, i, prev as char,
                                ).into(),
                            );
                        }
                    }
                    let v = match parse_num(&expr[start..i]) {
                        Some(v) => v,
                        None => return Err(
                            format!(
                                r#"bad number at position {}: "{}""#, i + 1, b as char,
                            ).into(),
                        ),
                    };
                    start = usize::MAX;
                    out_queue.push_back(Value(v));
                    if neg {
                        let op = Neg;
                        while op_stack.front().map_or(false, |o| op.should_pop(o)) {
                            out_queue.push_back(Op(op_stack.pop_front().unwrap()));
                        }
                        op_stack.push_front(op);
                        neg = false;
                    }
                    // Should always be some since the parens check above checks for mismatch
                    // parentheses
                    let i = op_stack.iter().position(|o| matches!(o, Paren)).unwrap();
                    let rest = op_stack.split_off(i + 1);
                    // Remove the parenthesis
                    op_stack.pop_back();
                    out_queue.extend(op_stack.into_iter().map(|o| Op(o)));
                    op_stack = rest;
                }
                _ if b.is_ascii_whitespace() => {
                    if start != usize::MAX {
                        let v = match parse_num(&expr[start..i]) {
                            Some(v) => v,
                            None => return Err(
                                format!(
                                    r#"bad number at position {}: "{}""#, i + 1, b as char,
                                ).into(),
                            ),
                        };
                        out_queue.push_back(Value(v));
                        start = usize::MAX;
                        if neg {
                            let op = Neg;
                            while op_stack.front().map_or(false, |o| op.should_pop(o)) {
                                out_queue.push_back(Op(op_stack.pop_front().unwrap()));
                            }
                            op_stack.push_front(op);
                            neg = false;
                        }
                    }
                }
                _ => {
                    if start == usize::MAX {
                        start = i;
                    }
                }
            }
            if !b.is_ascii_whitespace() {
                prev = b;
            }
        }
        if parens != 0 {
            return Err(format!("unclosed parentheses").into());
        }
        if start != usize::MAX {
            let v = match parse_num(&expr[start..]) {
                Some(v) => v,
                None => return Err(
                    format!(r"bad number at or after position {}", start + 1).into(),
                ),
            };
            out_queue.push_back(Value(v));
            if neg {
                let op = Neg;
                while op_stack.front().map_or(false, |o| op.should_pop(o)) {
                    out_queue.push_back(Op(op_stack.pop_front().unwrap()));
                }
                op_stack.push_front(op);
            }
        }
        out_queue.extend(op_stack.into_iter().map(|o| Op(o)));
        //out_queue.iter().for_each(|o| println!("{:?}", o));
        let mut cur = out_queue.cursor_front_mut();
        // Only iterate this many times to avoid a possible cycle
        for _ in 0..1_000 {
            if cur.peek_next().is_none() && cur.peek_prev().is_none() {
                match cur.current() {
                    Some(Value(res)) => return Ok(*res),
                    _ => return Err("bad expression".into()),
                }
            }
            if let Op(op) = cur.current().copied().unwrap() {
                cur.remove_current(); cur.move_prev();
                let n = op.num_args();
                let mut args = vec![0.0; n];
                for i in 0..n {
                    match cur.remove_current() {
                        Some(Value(v)) => args[n - i - 1] = v,
                        _ => panic!("we messed up"),
                    }
                    cur.move_prev();
                }
                cur.insert_after(Value(op.calc(args)?));
            }
            cur.move_next();
        }
        Err("cycle detected".into())
    }
}

fn parse_num(nstr: &str) -> Option<f64> {
    if let Ok(n) = nstr.parse() {
        Some(n)
    } else {
        match nstr.to_ascii_lowercase().as_str() {
            "pi" => Some(PI),
            "e" => Some(E),
            _ => None,
        }
    }
}

#[derive(Clone, Copy, Debug)]
enum Token<T: Copy + std::fmt::Debug> {
    Value(T),
    Op(Operator),
}

#[derive(Clone, Copy, Debug)]
enum Operator {
    Neg, // Negation
    Add,
    Sub,
    Mul,
    Div,
    Pow,
    // Parenthesis
    Paren,
    Fact, // Factorial
}

impl Operator {
    fn precedence(&self) -> u8 {
        use Operator::*;
        match self {
            Paren => 0, // NOTE: IDK
            Add | Sub => 2,
            Neg | Mul | Div => 3,
            Pow => 4,
            Fact => 255,
        }
    }

    // Returns true if the oeprator is left associative
    fn is_left(&self) -> bool {
        use Operator::*;
        match self {
            Neg | Add | Sub | Mul | Div => true,
            _ => false,
        }
    }

    // Returns true if the passed operator should be popped
    fn should_pop(&self, other: &Self) -> bool {
        // Should pop if self prec == other and self is left ass. or self prec < other
        self < other || (self == other && self.is_left())
    }

    fn num_args(&self) -> usize {
        use Operator::*;
        match self {
            Add | Sub | Mul | Div | Pow => 2,
            Neg | Fact => 1,
            Paren => panic!("no args taken for {:?}", self),
        }
    }

    fn calc(&self, args: Vec<f64>) -> Result<f64, Box<dyn std::error::Error>> {
        use Operator::*;
        if self.num_args() != args.len() {
            panic!("{:?} expected {} arg(s)", self, self.num_args());
        }
        let res = match self {
            Neg => -args[0],
            Add => args[0] + args[1],
            Sub => args[0] - args[1],
            Mul => args[0] * args[1],
            Div => {
                if args[1] == 0.0 {
                    return Err("divide by 0".into());
                }
                args[0] / args[1]
            }
            Pow => args[0].powf(args[1]),
            Paren => unreachable!(),
            Fact => {
                let v = args[0];
                if v.trunc() != v {
                    return Err("cannot take factorial of non-integer".into());
                }
                if v < 0.0 {
                    return Err("cannot take factorial of negative number".into());
                }
                let v = v as u64;
                if v == 0 || v == 1 {
                    1.0
                } else {
                    if let Some(v) = (3..=v).try_fold(2u64, |acc, i| acc.checked_mul(i)) {
                        v as f64
                    } else {
                        return Err("overflow on factorial".into());
                    }
                }
            }
        };
        Ok(res)
    }
}

impl From<u8> for Operator {
    // Parentheses shouldn't be passed
    fn from(u: u8) -> Self {
        use Operator::*;
        match u {
            b'+' => Add,
            b'-' => Sub,
            b'*' => Mul,
            b'/' => Div,
            b'^' => Pow,
            b'!' => Fact,
            _ => panic!("unknown operator: {}", u as char),
        }
    }
}

impl PartialEq for Operator {
    fn eq(&self, other: &Self) -> bool {
        self.precedence() == other.precedence()
    }
}

impl PartialOrd for Operator {
    fn partial_cmp(&self, &other: &Self) -> Option<std::cmp::Ordering> {
        self.precedence().partial_cmp(&other.precedence())
    }
}

fn print_help() {
    println!("\
    +: addition, -: subtraction, *: multiplication, /: division\n\
    ^: power, !: factorial\
    ");
}
