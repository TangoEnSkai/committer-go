# Benchmark Baselines

Run with: `make bench`

These baselines were captured on 2026-06-20 with Go 1.18.2 on Darwin arm64.
Update this file when significant performance changes are made.

```
goos: darwin
goarch: arm64
pkg: github.com/TangoEnSkai/committer-go/committer
BenchmarkValidateMessage-10                         	    5713	     17748 ns/op	   31548 B/op	     360 allocs/op
BenchmarkValidateMessage_Invalid-10                 	  122894	       953.9 ns/op	     472 B/op	      32 allocs/op
BenchmarkSuggest_ValidMessage-10                    	    7341	     17268 ns/op	   31548 B/op	     360 allocs/op
BenchmarkSuggest_InvalidType-10                     	    6499	     19637 ns/op	   32701 B/op	     432 allocs/op
BenchmarkCheckCommitTypeWithConfig_ExtraTypes-10    	 4120612	        28.94 ns/op	       8 B/op	       1 allocs/op
BenchmarkValidateFullMessage_WithBody-10            	    7046	     17527 ns/op	   31612 B/op	     360 allocs/op
BenchmarkFindTypeSuggestion-10                      	  153111	       783.6 ns/op	     176 B/op	      30 allocs/op
BenchmarkPatternMatch-10                            	    7078	     17411 ns/op	   31516 B/op	     358 allocs/op
BenchmarkValidateMessage_LongMessage-10             	  591363	       208.6 ns/op	     288 B/op	       3 allocs/op
PASS
ok  	github.com/TangoEnSkai/committer-go/committer	1.671s
```
