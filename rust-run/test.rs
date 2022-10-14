macro_rules! create_func {
    ($expr1:expr, $expr2:expr) => {
        fn test1() {
            println!("{}", $expr1);
        }
        fn test2() {
            println!("{}", $expr2);
        }
    };
}

macro_rules! add {
    ($x:expr, $y:expr) => { $x + $y };
}

macro_rules! pat {
    ($i:ident) => (Some($i))
}

macro_rules! func_maker {
    ($name:ident $params:tt $body:tt) => {
        fn $name $params $body
    };
}

create_func!("test1", "test2");

/*
func_maker!{
    fn test3() {
        println!("made");
    }
}
*/

fn main() {
    test1();
    test2();
    println!("{}", add!(1, 3));
    //let x = vec![1, 2, 3];
    if let pat!(x) = Some(1) {
        assert_eq!(x, 1);
    }
    //test3();
    dbg!();
}
