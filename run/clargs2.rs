#![allow(dead_code)]
use std::collections::HashMap;

#[derive(Clone)]
pub struct Flag<T: Clone> {
    pub flag: String,
    pub value: T,
    set: bool,
}

impl<T> Flag<T> where T: Clone {
    pub fn new<S: ToString>(flag: S, value: T) -> Self {
        Self {
            flag: flag.to_string(),
            value,
            set: false,
        }
    }

    pub fn is_set(&self) -> bool {
        self.set
    }
}

pub enum FlagType {
    Int,
    Float,
    Bool,
    Str,
}

pub struct FlagParser {
    flags: HashMap<String, FlagType>,
    int_flags: HashMap<String, Flag<i32>>,
    float_flags: HashMap<String, Flag<f64>>,
    bool_flags: HashMap<String, Flag<bool>>,
    string_flags: HashMap<String, Flag<String>>,
    other_flags: HashMap<String, Flag<String>>,
    other_args: Vec<String>,
}

impl FlagParser {
    pub fn new() -> Self {
        Self {
            flags: HashMap::new(),
            int_flags: HashMap::new(),
            float_flags: HashMap::new(),
            bool_flags: HashMap::new(),
            string_flags: HashMap::new(),
            other_flags: HashMap::new(),
            other_args: Vec::new(),
        }
    }

    pub fn add_int_flag(mut self, flag: Flag<i32>) -> Self {
        self.flags.insert(flag.flag.clone(), FlagType::Int);
        self.int_flags.insert(flag.flag.clone(), flag);
        self
    }

    pub fn add_float_flag(mut self, flag: Flag<f64>) -> Self {
        self.flags.insert(flag.flag.clone(), FlagType::Float);
        self.float_flags.insert(flag.flag.clone(), flag);
        self
    }

    pub fn add_bool_flag(mut self, flag: Flag<bool>) -> Self {
        self.flags.insert(flag.flag.clone(), FlagType::Bool);
        self.bool_flags.insert(flag.flag.clone(), flag);
        self
    }

    pub fn add_string_flag(mut self, flag: Flag<String>) -> Self {
        self.flags.insert(flag.flag.clone(), FlagType::Str);
        self.string_flags.insert(flag.flag.clone(), flag);
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
                match self.flags.get(&arg) {
                    Some(FlagType::Int) => match parts {
                        Some((flag, i)) => {
                            let i: i32 = i.parse().unwrap();
                            let mut fl = self.int_flags.get_mut(flag).unwrap();
                            fl.set = true;
                            fl.value = i;
                        },
                        None => match iter.next() {
                            Some(i) => {
                                let i: i32 = i.parse().unwrap();
                                let mut fl = self.int_flags.get_mut(&arg).unwrap();
                                fl.set = true;
                                fl.value = i;
                            },
                            None => {
                                self.other_args.push(arg);
                            },
                        },
                    },
                    Some(FlagType::Float) => match parts {
                        Some((flag, f)) => {
                            let f: f64 = f.parse().unwrap();
                            let mut fl = self.float_flags.get_mut(flag).unwrap();
                            fl.set = true;
                            fl.value = f;
                        },
                        None => match iter.next() {
                            Some(f) => {
                                let f: f64 = f.parse().unwrap();
                                let mut fl = self.float_flags.get_mut(&arg).unwrap();
                                fl.set = true;
                                fl.value = f;
                            },
                            None => {
                                self.other_args.push(arg);
                            },
                        },
                    },
                    Some(FlagType::Bool) => match parts {
                        Some((flag, b)) => {
                            let b = *bools.get(&b).unwrap();
                            let mut fl = self.bool_flags.get_mut(flag).unwrap();
                            fl.set = true;
                            fl.value = b;
                        },
                        None => {
                            let mut fl = self.bool_flags.get_mut(&arg).unwrap();
                            fl.set = true;
                            fl.value = true;
                        },
                    },
                    Some(FlagType::Str) => match parts {
                        Some((flag, s)) => {
                            let mut fl = self.string_flags.get_mut(flag).unwrap();
                            fl.set = true;
                            fl.value = s.to_owned();
                        },
                        None => match iter.next() {
                            Some(s) => {
                                let mut fl = self.string_flags.get_mut(&arg).unwrap();
                                fl.set = true;
                                fl.value = s;
                            },
                            None => {
                                self.other_args.push(arg);
                            },
                        },
                    },
                    None => match iter.next() {
                        Some(value) => {
                            let mut fl = Flag::new(arg.clone(), value);
                            fl.set = true;
                            self.other_flags.insert(arg, fl);
                        },
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

    pub fn get_int_flag(&self, flag: &String) -> Option<Flag<i32>> {
        self.int_flags.get(flag).cloned()
    }

    pub fn get_float_flag(&self, flag: &String) -> Option<Flag<f64>> {
        self.float_flags.get(flag).cloned()
    }

    pub fn get_bool_flag(&self, flag: &String) -> Option<Flag<bool>> {
        self.bool_flags.get(flag).cloned()
    }

    pub fn get_string_flag(&self, flag: &String) -> Option<Flag<String>> {
        self.string_flags.get(flag).cloned()
    }

    pub fn int_flags(&self) -> &HashMap<String, Flag<i32>> {
        &self.int_flags
    }

    pub fn float_flags(&self) -> &HashMap<String, Flag<f64>> {
        &self.float_flags
    }

    pub fn bool_flags(&self) -> &HashMap<String, Flag<bool>> {
        &self.bool_flags
    }

    pub fn string_flags(&self) -> &HashMap<String, Flag<String>> {
        &self.string_flags
    }

    pub fn other_flags(&self) -> &HashMap<String, Flag<String>> {
        &self.other_flags
    }

    pub fn other_args(&self) -> &Vec<String> {
        &self.other_args
    }
}
