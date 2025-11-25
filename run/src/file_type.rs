// TODO: Add shell files (sh, zsh, bash, etc) and check if they have execute permissions

use crate::args::*;
use crate::run_funcs::*;

// Run funcs return Some(status) if the program should exit, otherwise, return None if it should
// continue.
pub type RunFunc = fn(&Args) -> Option<i32>;

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum FileType {
    BASH,
    C,
    CPP,
    F90,
    GO,
    HS,
    JAV,
    JL,
    JS,
    ML,
    PL,
    PY,
    R,
    RS,
    SWIFT,
}

impl FileType {
    pub fn can_have_multiple(self) -> bool {
        use FileType::*;
        match self {
            C | CPP | GO | JAV => true,
            _ => false,
        }
    }

    pub fn can_have_none(self) -> bool {
        use FileType::*;
        match self {
            GO | RS => true,
            _ => false,
        }
    }

    // Returns true if the file type (which usually has a compilation command) has an option to be
    // compiled and/or run with a single command (e.g., `go run` or `runghc`).
    pub fn can_run_only(self) -> bool {
        use FileType::*;
        match self {
            GO | HS => true,
            _ => false,
        }
    }

    pub fn run_func(self) -> RunFunc {
        use FileType::*;
        match self {
            BASH => run_bash,
            C => run_c,
            CPP => run_cpp,
            F90 => run_f90,
            GO => run_go,
            HS => run_hs,
            JAV => run_jav,
            JL => run_jl,
            JS => run_js,
            ML => run_ml,
            PL => run_pl,
            PY => run_py,
            R => run_r,
            RS => run_rs,
            SWIFT => run_swift,
            //TS => run_ts,
        }
    }

    /*
    pub fn prog(self) -> &'static str {
        use FileType::*;
        match self {
            BASH => run_bash,
            C => run_c,
            CPP => run_cpp,
            F90 => run_f90,
            GO => run_go,
            HS => run_hs,
            JAV => run_jav,
            JL => run_jl,
            JS => run_js,
            ML => run_ml,
            PL => run_pl,
            PY => run_py,
            R => run_r,
            RS => run_rs,
            SWIFT => "swift",
            TS => "deno",
        }
    }
    */

    pub fn default_args(self) -> &'static [&'static str] {
        use FileType::*;
        match self {
            BASH => &[],
            C => &["-std=c18"],
            CPP => &["-std=gnu++17"],
            F90 => &["gfortran"],
            GO => &[],
            HS => &[],
            JAV => &[],
            JL => &[],
            JS => &[],
            ML => &[],
            PL => &[],
            PY => &[],
            R => &[],
            RS => &["--edition=2021"],
            SWIFT => &[],
            //TS => &[],
        }
    }
}

impl std::str::FromStr for FileType {
    type Err = &'static str;
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        use FileType::*;
        match s.to_lowercase().as_str() {
            "bash" => Ok(BASH),
            "c" => Ok(C),
            "cpp" | "cc" | "cxx" => Ok(CPP),
            "f77" | "f90" | "f95" => Ok(F90),
            "go" => Ok(GO),
            "hs" => Ok(HS),
            "jav" | "java" => Ok(JAV),
            "jl" => Ok(JL),
            "js" => Ok(JS),
            "ml" => Ok(ML),
            "pl" | "perl" => Ok(PL),
            "py" => Ok(PY),
            "r" | "R" | "Rscript" => Ok(R),
            "rs" => Ok(RS),
            "swift" => Ok(SWIFT),
            _ => Err("invalid file type"),
        }
    }
}
