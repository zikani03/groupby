Group By
========

> **THIS IS A WORK IN PROGRESS**
> **Note:** Still learning Rust so code will be crappy. ;)

A simple CLI program to group files into directories by year, month or day
by date created or modified.

## Example Usage

Let's say you have a gazillion files in `my_messy_directory` and you'd like to
group them by the date they were created.

You should be able to use the following command

```bash
$ groupby --created --depth 2 --dry-run my_messy_directory
```

This will group your files into year subdirectories and then month subdirectories
so that it looks like This

```
my_messy_directory
└───2016
|   └─── January
|   └─── February
|   |
|   .
|   .
|   └─── December
|
└───2015
    └─── January

```

## Installation

@TODO

## Building

```bash
$ git clone https://github.com/zikani03/groupby
$ cd groupby
$ cargo build
```

## LICENSE

MIT

===

Copyright (c) 2017, Zikani 