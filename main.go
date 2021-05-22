/*
hartool is a utility to parse a .HAR (HTTP Archive) file and re-create the files and directory structure.

It's hideously untested and definitely not feature complete, and almost certainly buggy.

Copyright 2021 Mike Hughes mike@mikehughes.info
*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println(fmt.Errorf("invalid number of arguments"))
		os.Exit(1)
	}
	fname := os.Args[1]
	if !strings.HasSuffix(fname, ".har") {
		fmt.Println(fmt.Errorf("must specify an .HAR file"))
		os.Exit(1)
	}
	f, err := os.Open(fname)
	if err != nil {
		fmt.Println(fmt.Errorf("could not open file %s: %v", fname, err))
		os.Exit(1)
	}
	harFile := new(Harfile)
	harContent, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(fmt.Errorf("could not read content from file %s: %v", fname, err))
		os.Exit(1)
	}
	if err = json.Unmarshal(harContent, harFile); err != nil {
		fmt.Println(fmt.Errorf("could not decode JSON content from file %s: %v", fname, err))
		os.Exit(1)
	}
	for _, e := range harFile.Log.Entries {
		u, err := url.Parse(e.Request.URL)
		if err != nil {
			fmt.Println(fmt.Errorf("could not decode URL %s: %v", u, err))
			continue
		}
		p := filepath.FromSlash(u.Hostname() + u.EscapedPath())
		d := filepath.Dir(p)
		if err = os.MkdirAll(d, os.ModePerm); err != nil {
			fmt.Println(fmt.Errorf("could not create directory %s: %v", d, err))
			continue
		}
		fout, err := os.OpenFile(p, os.O_CREATE, os.ModePerm)
		if err != nil {
			fmt.Println(fmt.Errorf("could not create file %s: %v", p, err))
			continue
		}
		var data []byte
		switch e.Response.Content.Encoding {
		case "base64":
			data, err = base64.StdEncoding.DecodeString(e.Response.Content.Text)
			if err != nil {
				fmt.Println(fmt.Errorf("could not decode base64 %s: %v", p, err))
				continue
			}
		default:
			data = []byte(e.Response.Content.Text)
		}
		n, err := fout.Write(data)
		if err != nil {
			fmt.Println(fmt.Errorf("could not write file %s: %v", p, err))
			continue
		}
		if n != e.Response.Content.Size {
			fmt.Println(fmt.Errorf("incomplete write %s: %d of %d bytes", p, n, e.Response.Content.Size))
			continue
		}
	}
}
