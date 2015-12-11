Combine several test profiles as generaged by 'go test -coverprofile'.

# Installation
```
$ go get github.com/deweerdt/covcombine
```

# Usage

```
$ go test -covprofile a package/a
$ go test -covprofile b package/b
$ covcombine -out combined a b
$ go tool cover -html combined
```
