#![allow(dead_code)]
use std::collections::HashMap;
use std::ptr::NonNull;

#[derive(Clone, Default)]
pub struct Arg<T: Clone + Default> {
    short_name: char,
    long_name: String,
    val: T,
    set: bool,
}

impl<T> Arg<T>
where
    T: Clone,
{
    pub fn new<S: ToString>(short_name: char, long_name: S, value: T) -> Self {
        Self {
            value,
            short_name: short_name.to_string(),
            long_name: long_name.to_string(),
            set: false,
        }
    }

    pub fn new<S: ToString>(name: S) -> Self {
        Self {
            long_name: name.into(),
            ..Default::default(),
        }
    }

    pub fn short(name: char) -> Self {
        self.short_name = name;
        self
    }

    pub fn value(self, val: T) -> Self {
        self.val = val;
        self
    }

    pub fn is_set(&self) -> bool {
        self.set
    }
}

enum ArgType {
    Int,
    Float,
    Bool,
    Str,
}

pub struct ArgParser {
    args: HashMap<String, ArgType>,

    int_args: HashMap<String, NonNull<Arg<i32>>>,
    float_args: HashMap<String, NonNull<Arg<f64>>>,
    bool_args: HashMap<String, NonNull<Arg<bool>>>,
    string_args: HashMap<String, NonNull<Arg<String>>>,

    short_int_args: HashMap<char, NonNull<Arg<i32>>>,
    short_float_args: HashMap<char, NonNull<Arg<f64>>>,
    short_bool_args: HashMap<char, NonNull<Arg<bool>>>,
    short_string_args: HashMap<char, NonNull<Arg<String>>>,

    other_args: HashMap<String, NonNull<Arg<String>>>,
    others_values: Vec<String>,
}

impl ArgParser {
    pub fn new() -> Self {
        Self {
            args: HashMap::new(),
            int_args: HashMap::new(),
            float_args: HashMap::new(),
            bool_args: HashMap::new(),
            string_args: HashMap::new(),
            other_args: HashMap::new(),
            other_args: Vec::new(),
        }
    }

    pub fn add_int_arg(mut self, arg: Arg<i32>) -> Self {
        let (short, long) = (arg.short_name.clone(), arg.long_name.clone());
        self.args.insert(short.clone(), ArgType::Int);
        self.args.insert(long.clone(), ArgType::Int);
        let arg = NonNull::new(Box::leak(Box::new(arg))).unwrap();
        self.int_args.insert(short, arg);
        self.int_args.insert(long, arg);
        self
    }

    pub fn add_float_arg(mut self, arg: Arg<f64>) -> Self {
        let (short, long) = (arg.short_name.clone(), arg.long_name.clone());
        self.args.insert(short.clone(), ArgType::Float);
        self.args.insert(long.clone(), ArgType::Float);
        let arg = NonNull::new(Box::leak(Box::new(arg))).unwrap();
        self.float_args.insert(short, arg);
        self.float_args.insert(long, arg);
        self
    }

    pub fn add_bool_arg(mut self, arg: Arg<bool>) -> Self {
        let (short, long) = (arg.short_name.clone(), arg.long_name.clone());
        self.args.insert(short.clone(), ArgType::Bool);
        self.args.insert(long.clone(), ArgType::Bool);
        let arg = NonNull::new(Box::leak(Box::new(arg))).unwrap();
        self.bool_args.insert(short, arg);
        self.bool_args.insert(long, arg);
        self
    }

    pub fn add_string_arg(mut self, arg: Arg<String>) -> Self {
        let (short, long) = (arg.short_name.clone(), arg.long_name.clone());
        self.args.insert(short.clone(), ArgType::Str);
        self.args.insert(long.clone(), ArgType::Str);
        let arg = NonNull::new(Box::leak(Box::new(arg))).unwrap();
        self.string_args.insert(short, arg);
        self.string_args.insert(long, arg);
        self
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
                let parts = arg.split_once("=");
                match self.args.get(&arg) {
                    Some(ArgType::Int) => match parts {
                        Some((arg, i)) => unsafe {
                            let i: i32 = i.parse().unwrap();
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
}

impl Drop for ArgParser {
    fn drop(&mut self) {
        unsafe {
            self.int_args.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.float_args.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.bool_args.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.string_args.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.other_args.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
        }
    }
}
