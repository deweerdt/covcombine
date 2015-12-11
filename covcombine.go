package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

type counts struct {
	nr_stmts int
	count    int
}

func main() {
	out := ""
	flag.StringVar(&out, "out", "", "out file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s -out <outfile> <profile files...>\n", os.Args[0])
	}
	flag.Parse()
	if out == "" || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	mode := ""
	var modelock sync.Mutex
	res := make(map[string]*counts)
	var reslock sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(flag.Args()))
	for _, n := range flag.Args() {
		go func() {
			f, err := os.Open(n)
			if err != nil {
				fmt.Printf("Error opening '%s': %s\n", n, err.Error())
				os.Exit(1)
			}
			defer f.Close()
			reader := bufio.NewReader(f)
			lineno := 1
			for {
				l, err := reader.ReadString('\n')
				if lineno == 1 {
					fields := strings.Split(l, " ")
					if len(fields) != 2 || fields[0] != "mode:" {
						fmt.Printf("Malformed first line in '%s', expecting mode\n", n)
						os.Exit(1)
					}
					modelock.Lock()
					if mode != "" && fields[1] != mode {
						fmt.Printf("Can't combine mismatched modes: current mode is '%s' and '%s' has mode '%s'\n", mode, n, fields[1])
						os.Exit(1)
					}
					mode = fields[1]
					modelock.Unlock()
					lineno++
					continue
				}
				lineno++
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Printf("Error reading '%s': %s\n", n, err.Error())
					os.Exit(1)
				}
				l = strings.TrimSpace(l)
				fields := strings.Split(l, " ")
				if len(fields) != 3 {
					fmt.Printf("Malformed line %d in '%s'\n", lineno, n)
					os.Exit(1)
				}
				nr_stmts, err := strconv.Atoi(fields[1])
				if err != nil {
					fmt.Printf("Malformed number of statememts (%s) line %d in '%s': %s\n", fields[1], lineno, n, err.Error())
					os.Exit(1)
				}
				cnt, err := strconv.Atoi(fields[2])
				if err != nil {
					fmt.Printf("Malformed counts (%s) line %d in '%s': %s\n", fields[1], lineno, n, err.Error())
					os.Exit(1)
				}
				reslock.Lock()
				c, ok := res[fields[0]]
				if !ok {
					res[fields[0]] = &counts{nr_stmts, cnt}
				} else {
					c.count += cnt
					c.nr_stmts += nr_stmts
				}
				reslock.Unlock()

			}
			wg.Done()
		}()
	}
	wg.Wait()
	f, err := os.Create(out)
	if err != nil {
		fmt.Printf("Can't open '%s' for writing: %s\n", out, err.Error())
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	w.WriteString(fmt.Sprintf("mode: %s", mode))
	for k, v := range res {
		w.WriteString(fmt.Sprintf("%s %d %d\n", k, v.nr_stmts, v.count))
	}
	w.Flush()

}
