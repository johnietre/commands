enum EvalError {
    NumErr(String)
}

fn factorial(i: i128) -> Result<i128, EvalError> {
    if i < 0 {
        return NumError(String::from("Negative factorial"));
    } else if i <= 1 {
        return 1;
    }
    match . {
        //
    }
}