#![allow(dead_code)]
fn factorial_recur(i: u128) -> u128 {
    if i == 1 {
        return 1;
    }
    i * factorial_recur(i-1)
}

fn factorial(mut i: u128) -> u128 {
    let mut res = 1u128;
    while i > 1 {
        res *= i;
        i -= 1;
    }
    res
}

fn main() {
    println!("{}", factorial(30));
}
