Group By
========

A simple CLI program to group files into directories by year, month or day
by date created or modified.

## Usage

```text
groupby [OPTIONS]

Usage of groupby:
  -DIRECTORY string
                Directory containing files to group
  -a            Include hidden files and directories (starting with .)
  -created
                Group files by the date they were created
  -day
                Alias for --depth=3, overrides --depth
  -depth int
                How deep to create the directory hierarchy (default 1)
  -dry-run
                Only show the output of how the files will be grouped
  -flatten
                Flatten the created directory tree folders
  -modified
                Group files by the date they were modified (default true)
  -month
                Alias for --depth=2, overrides --depth
  -p            Only show the output of how the files will be grouped (shorthand)
  -preview
                Only show the output of how the files will be grouped
  -v            Show verbose output (default true)
  -verbose
                Show verbose output (default true)
  -version
                Show the program version and exit
  -year
                Alias for --depth=1, overrides --depth
```

## Example Usage

Let's say you have a gazillion files in `my_messy_directory` and you'd like to
group them by the date they were created.

You should be able to use the following command

```bash
$ groupby -day -DIRECTORY=./groupby
```

This will group your files into year subdirectories and then month subdirectories
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
$ $GOPATH/bin/groupby --help
```
 
## LICENSE

MIT

---

Copyright (c) 2017, Zikani 