package main

import (
	"flag"
	"fmt"
	"os"
	"io"
	"encoding/json"
	"net/http"
	"net/http/fcgi"
	"path"
	"strings"
	"sort"
	"archive/zip"
	"errors"
)

type Config struct{
	Root string
	Docs map[string]string
}
type VPair struct {
	dir string
	zfile string
	rc *zip.ReadCloser
	files map[string]*zip.File
}

type DocArray struct {
	data []VPair
}

func (p *DocArray) Len() int            { return len(p.data) }
func (p *DocArray) Less(i, j int) bool  { return p.data[i].dir < p.data[j].dir }
func (p *DocArray) Swap(i, j int)       { p.data[i], p.data[j] = p.data[j], p.data[i] }

type FastCGIServer struct{
	docs DocArray
}

func NewFCGIServer(cfg Config) *FastCGIServer{
	b := new(FastCGIServer)
	root := path.Clean(cfg.Root) + "/"
	b.docs.data = make([]VPair, len(cfg.Docs))
	i := 0
	for k,v := range cfg.Docs {
		b.docs.data[i] = VPair{path.Clean(root + k) + "/",v, nil, nil}	// key is "root/virtual_dir/"
		i++
	}
	sort.Sort(&b.docs)

	return b
}

func (s FastCGIServer) E4xx(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	fmt.Fprint(w, "Error:", code)
}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fpath := path.Clean(req.URL.Path)		// "Root/virtual_dir/doc_path"

	// equal to C++ upper_bound
	//fmt.Println("Debug: fpath", fpath)
	idx := sort.Search(len(s.docs.data), func(i int) bool { return s.docs.data[i].dir >= fpath})
	//fmt.Println("Debug: idx ", idx)
	if idx == 0{
		s.E4xx(w, 404)
		return;
	}
	idx--

	item := s.docs.data[idx]

	if !strings.HasPrefix(fpath, item.dir) {
		s.E4xx(w, 404)
		return;
	}
	fileInZip := fpath[len(item.dir):]

	rc, err := s.getFile(&item, fileInZip)
	if err != nil {
		s.E4xx(w, 500)
		return
	}
	defer rc.Close()
	io.Copy(w, rc)

	return
}

func (s FastCGIServer) getFile(item *VPair, file string)(r io.ReadCloser, err error) {
	if item.rc == nil{
		item.rc, err = zip.OpenReader(item.zfile)
		if err == nil{
			item.files = make(map[string]*zip.File)
			for _,f := range item.rc.File {
				item.files[f.Name] = f
			}
		}
	}
	fh, ok := item.files[file]
	if !ok {
		return nil, errors.New("zip: file not found")
	}
	r, err = fh.Open()
	return
}

func main() int {
	configFile := flag.String("config", "docbin.conf", "docbin config file path")
	flag.Parse()

	f, err := os.Open(*configFile)
	if err != nil {
		fmt.Print("Failed to open config file %s\n", configFile)
		return 1
	}
	defer f.Close()

	jdecoder := json.NewDecoder(f)

	var cfg Config
	err = jdecoder.Decode(&cfg)
	if err != nil {
		fmt.Printf("Failed to parse config file %s, %+v\n", configFile, err)
		return 1
	}

	b := NewFCGIServer(cfg)


	fmt.Printf("Starting server")
	fcgi.Serve(nil, b)

	return 0
}
