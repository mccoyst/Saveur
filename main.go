// © Steve McCoy 2013. Licensed under the MIT license.

/*
Saveur periodically sends "put" events to all of your open acme windows.
*/
package main

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/goplan9/plan9"
	"code.google.com/p/goplan9/plan9/acme"
	"code.google.com/p/goplan9/plan9/client"
)

var interval = flag.Duration("i", 1*time.Minute, "duration between saves")

func main() {
	flag.Parse()

	putall()

	t := time.Tick(*interval)
	for _ = range t {
		putall()
	}
}

func putall() {
	fs, err := client.MountService("acme")
	if err != nil {
		panic(err)
	}

	index, err := fs.Open("index", plan9.OREAD)
	if err != nil {
		panic(err)
	}
	defer index.Close()

	ids := []int{}
	sc := bufio.NewScanner(index)
	for sc.Scan() {
		id, name, err := fields(sc.Text())
		if err != nil || name == "" {
			continue
		}
		if fi, err := os.Stat(name); err != nil || !fi.Mode().IsRegular() {
			continue
		}

		ids = append(ids, id)
	}
	if err = sc.Err(); err != nil {
		return
	}

	for _, id := range ids {
		w, err := acme.Open(id, nil)
		if err != nil {
			continue
		}
		w.Ctl("put")
		w.CloseFiles()
	}
}

func fields(line string) (int, string, error) {
	idstr := strings.TrimSpace(line[0:11])
	// tagcount := line[12:23] yes, skipping one for the space that is mandated
	// bodycount := line[24:35]
	// isdir := line[36:47]
	// dirty := line[48:59]
	tag := line[60:]

	id64, err := strconv.ParseInt(idstr, 10, 32)
	if err != nil {
		return 0, "", err
	}
	id := int(id64)

	if tag == "" {
		return id, "", nil
	}

	i := strings.Index(tag, " ")
	if i == -1 {
		return id, tag, nil
	}

	return id, tag[:i], nil
}
