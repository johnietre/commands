use clap::Parser;
use std::path::PathBuf;

#[derive(Parser, Default, Debug)]
/// Run programs with ease.
pub struct Args {
    /// The names of the files passed.
    pub file_names: Vec<PathBuf>,

    #[clap(short, long = "output")]
    /// The name of the output file, if applicable. If no output name is specified and the temp
    /// flag isn't passed, the output name is set to the name of the file without the extension
    /// (whatever follows the final '.'); if on Windows, '.exe' is appended.
    pub output_name: Option<PathBuf>,

    #[clap(short, long)]
    /// Only compile the program, if applicable.
    pub compile_only: bool,

    #[clap(long = "comp_arg")]
    /// Arguments for compilation, if applicable.
    pub comp_args: Vec<String>,

    #[clap(long = "exec_arg")]
    /// Arguments for execution, if applicable.
    pub exec_args: Vec<String>,

    #[clap(short = 't', long = "type", parse(try_from_str))]
    /// The type of file to use when running the program. If not specified, it is extrapolated from
    /// the first file name.
    pub file_type: Option<crate::file_type::FileType>,

    #[clap(short, long)]
    /// Run the program using a run only command, if applicable, otherwise, delete the output files
    /// afterwards. An example of a run only command is `go run` or `runghc`. This does has no
    /// effect on languages without compilation.
    pub no_out: bool,

    #[clap(long)]
    /// Compile the output into the system's temp directory.
    pub temp: bool,

    #[clap(short, long)]
    /// Pass default compilation flags, if applicable.
    pub default_flags: bool,

    #[clap(long)]
    /// Run without timing anything.
    pub no_time: bool,

    #[clap(long)]
    /// Keep all output files, if applicable. An example is keeping all the intermediate output
    /// files during OCaml compilation (.cmi, .cmx, .o).
    pub keep_all_out: bool,

    #[clap(long)]
    /// Compile to Web Assembly. Sets the compile_only flag to true.
    pub wasm: bool,

    #[clap(long)]
    /// The program to compile or run programs with.
    pub program: Option<String>,

    #[clap(long)]
    pub parse_includes: bool,

    #[clap(short, long)]
    /// Run the rest of the arguments as a Bash command. Must be the first arg (second if preceded
    /// by --no-time).
    pub bash: bool,

    #[clap(short, long)]
    /// Continuously watch the input file(s) and restart the program on each write.
    pub watch: bool,
}

impl Args {
    // Returns the program passed to the args or returns the altnerative passed
    pub fn prog<S: ToString>(&self, alt: S) -> String {
        self.program.as_ref().cloned().unwrap_or(alt.to_string())
    }
}
