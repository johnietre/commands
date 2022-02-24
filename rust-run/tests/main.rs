#![allow(unused_imports)]
use std::collections::hash_map::DefaultHasher;
use std::ffi::CStr;
use std::hash::{Hash, Hasher};
use std::os::raw::c_char;

extern "C" {
    fn tmpnam(buf: *mut c_char) -> *const c_char;
}

fn main() {
    unsafe {
        let ptr = tmpnam(0 as *mut c_char);
        println!("{}", CStr::from_ptr(ptr).to_string_lossy());
    }
    let mut hasher = DefaultHasher::new();
    hasher.write(b"johnie rodgers");
    println!("{:x}", hasher.finish());
}
