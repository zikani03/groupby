extern crate chrono;

use chrono::naive::date::NaiveDate;
use chrono::naive::datetime::NaiveDateTime;
use chrono::Datelike;

use std::borrow::BorrowMut;
use std::cmp::Ord;
use std::clone::Clone;
use std::collections::BTreeMap;
use std::fmt;
use std::ffi::OsString;
use std::fs::{self, DirEntry, Metadata};
use std::io;
use std::io::{Error, ErrorKind};
use std::path::{Path, PathBuf};
use std::result::Result;
use std::time::{SystemTime, Duration, UNIX_EPOCH};

pub const SUBDIRECTORY_INNER: &'static str = "├───";
pub const SUBDIRECTORY_PIPE: &'static str = "│";
pub const SUBDIRECTORY_LINK: &'static str = "└───";

// Timestamp Type for a file entry
#[derive(Clone, Copy)]
pub enum TimestampType {
    CREATED,
    MODIFIED,
}

/// File Tm - file time data struct
///
/// Stores the particular year, month and day that file
/// was created including the absolute path to that file
#[derive(PartialEq, Eq, PartialOrd, Ord, Clone)]
pub struct FileEntry {
    pub year: i32,
    pub month: u32,
    pub day: u32,
    pub file_name: String
}

trait ToNaiveDate {
    fn to_naive_date(&self) -> Option<NaiveDate>;
}

impl ToNaiveDate for SystemTime {
    // Files return created -> SystemTime
    // We want SystemTime -> Date
    // We can't convert SystemTime directly to DateTime
    fn to_naive_date(&self) -> Option<NaiveDate> {
        match self.duration_since(UNIX_EPOCH) {

            Ok(dur) => {
                let dt = NaiveDateTime::from_timestamp(dur.as_secs() as i64,
                                                       dur.subsec_nanos());
                let d: NaiveDate = dt.date();
                Some(d)
            },
            Err(_) => None
        }
    }
}


impl FileEntry {
    fn from_created_date(file: DirEntry) -> Option<Self> {
        Self::from_dir_entry(file, TimestampType::CREATED)
    }

    fn from_modified_date(file: DirEntry) -> Option<Self> {
        Self::from_dir_entry(file, TimestampType::MODIFIED)
    }

    fn from_dir_entry(file: DirEntry, timestamp_type: TimestampType) -> Option<FileEntry> {
        let meta_result = file.metadata();

        if meta_result.is_err() {
            return None;
        }

        let metadata = meta_result.ok().unwrap();

        let mut tm;
        match timestamp_type {
            TimestampType::CREATED => tm = metadata.created(),
            TimestampType::MODIFIED => tm = metadata.modified(),
        }

        // TODO: Clean up this code using nice stuff from Optional like map, or_else..
        if let Ok(timestamp_val) = tm {
            if let Some(naive_date) = timestamp_val.to_naive_date() {

                let file_name = file.file_name()
                                    .to_os_string()
                                    .into_string()
                                    .unwrap();

                let val = FileEntry {
                    file_name: file_name,
                    year:  naive_date.year(),
                    month: naive_date.month(),
                    day:   naive_date.day(),
                };
                // have to put a return here because we're inside
                // an expression that MUST return ()
                return Some(val);
            }
        }
        None
    }

    /// Gets the FileEntry items for each file a the given directory
    fn read_entries<'r>(dir: &Path, timestamp_type: TimestampType) -> io::Result<Vec<FileEntry>> {
        let mut file_entries = Vec::<Self>::new();

        if dir.is_dir() {
            if let Ok(entries) = fs::read_dir(dir) {
                for entry in entries {
                    if let Ok(entry) = entry {
                        // We shouldn't care about directories yet
                        // if entry.path().is_file()  {
                        if let Some(ft) = FileEntry::from_dir_entry(entry, timestamp_type) {
                            file_entries.push(ft);
                        }
                    } else {
                        println!("Could not get get entry");
                    }
                }
            } else {
                return Err(Error::new(ErrorKind::Other, "Failed to read contents of the directory"));
            }
        } else {
            return Err(Error::new(ErrorKind::Other, "Path must be a path to a directory"));
        }
        Ok(file_entries)
    }

    /// Groups entries based on the year, month and day in the FileEntry
    /// @param entries
    /// @param depth - the depth is how deep the group tree will be
    ///                0 - entries will be grouped by year, ie. group will be 2 levels deep
    ///                1 - entries will be grouped by year then month, ie. group will be 2 levels deep
    ///                2 - entries will be grouped by year, then month then day
    pub fn group_entries_by_date(path: &Path, timestamp_field: TimestampType, depth: i32)
                                -> Option<GroupedEntryTree> {

        let root_dir_name  = path.file_name()
                                    .map(|s| s.to_os_string())
                                    .unwrap()
                                    .into_string()
                                    .unwrap();

        match FileEntry::read_entries(path, timestamp_field) {
            Ok(entries) => {
                let mut tree = GroupedEntryTree::new(root_dir_name, entries.len() as i32);

                /*
                let files_grouped_by_year = entries.group_by(|ref entry| entry.year);

                println!("{}", dir_name);

                print_btreemap(files_grouped_by_year, |ref v| v.file_path);
                */
                Some(tree)
            },
            Err(e) => None
        }
    }
}

pub struct GroupedEntryTree {
    root: EntryNode,
    no_entries: i32
}

type EntryLink = Box<Option<EntryNode>>;

struct EntryNode {
    value: String,
    depth: i32,
    next: EntryLink,
    children: EntryLink 
}

impl GroupedEntryTree {
    fn new(root_name: String, no_entries: i32) -> Self {
        GroupedEntryTree {
            root: EntryNode {
                value: root_name,
                depth: 0,
                next: Box::new(None),
                children: Box::new(None)
            },
            no_entries: no_entries
        }
    }

    fn add_entry(&self, entry: FileEntry) {
        unimplemented!()
        /*
        match self.root.children.borrow_mut() {
            Some(node) => {
                self.root.add_entry_to_child(entry)
            },
            None => {
                self.children = Box::new(Option(EntryNode {
                    value: entry.file_name().clone(),
                    depth: 1,
                    next: Box::new(None),
                    children: Box::new(None)
                }))
            }
        }
        */
    }


    pub fn write_to_disk(&self) {
        unimplemented!()
        /*
        // iterate over only the internal nodes of the tree
        for node in self.internal_nodes() {
            // create_directory(node)
        }

        // iterate over only the internal nodes of the tree
        for node in self.to_iter() {
            match create_file(node) {
                Ok(file) => ()
                Err(e) => ()
            }
        }
        */
    }
}

impl fmt::Display for GroupedEntryTree {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.root.value);
        write!(f, "├───")
    }
}

impl EntryNode {

    fn add_entry(&self, entry: FileEntry) {
        unimplemented!()
        /*
        match self.children().take() {
            Some(node) => {
                
            }
        }
        */
    }
}

/*
fn print_btreemap<K, V>(tree: ref BTreeMap<K, V>, val_print: Fn<V> -> String) {
    for (entry, values) in &tree {
        println!("{} {:?}", SUBDIRECTORY_LINK, entry);
        for v in values.iter() {
            println!("{}   {} {:?}",
                        SUBDIRECTORY_PIPE,
                        SUBDIRECTORY_INNER,
                        val_print(v));
        }
        println!("{}", SUBDIRECTORY_PIPE);
    }
}
*/