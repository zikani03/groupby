extern crate chrono;
extern crate clap;

pub mod group_by;
pub mod file_entry;

use clap::{App, Arg };

use file_entry::{FileEntryTree,TimestampType, TreeDepth};

pub const SUBDIRECTORY_INNER: &'static str = "├───";
pub const SUBDIRECTORY_PIPE: &'static str = "│";
pub const SUBDIRECTORY_LINK: &'static str = "└───";

/// Usage
///
/// groupby [options] DIRECTORY
///
/// ## Options
///
/// ```
/// -c   --created Group files by the date they were created
/// -m   --modified Group files by the date they were modified
/// -n   --dry-run Show the output of how the files will be grouped
/// -d   --depth N How deep to create the directory hierarchy (maximum: 3)
///                corresponding to 1 - year, 2 - month, 3 - day
/// -D   --group-dirs Move directories into groups as well - by default only
///                   regular files are grouped
/// -R   --recurse Group files in subdirectories
/// -h   --help Show the help information and exit
/// -v   --verbose Show verbose output
///      --version Show the program version
/// ```
///
/// ## Examples
///
/// ```
/// $ groupby -c -D -R -v -d 3 ./my_directory
/// $ groupby --modified -DRv -d 3 ./my_directory
/// ```
///
/// 2016
/// ├─── Jan
/// │    └── 01
/// │        └── my_file.txt
/// └── Feb
///     └── 01
///         └── my_file_2.txt
///
fn main() {
    // Create the clap application
    let matches = App::new("groupby")
        .version("0.1")
        .author("Zikani Nyirenda Mwase")
        .about("Group files into directories")
        .arg(Arg::with_name("DIRECTORY")
            .required(true)
            .help("The directory to check for files to group"))
        .arg(Arg::with_name("created")
            .short("c")
            .long("created")
            .takes_value(false)
            .help("Group files by date they were created"))
        .arg(Arg::with_name("modified")
            .short("m")
            .long("modified")
            .takes_value(false)
            .help("Group files by date they were modified"))
        .arg(Arg::with_name("dry_run")
            .short("n")
            .long("dry-run")
            .takes_value(false)
            .help("Perform a dry-run - doesn't actually move files to subdirectories"))
        .arg(Arg::with_name("depth")
            .short("d")
            .long("depth")
            .takes_value(true)
            .help("Depth of the directory hierarchy.
                   1 = Year
                   2 = Year and Month
                   3 = Year, Month and Day"))
        .arg(Arg::with_name("directories")
            .short("D")
            .long("directories")
            .takes_value(false)
            .help("Group files as well as directories"))
        .arg(Arg::with_name("recurse")
            .short("R")
            .long("recurse")
            .takes_value(false)
            .help("Look for files to group in subdirectories"))
        .arg(Arg::with_name("verbose")
            .short("v")
            .long("verbose")
            .takes_value(false)
            .help("Show verbose output"))
        .arg(Arg::with_name("version")
            .long("version")
            .takes_value(false)
            .help("Show the program version and exit"))
        .get_matches();

    let dir_name = matches.value_of("DIRECTORY").unwrap();

    // Defaults to created and not modified
    let created: bool = matches.is_present("created");
    let modified: bool = matches.is_present("modified");

    if created && modified {
        println!("You cannot specify both -c(reated) and -m(odified) please use one or the other");
        std::process::exit(-1)
    }

    let dry_run: bool = true; // matches.is_present("dry_run");

    let group_depth;
    
    match i32::from_str_radix(matches.value_of("depth").unwrap_or("1"), 10) {
        Ok(1) => group_depth = TreeDepth::YEAR,
        Ok(2) => group_depth = TreeDepth::MONTH,
        Ok(3) => group_depth = TreeDepth::DAY,
        Ok(_) => group_depth = TreeDepth::YEAR,
        Err(_) => group_depth = TreeDepth::YEAR
    };

    let timestamp_type = if created {
        TimestampType::CREATED
    } else {
        TimestampType::MODIFIED
    };
    
    match FileEntryTree::new(dir_name, timestamp_type, group_depth) {
        Some(grouped_entries) => {
            if dry_run {
                println!("{}", grouped_entries)
            } else {
                print!("Failed to create grouped tree");
            }
        },
        None => ()
    }
}
