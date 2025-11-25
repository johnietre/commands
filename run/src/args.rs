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

    #[clap(long)]
    /// Configure when/how to restart a program automatically.
    pub restart_conf: Option<RestartConfig>,

    #[clap(long)]
    /// Where to pipe the execution command's stdout to. Not being set defaults to piping to the
    /// parent's stdout. Passing an empty string discards stdout. Setting it to something creates
    /// (or possibly appends to) the file specified. Without setting the appropriate append flag,
    /// the file is overwritten if it already exists.
    pub stdout: Option<String>,

    #[clap(long)]
    /// Append the stdout output to the file specified. No effect if stdout is not piped to a file.
    /// Creates the file is it doesn't already exist.
    pub append_stdout: Option<bool>,

    #[clap(long)]
    /// Where to pipe the execution command's stderr to. Not being set defaults to piping to the
    /// parent's stderr. Passing an empty string discards stderr. Setting it to something creates
    /// (or possibly appends to) the file specified. Without setting the appropriate append flag,
    /// the file is overwritten if it already exists.
    pub stderr: Option<String>,

    #[clap(long)]
    /// Append the stderr output to the file specified. No effect if stderr is not piped to a file.
    /// Creates the file is it doesn't already exist.
    pub append_stderr: Option<bool>,

    #[clap(long)]
    /// Where to read the execution command's stdin from. Not being set default to piping from the
    /// parent's stdin. Passing an empty string pipes from null (i.e., closes input). Settings it
    /// to something opens the file specified to read stdin from.
    pub stdin: Option<String>,

    #[clap(long)]
    pub comp_stdout: Option<String>,

    #[clap(long)]
    pub append_comp_stdout: Option<bool>,

    #[clap(long)]
    pub comp_stderr: Option<String>,

    #[clap(long)]
    pub append_comp_stderr: Option<bool>,

    #[clap(long)]
    pub comp_stdin: Option<String>,
}

impl Args {
    // Returns the program passed to the args or returns the altnerative passed
    pub fn prog<S: ToString>(&self, alt: S) -> String {
        self.program.as_ref().cloned().unwrap_or(alt.to_string())
    }

    pub fn stdout_name(&self, comp: bool) -> Option<OutputName> {
        if comp {
            Some(OutputName::from_parts(self.comp_stdout.clone()?, self.append_comp_stdout))
        } else {
            Some(OutputName::from_parts(self.stdout.clone()?, self.append_stdout))
        }
    }

    pub fn stderr_name(&self, comp: bool) -> Option<OutputName> {
        if comp {
            Some(OutputName::from_parts(self.comp_stderr.clone()?, self.append_comp_stderr))
        } else {
            Some(OutputName::from_parts(self.stderr.clone()?, self.append_stderr))
        }
    }

    pub fn stdin_name(&self, comp: bool) -> Option<OutputName> {
        if comp {
            Some(OutputName::from_parts(self.comp_stdin.clone()?, None))
        } else {
            Some(OutputName::from_parts(self.stdin.clone()?, None))
        }
    }
}

#[derive(Serialize, Deserialize)]
pub struct RestartConfig {
    //#[serde(default)]
    //pub email: Option<EmailConfig>,
    #[serde(default)]
    pub after_exit: Option<Program>,

    #[serde(default)]
    pub fail_statuses: Option<Vec<i32>>,

    #[serde(default)]
    pub cont_statuses: Option<Vec<i32>>,

    #[serde(default)]
    pub max_restarts: u32,

    #[serde(default)]
    pub delay: i64,
}

pub struct Program {
    pub name: String,

    #[serde(default)]
    pub args: Option<Vec<String>>,

    #[serde(default)]
    pub must_finished: Option<bool>,

    #[serde(default)]
    pub fail_statuses: Option<Vec<i32>>,

    #[serde(default)]
    pub cont_statuses: Option<Vec<i32>>,

    #[serde(default)]
    pub stdout: Option<String>,

    #[serde(default)]
    pub stderr: Option<String>,

    #[serde(default)]
    pub stdin: Option<String>,
}

pub enum OutputName {
    Name(String),
    NameOpts { name: String, append: bool },
}

impl OutputName {
    pub fn from_parts(name: String, append: Option<bool>) -> Self {
        match append {
            Some(append) => OutputName { name, append },
            None => OutputName::Name(name),
        }
    }

    pub fn name(&self) -> &str {
        match self {
            OutputName::Name(name) => name.as_str(),
            OutputName::NameOpts { name, .. } => name.as_str(),
        }
    }

    pub fn append(&self) -> &bool {
        match self {
            OutputName::Name(_) => false,
            OutputName::NameOpts { append, .. } => append,
        }
    }
}

#[allow(dead_code)]
#[derive(Serialize, Deserilaize)]
pub struct EmailConfig {
    pub from: String,
    pub to: String,
    pub subject: Option<String>,
    pub username: Option<String>,
    pub password: Option<String>,
    pub username_env: Option<String>,
    pub password_env: Option<String>,
}
