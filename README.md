Group By
========

A simple CLI program to group files into directories by year, month or day created (or modified).

## Usage

```text
groupby [OPTIONS]

Usage of groupby:
  -a            Include hidden files and directories (starting with .)
  -copy-only
                Only copy files, do not move them
  -created
                Group files by the date they were created (default)
  -d DIRECTORY
                Directory containing files to group
  -day
                Group by year, month and then day
  -dry-run
                Only show the output of how the files will be grouped
  -flatten
                Flatten the created directory tree folders
  -ignore-directories
                Ignore directories and only group files
  -modified
                Group files by the date they were modified (default true)
  -month
                Group by year, and then month
  -o DIRECTORY
                Directory to move grouped files to
  -p            Only show the output of how the files will be grouped (shorthand)
  -preview
                Only show the output of how the files will be grouped
  -v            Show verbose output
  -verbose
                Show verbose output
  -version
                Show the program version and exit
  -year
                Group by year only
```

## Example Usage

Once installed, you should be able to use the `groupby` as in the example, below:

```bash
$ groupby -day -d=./groupby
```

This will group your files into year, month and then day subdirectories
so that it looks like This

```
./groupby
└── 2017
   ├── July
      └── 15
         └── LICENSE
   └── August
      └── 21
         ├── README.md
         └── groupby.go
```

## Installation

```bash
$ go get -u github.com/zikani03/groupby
$ $GOPATH/bin/groupby -h
```
 
## LICENSE

MIT

---

Copyright (c) 2018, Zikani 