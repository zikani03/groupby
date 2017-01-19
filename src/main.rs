extern crate chrono;
extern crate clap;

mod lib;

use clap::{App,Arg};

use group_by::GroupBy;

use std::path::{Path, PathBuf};
use std::cmp;
use std::ffi::OsString;
use std::fs::{self, DirEntry, Metadata};
use std::io;
use std::result::Result;
use std::time::{SystemTime,Duration,UNIX_EPOCH};


use chrono::naive::date::NaiveDate;
use chrono::naive::datetime::NaiveDateTime;
use chrono::Datelike;
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
pub const SUBDIRECTORY_INNER: &'static str = "├───";
pub const SUBDIRECTORY_PIPE: &'static str =  "│";
pub const SUBDIRECTORY_LINK: &'static str =  "└───";


/// File Tm - file time data struct
///
/// Stores the particular year, month and day that file
/// was created including the absolute path to that file
#[derive(PartialEq, Eq, PartialOrd, Ord, Clone)]
struct FileTm {
    pub year: i32,
    pub month: u32,
    pub day: u32,
    pub file_path: OsString, 
}

#[derive(Clone, Copy)]
enum TimestampType {
    CREATED,
    MODIFIED
}

impl FileTm {

    pub fn from_created_date(file: DirEntry) -> Option<FileTm> {
        Self::from_dir_entry(file, TimestampType::CREATED)
    }
    
    pub fn from_modified_date(file: DirEntry) -> Option<FileTm> {
        Self::from_dir_entry(file, TimestampType::MODIFIED)
    }

    pub fn from_dir_entry(file: DirEntry, timestamp_type: TimestampType) -> Option<FileTm> {
        let meta_result = file.metadata();
        
        if meta_result.is_err() {
            return None;
        }

        let metadata = meta_result.ok().unwrap();
        
        let mut tm = metadata.created(); 

        match timestamp_type {
            TimestampType::CREATED => tm = metadata.created(),
            TimestampType::MODIFIED => tm = metadata.modified(),
        }

        if let Ok(timestamp_val) = tm {
            if let Some(d) = Self::systemtime_as_date(timestamp_val) {
                let val = FileTm {
                    file_path: file.file_name(),
                    year: d.year(),
                    month: d.month(),
                    day: d.day()
                };
                return Some(val);
            }
        }
        None
    }

    // Files return created -> SystemTime
    // We want SystemTime -> Date
    // We can't convert SystemTime directly to DateTime
    // So
    fn systemtime_as_date(tm: SystemTime) -> Option<NaiveDate> {
        if let Ok(dur) = tm.duration_since(UNIX_EPOCH) {
            let dt: NaiveDateTime = NaiveDateTime::from_timestamp(dur.as_secs() as i64, dur.subsec_nanos());
            let d: NaiveDate = dt.date();
            return Some(d);
        }
        None 
    }
}

fn print_btreemap<K,V> (tree: BTreeMap<K,V>, val_print: Fn<V>) {
    for (entry, values) in &tree {
        println!("{} {:?}", SUBDIRECTORY_LINK, entry);
        
        if () {
            print_btreemap(values);
        } else {
            for v in values.iter() {
                println!("{}   {} {:?}", SUBDIRECTORY_PIPE, SUBDIRECTORY_INNER, val_print(v));
            }
        }
        println!("{}", SUBDIRECTORY_PIPE);
    }
}

fn main() {
    let matches = App::new("GroupBy")
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
                        .help("Depth of the directory hierarchy. 1 = Year/, 2 = Year/Month/, 3 is Year/Month/Day/"))
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
    
    let dry_run: bool = matches.is_present("dry_run");

    let group_depth = i32::from_str_radix(
                                matches.value_of("depth").unwrap_or("2"),
                                10);

    let timestamp_type = if created {
         TimestampType::CREATED 
    } else {
        TimestampType::MODIFIED
    };
    
    if let Ok(file_times) = get_file_times(Path::new(dir_name), timestamp_type) {
        
        let files_grouped_by_month = file_times.group_by(|ref file_tm| {
                                    file_tm.month
                                });
        println!("{}", dir_name);
        
        print_btreemap(files_grouped_by_month, |ref v| { v.file_path });

        let files_grouped_by_year = files_grouped_by_month.group_by(|ref month_group| {
            if let Some(entry) = month_group.get(0) {
                return entry.year
            }
            return 0
        });
    }
        
}

/// Gets the FileTm for each file in the given directory
fn get_file_times<'r>(dir: &Path, timestamp_type: TimestampType) -> io::Result<Vec<FileTm>> {
    let mut file_times = Vec::<FileTm>::new();
    
    if dir.is_dir() {
        if let Ok(entries) = fs::read_dir(dir) { 
            for entry in entries {
                if let Ok(entry) = entry {
                    // We shouldn't care about directories yet
                    // if entry.path().is_file()  {
                    if let Some(ft) = FileTm::from_dir_entry(entry, timestamp_type) {
                        file_times.push(ft);
                    }
                } else {
                    println!("Could not get get entry");
                }
            }
        }
    }
    Ok(file_times)
}
