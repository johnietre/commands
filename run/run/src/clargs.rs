#![allow(dead_code)]
use std::collections::HashMap;
use std::ptr::NonNull;

#[derive(Clone)]
pub struct Flag<T: Clone> {
    pub short_name: String,
    pub long_name: String,
    pub value: T,
    set: bool,
}

impl<T> Flag<T>
where
    T: Clone,
{
    pub fn new<S: ToString>(short_name: S, long_name: S, value: T) -> Self {
        Self {
            value,
            short_name: short_name.to_string(),
            long_name: long_name.to_string(),
            set: false,
        }
    }

    pub fn is_set(&self) -> bool {
        self.set
    }
}

enum FlagType {
    Int,
    Float,
    Bool,
    Str,
}

pub struct FlagParser {
    flags: HashMap<String, FlagType>,
    int_flags: HashMap<String, NonNull<Flag<i32>>>,
    float_flags: HashMap<String, NonNull<Flag<f64>>>,
    bool_flags: HashMap<String, NonNull<Flag<bool>>>,
    string_flags: HashMap<String, NonNull<Flag<String>>>,
    other_flags: HashMap<String, NonNull<Flag<String>>>,
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
        let (short, long) = (flag.short_name.clone(), flag.long_name.clone());
        self.flags.insert(short.clone(), FlagType::Int);
        self.flags.insert(long.clone(), FlagType::Int);
        let flag = NonNull::new(Box::leak(Box::new(flag))).unwrap();
        self.int_flags.insert(short, flag);
        self.int_flags.insert(long, flag);
        self
    }

    pub fn add_float_flag(mut self, flag: Flag<f64>) -> Self {
        let (short, long) = (flag.short_name.clone(), flag.long_name.clone());
        self.flags.insert(short.clone(), FlagType::Float);
        self.flags.insert(long.clone(), FlagType::Float);
        let flag = NonNull::new(Box::leak(Box::new(flag))).unwrap();
        self.float_flags.insert(short, flag);
        self.float_flags.insert(long, flag);
        self
    }

    pub fn add_bool_flag(mut self, flag: Flag<bool>) -> Self {
        let (short, long) = (flag.short_name.clone(), flag.long_name.clone());
        self.flags.insert(short.clone(), FlagType::Bool);
        self.flags.insert(long.clone(), FlagType::Bool);
        let flag = NonNull::new(Box::leak(Box::new(flag))).unwrap();
        self.bool_flags.insert(short, flag);
        self.bool_flags.insert(long, flag);
        self
    }

    pub fn add_string_flag(mut self, flag: Flag<String>) -> Self {
        let (short, long) = (flag.short_name.clone(), flag.long_name.clone());
        self.flags.insert(short.clone(), FlagType::Str);
        self.flags.insert(long.clone(), FlagType::Str);
        let flag = NonNull::new(Box::leak(Box::new(flag))).unwrap();
        self.string_flags.insert(short, flag);
        self.string_flags.insert(long, flag);
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
                        Some((flag, i)) => unsafe {
                            let i: i32 = i.parse().unwrap();
                            let mut fl = self.int_flags.get_mut(flag).unwrap().as_mut();
                            fl.set = true;
                            fl.value = i;
                        },
                        None => match iter.next() {
                            Some(i) => unsafe {
                                let i: i32 = i.parse().unwrap();
                                let fl = self.int_flags.get_mut(&arg).unwrap().as_mut();
                                fl.set = true;
                                fl.value = i;
                            },
                            None => {
                                self.other_args.push(arg);
                            }
                        },
                    },
                    Some(FlagType::Float) => match parts {
                        Some((flag, f)) => unsafe {
                            let f: f64 = f.parse().unwrap();
                            let fl = self.float_flags.get_mut(flag).unwrap().as_mut();
                            fl.set = true;
                            fl.value = f;
                        },
                        None => match iter.next() {
                            Some(f) => unsafe {
                                let f: f64 = f.parse().unwrap();
                                let fl = self.float_flags.get_mut(&arg).unwrap().as_mut();
                                fl.set = true;
                                fl.value = f;
                            },
                            None => {
                                self.other_args.push(arg);
                            }
                        },
                    },
                    Some(FlagType::Bool) => match parts {
                        Some((flag, b)) => unsafe {
                            let b = *bools.get(&b).unwrap();
                            let fl = self.bool_flags.get_mut(flag).unwrap().as_mut();
                            fl.set = true;
                            fl.value = b;
                        },
                        None => unsafe {
                            let fl = self.bool_flags.get_mut(&arg).unwrap().as_mut();
                            fl.set = true;
                            fl.value = true;
                        },
                    },
                    Some(FlagType::Str) => match parts {
                        Some((flag, s)) => unsafe {
                            let fl = self.string_flags.get_mut(flag).unwrap().as_mut();
                            fl.set = true;
                            fl.value = s.to_owned();
                        },
                        None => match iter.next() {
                            Some(s) => unsafe {
                                let fl = self.string_flags.get_mut(&arg).unwrap().as_mut();
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
                                Flag::new(arg.clone(), String::new(), value)
                            } else {
                                Flag::new(String::new(), arg.clone(), value)
                            };
                            fl.set = true;
                            self.other_flags.insert(arg, Box::leak(Box::new(fl)).into());
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

    pub fn get_int_flag(&self, flag: &String) -> Option<&Flag<i32>> {
        self.int_flags.get(flag).map(|f| unsafe { f.as_ref() })
    }

    pub fn get_float_flag(&self, flag: &String) -> Option<&Flag<f64>> {
        self.float_flags.get(flag).map(|f| unsafe { f.as_ref() })
    }

    pub fn get_bool_flag(&self, flag: &String) -> Option<&Flag<bool>> {
        self.bool_flags.get(flag).map(|f| unsafe { f.as_ref() })
    }

    pub fn get_string_flag(&self, flag: &String) -> Option<&Flag<String>> {
        self.string_flags.get(flag).map(|f| unsafe { f.as_ref() })
    }

    pub fn int_flags(&self) -> HashMap<&str, &Flag<i32>> {
        self.int_flags
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn float_flags(&self) -> HashMap<&str, &Flag<f64>> {
        self.float_flags
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn bool_flags(&self) -> HashMap<&str, &Flag<bool>> {
        self.bool_flags
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn string_flags(&self) -> HashMap<&str, &Flag<String>> {
        self.string_flags
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn other_flags(&self) -> HashMap<&str, &Flag<String>> {
        self.other_flags
            .iter()
            .map(|(k, v)| unsafe { (k.as_ref(), v.as_ref()) })
            .collect()
    }

    pub fn other_args(&self) -> &Vec<String> {
        &self.other_args
    }
}

impl Drop for FlagParser {
    fn drop(&mut self) {
        unsafe {
            self.int_flags.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.float_flags.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.bool_flags.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.string_flags.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
            self.other_flags.values().for_each(|k| {
                Box::from_raw(k.as_ptr());
            });
        }
    }
}
