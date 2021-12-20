use std::ptr::NonNull;

fn main() {
    let test = Test { v: NonNull::new(Box::leak(Box::new(32))).unwrap() };
    test.get();
}

struct Test {
    v: NonNull<i32>,
}

impl Test {
    fn get(&self) -> &i32 {
        unsafe { self.v.as_ref() }
    }
}

impl Drop for Test {
    fn drop(&mut self) {
        unsafe { Box::from_raw(self.v.as_ptr()); }
    }
}
