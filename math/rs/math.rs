mod utils;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() != 2 {
        utils::die("Usage: math [expression]", 1);
    }
    let expr = &args[1];
    utils::check(&expr);
}
