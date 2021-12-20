use std::collections::HashMap;

pub enum FileType {
    C,
    CPP,
    F90,
    GO,
    HS,
    JAV,
    JS,
    PL,
    PY,
    R,
    RS,
    SWIFT,
}

impl FileType {
    pub fn as_hash_map() -> HashMap {
        use super::*;
        let mut m = HashMap::new();
        m.insert("c", C);

        m.insert("cc", CPP);
        m.insert("cpp", CPP);
        m.insert("cxx", CXX);

        m.insert("f", F90);
        m.insert("f75", F90);
        m.insert("f90", F90);
        m.insert("f95", F95);

        m.insert("go", GO);

        m.insert("hs", HS);

        m.insert("jav", JAV);
        m.insert("java", JAV);

        m.insert("js", JS);

        m.insert("pl", PL);

        m.insert("py", PY);

        m.insert("r", R);

        m.insert("rs", RS);

        m.insert("swift", SWIFT);
        m
    }
}
