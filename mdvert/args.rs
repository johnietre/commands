use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;

pub struct ArgParser {
    flags: HashMap<String, Rc<RefCell<ArgValue>>>,
    parsed_flags: HashMap<String, Rc<ArgValue>>,
    floating: Vec<String>,
}

impl ArgParser {
    pub fn new() -> Self {
        Self {
            flags: HashMap::new(),
            parsed_flags: HashMap::new(),
            floating: Vec::new(),
        }
    }

    pub fn usize_arg(
        &mut self,
        short: impl ToString,
        long: impl ToString,
        default: Option<usize>,
    ) -> &mut Self {
        let (short, long) = (short.to_string(), long.to_string());
        let arg_value = Rc::new(RefCell::new(ArgValue::Usize(Value::new(default))));
        if short != "" {
            assert!(!self.flags.contains_key(&short), "flag exists: {}", short);
            self.flags.insert(short, Rc::clone(&arg_value));
        }
        if long != "" {
            assert!(!self.flags.contains_key(&long), "flag exists: {}", long);
            self.flags.insert(long, arg_value);
        }
        self
    }

    pub fn string_arg(
        &mut self,
        short: impl ToString,
        long: impl ToString,
        default: Option<String>,
    ) -> &mut Self {
        let (short, long) = (short.to_string(), long.to_string());
        let arg_value = Rc::new(RefCell::new(ArgValue::String(Value::new(default))));
        if short != "" {
            assert!(!self.flags.contains_key(&short), "flag exists: {}", short);
            self.flags.insert(short, Rc::clone(&arg_value));
        }
        if long != "" {
            assert!(!self.flags.contains_key(&long), "flag exists: {}", long);
            self.flags.insert(long, arg_value);
        }
        self
    }

    //fn get_flag<'a>(&'a self, flag: &str) -> Option<Ref<'a, ArgValue>> {
    pub fn get_flag<'a>(&self, flag: &str) -> Option<&ArgValue> {
        self.parsed_flags.get(flag).map(|v| &**v)
    }

    #[allow(dead_code)]
    pub fn floating(&self) -> &[String] {
        self.floating.as_slice()
    }

    pub fn parse<I: IntoIterator<Item = String>>(&mut self, args: I) {
        let mut prev_flag = None::<String>;
        for arg in args {
            if let Some(flag) = prev_flag.take() {
                let Ok(value) = arg.parse() else {
                    eprintln!("invalid value passed for {}", flag);
                    exit(1)
                };
                if self.flags[&flag].borrow_mut().set_value(value).is_err() {
                    eprintln!("bad value for {}", flag);
                    exit(1);
                }
                continue;
            }
            if arg.starts_with("--") {
                let flag = arg.get(2..).unwrap();
                match arg.get(2..).unwrap().split_once('=') {
                    Some((flag, value)) => {
                        if !self.flags.contains_key(flag) {
                            self.floating.push(arg);
                        } else {
                            let Ok(value) = value.parse() else {
                                eprintln!("invalid value passed for {}", flag);
                                exit(1)
                            };
                            if self.flags[flag].borrow_mut().set_value(value).is_err() {
                                eprintln!("bad value for {}", flag);
                                exit(1);
                            }
                        }
                    }
                    None => {
                        /*
                        if !self.flags.contains_key(flag) {
                            self.floating.push(arg);
                        } else {
                            prev_flag.replace(flag.to_string());
                        }
                        */
                        if let Some(fr) = self.flags.get(flag) {
                            let mut f = fr.borrow_mut();
                            if matches!(f, ArgValue::Bool(_)) {
                                let _ = f.set_value("true");
                                prev_flag = None;
                            } else {
                                prev_flag.replace(flag.to_string());
                            }
                        } else {
                            self.floating.push(arg);
                        }
                    }
                }
            } else if arg.starts_with('-') {
                if !self.flags.contains_key(&arg[1..]) {
                    self.floating.push(arg);
                } else {
                    prev_flag.replace(arg[1..].to_string());
                }
            } else {
                self.floating.push(arg);
            }
        }

        // Change the flags into parsed_flags by changing the Rc<RefCell<T>> into Rc<T>
        // The needs map is used when an Rc is help by more than 1 value, therefore, it can't be
        // unwrapped. Needs stores the flags that couldn't be immediately unwrapped to be added
        // back when the other owner of the Rc is encountered and the Rc can be unwrapped.
        let mut needs = HashMap::new();
        for (flag, value) in self.flags.drain() {
            let addr = value.as_ptr() as usize;
            match Rc::try_unwrap(value) {
                Ok(cell) => {
                    let value = Rc::new(cell.into_inner());
                    if let Some(other_flag) = needs.remove(&addr) {
                        self.parsed_flags.insert(other_flag, Rc::clone(&value));
                    }
                    self.parsed_flags.insert(flag, value);
                }
                Err(_) => {
                    needs.insert(addr, flag);
                }
            }
        }
    }
}

#[derive(Debug)]
pub enum ArgValue {
    Usize(Value<usize>),
    String(Value<String>),
    Bool(Value<bool>),
}

impl ArgValue {
    fn set_value(&mut self, mut value: String) -> Result<(), Box<dyn std::error::Error>> {
        match self {
            ArgValue::Usize(v) => v.value = Some(value.parse()?),
            ArgValue::String(v) => v.value = Some(value),
            ArgValue::Bool(v) => {
                value.make_ascii_lowercase();
                match value.as_str() {
                    "1" | "t" | "true" => v.value = Some(true),
                    "0" | "f" | "false" => v.value = Some(false),
                    // This will always throw an error since parsing a bool only accepts "true"
                    // and "false"
                    _ => v.value = Some(value.parse()?),
                }
            }
        }
        Ok(())
    }

    pub fn usize_value(&self) -> Option<&usize> {
        match self {
            ArgValue::Usize(v) => v.value(),
            _ => panic!("value not a usize"),
        }
    }

    pub fn string_value(&self) -> Option<&str> {
        match self {
            ArgValue::String(v) => v.value(),
            _ => panic!("value not a string"),
        }
    }

    pub fn bool_value(&self) -> Option<&bool> {
        match self {
            ArgValue::Bool(v) => v.value(),
            _ => panic!("value not a bool"),
        }
    }

    #[allow(dead_code)]
    pub fn passed_value(&self) -> bool {
        match self {
            ArgValue::Usize(v) => v.passed(),
            ArgValue::String(v) => v.passed(),
            ArgValue::Bool(v) => v.passed(),
        }
    }
}

#[derive(Debug)]
struct Value<T: FromStr + std::fmt::Debug> {
    value: Option<T>,
    default: Option<T>,
}

impl<T: FromStr + std::fmt::Debug> Value<T> {
    fn new(default: Option<T>) -> Self {
        Value {
            value: None,
            default,
        }
    }

    fn value(&self) -> Option<&T> {
        self.value.as_ref().or(self.default.as_ref())
    }

    #[allow(dead_code)]
    fn passed(&self) -> bool {
        self.value.is_some()
    }
}
