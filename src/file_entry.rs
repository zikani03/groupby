extern crate chrono;

use chrono::naive::date::NaiveDate;
use chrono::naive::datetime::NaiveDateTime;
use chrono::Datelike;

use std::mem;

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

/// Trait for converting a value to a chrono::NaiveDate
trait ToNaiveDate {
    fn to_naive_date(&self) -> Option<NaiveDate>;
}

/// Implements the `ToNaiveDate` trait for rust's `SystemTime` struct
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

// Timestamp Type for a file entry
#[derive(Clone, Copy)]
pub enum TimestampType {
    CREATED,
    MODIFIED,
}

#[derive(Clone, Copy, PartialEq, Debug)]
pub enum TreeDepth {
    YEAR = 1,
    MONTH = 2,
    DAY = 3
}

impl TreeDepth {
    pub fn deeper(self) -> Self {
        match self {
            TreeDepth::YEAR => TreeDepth::MONTH,
            TreeDepth::MONTH => TreeDepth::DAY,
            TreeDepth::DAY => TreeDepth::DAY
        }
    }
}

/// File Tm - file time data struct
///
/// Stores the particular year, month and day that file
/// was created including the absolute path to that file
#[derive(PartialEq, Eq, PartialOrd, Ord, Clone)]
struct FileEntry {
    pub year: i32,
    pub month: u32,
    pub day: u32,
    pub file_name: String
}

impl FileEntry {
    fn from_created_date(file: DirEntry) -> Option<Self> {
        Self::from_dir_entry(file, TimestampType::CREATED)
    }

    fn from_modified_date(file: DirEntry) -> Option<Self> {
        Self::from_dir_entry(file, TimestampType::MODIFIED)
    }

    fn month_entry(&self) -> FileEntry {
        FileEntry {
            file_name: format!("{}", self.month),
            month: self.month,
            day: self.day,
            year: self.year
        }
    }

    fn year_entry(&self) -> FileEntry {
        FileEntry {
            file_name: format!("{}", self.year),
            month: self.month,
            day: self.day,
            year: self.year
        }
    }

    fn day_entry(&self) -> FileEntry {
        FileEntry {
            file_name: format!("{}", self.day),
            month: self.month,
            day: self.day,
            year: self.year
        }
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

    fn from_name(name: String) -> Option<Self> {
        let path = Path::new(name.as_str());

        if path.exists() {   
            return Some(FileEntry {
                year: 0,
                month: 0,
                day: 0,
                file_name: name.clone()
            })
        }
        None
    }
}

pub struct FileEntryTree {
    root: EntryNode,
    timestamp_type: TimestampType,
    no_entries: i32
}

type EntryLink =Option<Box<EntryNode>>;

struct EntryNode {
    value: FileEntry,
    depth: TreeDepth,
    next: EntryLink,
    children: EntryLink 
}

impl FileEntryTree {
    /// Groups entries based on the year, month and day in the FileEntry
    /// @param entries
    /// @param depth - the depth is how deep the group tree will be
    ///                0 - entries will be grouped by year, ie. group will be 2 levels deep
    ///                1 - entries will be grouped by year then month, ie. group will be 2 levels deep
    ///                2 - entries will be grouped by year, then month then day
    pub fn new(dir_name: &str, timestamp_type: TimestampType, depth: TreeDepth) -> Option<Self> {
        let mut tree = FileEntryTree {
            root: EntryNode {
                value: FileEntry::from_name(String::from(dir_name)).unwrap(),
                depth: depth,
                next: None,
                children: None
            },
            timestamp_type: timestamp_type,
            no_entries: 1
        };
        
        Some(tree)
    }

    /// Reads the entries from the file system using Self.root.file_name()
    /// and builds the final tree 
    fn build(&mut self) -> &Self {
        let file_path = self.root.file_name();
        let ref root_path = Path::new(file_path.as_str());
        match FileEntry::read_entries(root_path, self.timestamp_type) {
            Ok(entries) => {
                for e in entries {
                    self.add_entry(e)
                }
            },
            Err(e) => panic!("Failed to read entries from directory {:?}", e)
        };
        self
    }

    fn add_entry(&mut self, entry: FileEntry) {
        match mem::replace(&mut self.root.children, None) {
            Some(mut boxed_node) => {
                boxed_node.add_child(entry);
            },
            None => {
                let mut year_node = EntryNode::new(entry.year_entry(), TreeDepth::YEAR);
                year_node.add_child(entry);
                self.root.next = Some(Box::new(year_node));
            }
        };
        self.no_entries += 1;
    }

    pub fn size(&self) -> i32 {
        self.no_entries
    }
}

impl fmt::Display for FileEntryTree {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.root.value.file_name);
        write!(f, "├───")
    }
}

impl EntryNode {

    fn new(entry: FileEntry, depth: TreeDepth) -> Self {
        EntryNode {
            value: entry,
            depth: depth,
            next: None,
            children: None
        }
    }

    fn file_name(&self) -> String {
        let ref val = self.value;
        val.file_name.clone()
    }

    fn depth(&self) -> TreeDepth {
        self.depth
    }
    
    fn next_depth(&self) -> TreeDepth {
        match self.children {
            Some(ref boxed_node) => {
                TreeDepth::deeper(boxed_node.depth())
            },
            None => TreeDepth::deeper(self.depth)
        }
    }

    fn children(&self) -> Option<&EntryNode> {
        if let Some(ref children) = self.children {
            return Some(&children)
        }
        None
    }

    fn set_next(&mut self, node: EntryNode) {
        self.next = Some(Box::new(node));
    } 

    fn add_child(&mut self, entry: FileEntry) {
        // Check if we need to go deeper
        // the parent_depth is how deep a tree we're allowed to create
        // next_depth allows us to stop once we've reached an acceptable depth? 
        let parent_depth = self.depth;
        let next_depth;
        match TreeDepth::deeper(parent_depth) {
            TreeDepth::DAY => next_depth = TreeDepth::DAY,
            TreeDepth::MONTH => next_depth = TreeDepth::YEAR,
            TreeDepth::YEAR  => next_depth = TreeDepth::DAY
        }

        let mut next_node;
        match next_depth {
            TreeDepth::YEAR => {
                let mut year_node = Box::new(EntryNode::new(entry.year_entry(), self.next_depth()));
                year_node.add_child(entry);
                next_node = year_node;
            },
            TreeDepth::MONTH => {
                let mut month_node = Box::new(EntryNode::new(entry.month_entry(), self.next_depth()));
                month_node.add_child(entry);
                next_node = month_node;
            },
            TreeDepth::DAY => {
                next_node = Box::new(EntryNode::new(entry, TreeDepth::DAY));
            }
        };

        match mem::replace(&mut self.children, None) {
            Some(mut boxed_node) => {
                boxed_node.children = Some(next_node);
            },
            None => {
               self.children = Some(next_node);
            }
        };
    }
}

#[cfg(test)]
mod test {
    use super::*;
    
    #[test]
    fn test_entry_node_depth() {
        let entry_node = EntryNode::new(FileEntry {
                year: 2017,
                month: 7,
                day: 1,
                file_name: String::from("/etc")
            },
            TreeDepth::YEAR);
        
        assert_eq!(entry_node.file_name(), String::from("/etc"));
        assert_eq!(entry_node.depth(), TreeDepth::YEAR);
        assert_eq!(entry_node.next_depth(), TreeDepth::MONTH);
    }
    
    #[test]
    fn test_entry_node_add_child() {
        let mut parent_node = EntryNode::new(FileEntry {
                year: 2017,
                month: 7,
                day: 1,
                file_name: String::from("/etc")
            },
            TreeDepth::YEAR);

        let child_entry = FileEntry {
            year: 2017,
            month: 7,
            day: 1,
            file_name: String::from("passwd")
        };
        
        // Since the max for the parent node is a MONTH
        // we should expect the following structure 
        // /etc
        //  ├───2017
        //     └── passwd
        parent_node.add_child(child_entry);

        let c = parent_node.children().unwrap();
        
        assert_eq!(c.file_name(), String::from("2017"));
        assert_eq!(c.children().unwrap().file_name(), String::from("passwd"));
    }
    
    #[test]
    fn test_entry_node_add_deeper_child() {
        let mut parent_node = EntryNode::new(FileEntry {
                year: 2017,
                month: 7,
                day: 1,
                file_name: String::from("/etc")
            },
            TreeDepth::MONTH);

        let child_entry = FileEntry {
            year: 2017,
            month: 7,
            day: 1,
            file_name: String::from("passwd")
        };
        
        // Since the max for the parent node is a MONTH,
        // we should expect the following structure 
        // /etc
        //  ├───2017
        //      ├───7
        //          └── passwd
        parent_node.add_child(child_entry);

        let c = parent_node.children().unwrap();
        
        assert_eq!(parent_node.file_name(), String::from("/etc"));
        assert_eq!(c.file_name(), String::from("2017"));
        assert_eq!(c.children().unwrap().file_name(), String::from("7"));
        assert_eq!(c.children().unwrap().children().unwrap().file_name(), String::from("passwd"));
    }
    
    #[test]
    fn test_create_file_entry_tree() {
        let tree_optional = FileEntryTree::new(".", TimestampType::CREATED, TreeDepth::YEAR);

        assert_eq!(tree_optional.is_some(), true);
    }


    #[test]
    fn test_add_entry() {
        let mut tree = FileEntryTree::new(".", TimestampType::CREATED, TreeDepth::YEAR).unwrap();

        tree.add_entry(FileEntry {
            year: 2017,
            month: 7,
            day: 1,
            file_name: String::from("main.rs")
        });

        assert_eq!(tree.size(), 2);
    }
}