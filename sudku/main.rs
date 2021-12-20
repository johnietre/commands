#![allow(unused_imports)]
use std::env;

use rand::Rng;
use termion::{self, color, style};

fn main() {
    let args: Vec<String> = env::args().skip(1).collect();
    if args.len() == 1 {
        println!("{}", args[0]);
    }

    println!("{:?}", termion::terminal_size().unwrap());
    Grid::new_9().print();
    println!("");
    Grid::new_16().print();
    if let Some(n) = Grid::new_9().get(0, 0) {
        println!("{}", n);
    }
}

type Row = Vec<u8>;

struct Grid {
    rows: Vec<Row>,
}

impl Grid {
    fn new_9() -> Grid {
        let mut rows = vec![vec![0u8; 9]; 9];
        for i in 0..9 {
            for j in 0..9 {
                rows[i][j] = ((i + j) % 9 + 1) as u8;
            }
        }
        Grid {
            rows,
        }
    }

    fn new_16() -> Grid {
        let mut rows = vec![vec![0u8; 16]; 16];
        for i in 0..16 {
            for j in 0..16 {
                rows[i][j] = ((i + j) % 16 + 1) as u8;
            }
        }
        Grid {
            rows,
        }
    }

    fn get(&self, r: usize, c: usize) -> Option<u8> {
        //self.rows.get(r).map(|row| row.get(c).map(|n_ref| *n_ref))
        self.rows.get(r).and_then(|row| row.get(c).cloned())
    }

    fn print(&self) {
        let grid_size = self.rows.len();
        let square_size = (grid_size as f32).sqrt() as usize;
        for (r, row) in self.rows.iter().enumerate() {
            let r1 = r + 1;
            if r1 % square_size == 0 && r1 != grid_size {
                print!("{}", style::Underline);
            }
            for (i, n) in row.iter().enumerate() {
                if i % square_size == 0 && i != 0 {
                    print!("|");
                }
                match n {
                    16 => print!("G"),
                    _ => print!("{:X}", n),
                }
            }
            println!("{}", style::Reset);
        }
    }
}
