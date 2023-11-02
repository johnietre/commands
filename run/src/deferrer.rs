use std::collections::LinkedList;
use std::sync::{
    atomic::{AtomicBool, Ordering},
    Arc,
};

pub type DeferFunc = Box<dyn Fn() + Send + Sync + 'static>;

pub struct Defer {
    f: DeferFunc,
    should_run: Arc<AtomicBool>,
}

impl Defer {
    fn new(f: DeferFunc) -> Self {
        Self {
            should_run: Arc::new(AtomicBool::new(true)),
            f,
        }
    }

    pub fn run(&self) {
        if self.should_run.load(Ordering::SeqCst) {
            (self.f)();
        }
    }

    pub fn should_run(&self, b: bool) {
        self.should_run.store(b, Ordering::SeqCst);
    }
}

// The bool tells whether it has executed or not
// In order to still run when original function scope ends but it's in multiple threads,
// wrap in an Arc and use Weak for the threads
#[derive(Default)]
pub struct Deferrer(LinkedList<Arc<Defer>>, bool);

impl Deferrer {
    pub const fn new() -> Self {
        Self(LinkedList::new(), false)
    }

    pub fn push_back(&mut self, f: DeferFunc) -> Arc<Defer> {
        let f = Arc::new(Defer::new(f));
        self.0.push_back(Arc::clone(&f));
        f
    }

    pub fn push_front(&mut self, f: DeferFunc) -> Arc<Defer> {
        let f = Arc::new(Defer::new(f));
        self.0.push_front(Arc::clone(&f));
        f
    }

    pub fn force_run(&mut self) {
        if !self.1 {
            self.0.iter().for_each(|d| d.run());
            self.1 = true;
        }
    }
}

impl Drop for Deferrer {
    fn drop(&mut self) {
        self.force_run();
    }
}
