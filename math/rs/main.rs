mod funcs;
mod utils;

fn main() {
    let args = std::env::args().collect::<Vec<String>>();
    utils::check_parens(args.get(1).unwrap());
    println!("Ok");
}
