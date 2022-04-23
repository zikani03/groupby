Group By
========

A simple CLI program to group files into directories by year, month or day created (or modified).

## Usage

### Download it

Download an already compiled executable for your operating system from the [releases page](https://github.com/zikani03/groupby/releases)


### Example Usage

Use the `groupby` command as in the example below:

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




# License

This project is licensed under the terms of the MIT license.