# gogo

gogo to generate a golang project based on go-starter





```bash
$ ./gogo -h
Usage of ./gogo:
  -dir string
    	target directory (default ".")
  -disableCache
    	disable cache of go-starter project downloading
  -pkg string
    	package name, default to last element of target directory
    	
$ ./gogo -dir ../gogotest -disableCache
gogotest created successfully in ../gogotest!

```


function:

1. replacing `go-starter` to pkg
1. replacing `GO_STARTER` to SNAKE_CASE of pkg
