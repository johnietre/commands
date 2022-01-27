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

impl<T> Arg<T>
where
    T: Clone,
{
    pub fn new<S: ToString>(name: S) -> Self {
        Self {
            name: name.into(),
            ..Default::default(),
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

impl<T> Into<RcArg<T>> for Arg<T> {
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
    args_types: HashMap<String, ArgType>,
    short_args_types: HashMap<char, ArgType>,

    int_args: HashMap<String, RcArg<i32>>,
    float_args: HashMap<String, RcArg<f64>>,
    bool_args: HashMap<String, RcArg<bool>>,
    string_args: HashMap<String, RcArg<String>>,

    short_int_args: HashMap<char, RcArg<i32>>,
    short_float_args: HashMap<char, RcArg<f64>>,
    short_bool_args: HashMap<char, RcArg<bool>>,
    short_string_args: HashMap<char, RcArg<String>>,

    other_args: HashMap<String, RcArg<String>>,
    others_values: Vec<String>,

    exit_on_error: bool
}

impl ArgParser {
    pub fn new() -> Self {
        Self {
            args_types: HashMap::new(),
            short_args_types: HashMap::new(),
            int_args: HashMap::new(),
            float_args: HashMap::new(),
            bool_args: HashMap::new(),
            string_args: HashMap::new(),
            short_int_args: HashMap::new(),
            short_float_args: HashMap::new(),
            short_bool_args: HashMap::new(),
            short_string_args: HashMap::new(),
            other_args: HashMap::new(),
            other_value: Vec::new(),
            exit_on_error: false,
        }
    }

    pub fn add_int_arg(mut self, arg: Arg<i32>) -> Self {
        let (short, long) = (arg.short_name, arg.long_name.clone());
        self.args.insert(long.clone(), ArgType::Int);
        self.short_args.insert(short, ArgType::Int);
        let arg = arg.into();
        self.int_args.insert(long, arg);
        self.short_int_args.insert(short, Rc::clone(arg));
        self
    }

    pub fn add_float_arg(mut self, arg: Arg<f64>) -> Self {
        let (short, long) = (arg.short_name, arg.long_name.clone());
        self.args.insert(long.clone(), ArgType::Float);
        self.short_args.insert(short, ArgType::Float);
        let arg = arg.into();
        self.float_args.insert(long, arg);
        self.short_float_args.insert(short, Rc::clone(arg));
        self
    }

    pub fn add_bool_arg(mut self, arg: Arg<bool>) -> Self {
        let (short, long) = (arg.short_name, arg.long_name.clone());
        self.args.insert(long.clone(), ArgType::Bool);
        self.short_args.insert(short, ArgType::Bool);
        let arg = arg.into();
        self.bool_args.insert(long, arg);
        self.short_bool_args.insert(short, Rc::clone(arg));
        self
    }

    pub fn add_string_arg(mut self, arg: Arg<String>) -> Self {
        let (short, long) = (arg.short_name, arg.long_name.clone());
        self.args.insert(long.clone(), ArgType::Str);
        self.short_args.insert(short, ArgType::Str);
        let arg = arg.into();
        self.string_args.insert(long, arg);
        self.short_string_args.insert(short, Rc::clone(arg));
        self
    }

    pub fn exit_on_error(self, exit_on_error: bool) -> Self {
        Self { exit_on_error, ..self }
    }

    pub fn parse(mut self, args: Vec<String>) -> Self {
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
            if arg.starts_with("-") {
                arg.remove(0);
                if !args.starts_with("-") {
                    // TODO: Handle just a dash being passed
                    if args.len() > 0 {
                        self.parse_multiple_short(arg);
                    }
                    continue;
                }
                arg.remove(0);
                // TODO: Handle just two dashes being passed
                if args.len() == 0 {
                    continue;
                }
                let parts = arg.split_once("=");
                match self.args.get(&arg) {
                    Some(ArgType::Int) => match parts {
                        Some((arg, i)) => {
                            let i: i32 = i.parse();
                            let mut fl = self.int_args.get_mut(arg).unwrap().as_mut();
                            fl.set = true;
                            fl.value = i;
                        },
                        None => match iter.next() {
                            Some(i) => unsafe {
                                let i: i32 = i.parse().unwrap();
                                let fl = self.int_args.get_mut(&arg).unwrap().as_mut();
                                fl.set = true;
                                fl.value = i;
                            },
                            None => {
                                self.other_args.push(arg);
                            }
                        },
                    },
                    Some(ArgType::Float) => match parts {
                        Some((arg, f)) => unsafe {
                            let f: f64 = f.parse().unwrap();
                            let fl = self.float_args.get_mut(arg).unwrap().as_mut();
                            fl.set = true;
                            fl.value = f;
                        },
                        None => match iter.next() {
                            Some(f) => unsafe {
                                let f: f64 = f.parse().unwrap();
                                let fl = self.float_args.get_mut(&arg).unwrap().as_mut();
                                fl.set = true;
                                fl.value = f;
                            },
                            None => {
                                self.other_args.push(arg);
                            }
                        },
                    },
                    Some(ArgType::Bool) => match parts {
                        Some((arg, b)) => unsafe {
                            let b = *bools.get(&b).unwrap();
                            let fl = self.bool_args.get_mut(arg).unwrap().as_mut();
                            fl.set = true;
                            fl.value = b;
                        },
                        None => unsafe {
                            let fl = self.bool_args.get_mut(&arg).unwrap().as_mut();
                            fl.set = true;
                            fl.value = true;
                        },
                    },
                    Some(ArgType::Str) => match parts {
                        Some((arg, s)) => unsafe {
                            let fl = self.string_args.get_mut(arg).unwrap().as_mut();
                            fl.set = true;
                            fl.value = s.to_owned();
                        },
                        None => match iter.next() {
                            Some(s) => unsafe {
                                let fl = self.string_args.get_mut(&arg).unwrap().as_mut();
                                fl.set = true;
                                fl.value = s;
                            },
                            None => {
                                self.other_args.push(arg);
                            }
                        },
                    },
                    None => match iter.next() {
                        Some(value) => {
                            let count = arg.bytes().take_while(|&b| b == b'-').count();
                            let mut fl = if count == 1 {
                                Arg::new(arg.clone(), String::new(), value)
                            } else {
                                Arg::new(String::new(), arg.clone(), value)
                            };
                            fl.set = true;
                            self.other_args.insert(arg, Box::leak(Box::new(fl)).into());
                        }
                        None => {
                            self.other_args.push(arg);
                        }
                    },
                }
            } else {
                self.other_args.push(arg);
            }
        }
        self
    }

    pub fn get_int_arg(&self, arg: &String) -> Option<&Arg<i32>> {
        self.int_args.get(arg).map(|f| unsafe { f.as_ref() })
    }

    pub fn get_float_arg(&self, arg: &String) -> Option<&Arg<f64>> {
        self.float_args.get(arg).map(|f| unsafe { f.as_ref() })
    }

    pub fn get_bool_arg(&self, arg: &String) -> Option<&Arg<bool>> {
        self.bool_args.get(arg).map(|f| unsafe { f.as_ref() })
    }

    pub fn get_string_arg(&self, arg: &String) -> Option<&Arg<String>> {
        self.string_args.get(arg).map(|f| unsafe { f.as_ref() })
    }

    pub fn int_args(&self) -> HashMap<&str, &Arg<i32>> {
        self.int_args
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn float_args(&self) -> HashMap<&str, &Arg<f64>> {
        self.float_args
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn bool_args(&self) -> HashMap<&str, &Arg<bool>> {
        self.bool_args
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn string_args(&self) -> HashMap<&str, &Arg<String>> {
        self.string_args
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn other_args(&self) -> HashMap<&str, &Arg<String>> {
        self.other_args
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn other_args(&self) -> &Vec<String> {
        &self.other_args
    }

    fn parse_multiple_short(&mut self) {
    }

    fn parse_long(&mut self) {}
}
