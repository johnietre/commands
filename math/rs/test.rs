fn main() {
    let iter = vec![1, 2, 3, 4, 5].into_iter();
    let iter =  iter
        .filter(|i| {
            println!("filter1: {}", i);
            i % 2 == 1
        });
    let iter = iter
        .filter(|i| {
            println!("filter2: {}", i);
            *i != 3
        });
    let iter = iter
        .map(|i| {
            println!("map3: {}", i);
            i * 10
        });
    iter
        .for_each(|i| {
            println!("for_each4: {}", i);
        });
    //println!("{}", sum);
}
