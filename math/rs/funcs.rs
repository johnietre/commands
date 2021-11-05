pub fn avg(nums: &Vec<f64>) -> f64 {
    nums.iter().sum::<f64>() / nums.len() as f64
}

pub fn mean(nums: &Vec<f64>) -> f64 {
    avg(nums)
}

pub fn factorial(n: f64) -> f64 {
    // Possibly handle for negative numbers
    match n as i64 {
        0 => 1.0,
        n => (1..=n).product::<i64>() as f64,
    }
}
