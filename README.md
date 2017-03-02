data-profiler ![travis-status](https://travis-ci.org/giagiannis/data-profiler.svg?branch=master)
=============
data-profiler is a project used to perform data profiling in an operator agnostic way.

Installation
------------

```bash
git clone https://github.com/giagiannis/data-profiler
cd data-profiler
go install ./...
```

Usage
-----
use the data-profiler-utils binary to use this project

Example 1: estimating similarities using the Bhattacharyya coefficient
```bash
data-profiler-utils similarities -i <INPUT DIR> -o <OUTPUT FILE> -l <LOG FILE> -opt tree.scale=0.5
```

Example 2: estimating an approximate similarity matrix with 100 datasets using the Bhattacharyya coefficient 
```bash
data-profiler-utils similarities -i <INPUT DIR> -o <OUTPUT FILE> -l <LOG FILE> -opt tree.scale=0.5
```


License
-------
Apache License v2.0 (see [LICENSE](LICENSE) file for more)
