#![allow(dead_code)]
use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;

type RcArg<T> = Rc<RefCell<Arg<T>>>;

#[derive(Clone, Default)]
pub struct Arg<T: Clone + Default> {
    name: String,
    short_name: char,
    val: T,
    set: bool,
}

impl<T: Clone + Default> Arg<T> {
    pub fn new<S: ToString>(name: S) -> Self {
        Self {
            name: name.to_string(),
            ..Default::default()
        }
    }

    pub fn short(self, short_name: char) -> Self {
        Self { short_name, ..self }
    }

    pub fn value(self, val: T) -> Self {
        Self { val, ..self }
    }

    pub fn is_set(&self) -> bool {
        self.set
    }
}

impl<T: Clone + Default> Into<RcArg<T>> for Arg<T> {
    fn into(self) -> RcArg<T> {
        Rc::new(RefCell::new(self))
    }
}

enum ArgType {
    Int,
    Float,
    Bool,
    Str,
}

pub struct ArgParser {
    arg_types: HashMap<String, ArgType>,
    short_arg_types: HashMap<String, ArgType>,

    int_args: HashMap<String, RcArg<i32>>,
    float_args: HashMap<String, RcArg<f64>>,
    bool_args: HashMap<String, RcArg<bool>>,
    string_args: HashMap<String, RcArg<String>>,

    short_int_args: HashMap<String, RcArg<i32>>,
    short_float_args: HashMap<String, RcArg<f64>>,
    short_bool_args: HashMap<String, RcArg<bool>>,
    short_string_args: HashMap<String, RcArg<String>>,

    other_args: HashMap<String, RcArg<String>>,
    values: Vec<String>,

    exit_on_error: bool
}

impl ArgParser {
    pub fn new() -> Self {
        Self {
            arg_types: HashMap::new(),
            short_arg_types: HashMap::new(),
            int_args: HashMap::new(),
            float_args: HashMap::new(),
            bool_args: HashMap::new(),
            string_args: HashMap::new(),
            short_int_args: HashMap::new(),
            short_float_args: HashMap::new(),
            short_bool_args: HashMap::new(),
            short_string_args: HashMap::new(),
            other_args: HashMap::new(),
            values: Vec::new(),
            exit_on_error: false,
        }
    }

    pub fn add_int_arg(mut self, arg: Arg<i32>) -> Self {
        let name = "--".to_string() + &arg.name;
        let short = "-".to_string() + &arg.short_name.to_string();
        let rc_arg = arg.into();
        self.arg_types.insert(name.clone(), ArgType::Int);
        self.int_args.insert(name, Rc::clone(&rc_arg));
        if short != "-" {
            self.short_arg_types.insert(short.clone(), ArgType::Int);
            self.short_int_args.insert(short, Rc::clone(&rc_arg));
        }
        self
    }

    pub fn add_float_arg(mut self, arg: Arg<f64>) -> Self {
        let name = "--".to_string() + &arg.name;
        let short = "-".to_string() + &arg.short_name.to_string();
        let rc_arg = arg.into();
        self.arg_types.insert(name.clone(), ArgType::Float);
        self.float_args.insert(name, Rc::clone(&rc_arg));
        if short != "-" {
            self.short_arg_types.insert(short.clone(), ArgType::Float);
            self.short_float_args.insert(short, Rc::clone(&rc_arg));
        }
        self
    }

    pub fn add_bool_arg(mut self, arg: Arg<bool>) -> Self {
        let name = "--".to_string() + &arg.name;
        let short = "-".to_string() + &arg.short_name.to_string();
        let rc_arg = arg.into();
        self.arg_types.insert(name.clone(), ArgType::Bool);
        self.bool_args.insert(name, Rc::clone(&rc_arg));
        if short != "-" {
            self.short_arg_types.insert(short.clone(), ArgType::Bool);
            self.short_bool_args.insert(short, Rc::clone(&rc_arg));
        }
        self
    }

    pub fn add_string_arg(mut self, arg: Arg<String>) -> Self {
        let name = "--".to_string() + &arg.name;
        let short = "-".to_string() + &arg.short_name.to_string();
        let rc_arg = arg.into();
        self.arg_types.insert(name.clone(), ArgType::Str);
        self.string_args.insert(name, Rc::clone(&rc_arg));
        if short != "-" {
            self.short_arg_types.insert(short.clone(), ArgType::Str);
            self.short_string_args.insert(short, Rc::clone(&rc_arg));
        }
        self
    }

    pub fn exit_on_error(self, exit_on_error: bool) -> Self {
        Self { exit_on_error, ..self }
    }

    pub fn parse(mut self, args: Vec<String>) -> Self {
        use self::ArgType;
        let mut bools: HashMap<&str, bool> = HashMap::new();
        bools.insert("t", true);
        bools.insert("T", true);
        bools.insert("true", true);
        bools.insert("1", true);
        bools.insert("f", false);
        bools.insert("F", false);
        bools.insert("false", false);
        bools.insert("0", false);

        let mut iter = args.into_iter();
        while let Some(mut arg) = iter.next() {
            if arg.starts_with("--") {
                //
            } else if arg.starts_with("-") {
                let has_value;
                let (short, value) = if let Some((s, v)) = arg.split_once("=") {
                    has_value = true;
                    (s.to_string(), v.to_string())
                } else {
                    has_value = false;
                    (arg.clone(), String::new())
                };
                match self.short_arg_types.get(&short) {
                    Some(Int) => (),
                    Some(Float) => (),
                    Some(Bool) => (),
                    Some(Str) => (),
                    None => self.values.push(arg),
                }
            } else {
                self.values.push(arg);
            }
        }
        self
    }

    pub fn get_int_arg(&self, arg: &String) -> Option<&Arg<i32>> {
        self.int_args.get(arg).map(|a| &*a.borrow())
    }

    pub fn get_float_arg(&self, arg: &String) -> Option<&Arg<f64>> {
        self.float_args.get(arg).map(|a| &*a.borrow())
    }

    pub fn get_bool_arg(&self, arg: &String) -> Option<&Arg<bool>> {
        self.bool_args.get(arg).map(|a| &*a.borrow())
    }

    pub fn get_string_arg(&self, arg: &String) -> Option<&Arg<String>> {
        self.string_args.get(arg).map(|a| &*a.borrow())
    }

    pub fn int_args(&self) -> HashMap<&str, &Arg<i32>> {
        self.int_args
            .iter()
            .map(|(k, v)| (k.as_str(), &*v.borrow()))
            .collect()
    }

    pub fn float_args(&self) -> HashMap<&str, &Arg<f64>> {
        self.float_args
            .iter()
            .map(|(k, v)| (k.as_str(), &*v.borrow()))
            .collect()
    }

    pub fn bool_args(&self) -> HashMap<&str, &Arg<bool>> {
        self.bool_args
            .iter()
            .map(|(k, v)| (k.as_str(), &*v.borrow()))
            .collect()
    }

    pub fn string_args(&self) -> HashMap<&str, &Arg<String>> {
        self.string_args
            .iter()
            .map(|(k, v)| (k.as_str(), &*v.borrow()))
            .collect()
    }

    pub fn other_args(&self) -> HashMap<&str, &Arg<String>> {
        self.other_args
            .iter()
            .map(|(k, v)| (k.as_str(), &*v.borrow()))
            .collect()
    }

    pub fn values(&self) -> &Vec<String> {
        &self.values
    }
}
