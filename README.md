data-profiler [![Build Status](https://travis-ci.org/giagiannis/data-profiler.svg?branch=master)](https://travis-ci.org/giagiannis/data-profiler) [![goreport](https://goreportcard.com/badge/github.com/giagiannis/data-profiler)](https://goreportcard.com/report/github.com/giagiannis/data-profiler) [![Coverage Status](https://coveralls.io/repos/github/giagiannis/data-profiler/badge.svg?branch=master)](https://coveralls.io/github/giagiannis/data-profiler?branch=master)
=============
data-profiler is a project used to perform data profiling in an operator agnostic way.

Installation
------------
You must install golang prior to the following installation:
```bash
# GOPATH must be created and set
go get github.com/giagiannis/data-profiler
cd $GOPATH/src/
go install ./...
```

Usage
-----
Use the data-profiler-utils binary to use this project

Example 1: estimating similarities using the Bhattacharyya coefficient
```bash
$GOPATH/bin/data-profiler-utils similarities -i <INPUT DIR> -o <OUTPUT FILE> -l <LOG FILE> -opt tree.scale=0.5
```

Example 2: estimating an approximate similarity matrix with 100 datasets using the Bhattacharyya coefficient
```bash
$GOPATH/bin/data-profiler-utils similarities -i <INPUT DIR> -o <OUTPUT FILE> -l <LOG FILE> -opt tree.scale=0.5 -p APRX,count=100
```

Example 3: transforming a similarity matrix to a Dataset Space using 3 dimensions
```bash
$GOPATH/bin/data-profiler-utils mds -k 3 -o <OUTPUT> -sc $GOPATH/src/github.com/giagiannis/data-profiler/_rscripts/mdscaling.R -sim <SIMILARITY MATRIX>
```


License
-------
Apache License v2.0 (see [LICENSE](LICENSE) file for more)


Contact
-------
Giannis Giannakopoulos ggian@cslab.ece.ntua.gr
