package main

import (
	"flag"
	"sync"
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
	Dash string
	Root string
	Docs map[string][]string
}
type FileStoreInfo struct{
	file *zip.File
	compress uint16
}
type VPair struct {
	dir string
	zfile string
	index string
	prefix string
	rc *zip.ReadCloser
	files map[string]FileStoreInfo
}

type DocArray struct {
	data []VPair
}

func (p *DocArray) Len() int            { return len(p.data) }
func (p *DocArray) Less(i, j int) bool  { return p.data[i].dir < p.data[j].dir }
func (p *DocArray) Swap(i, j int)       { p.data[i], p.data[j] = p.data[j], p.data[i] }

type FastCGIServer struct{
	zipMutex sync.Mutex
	entryMutex sync.Mutex
	root string
	dashboard string
	docs DocArray
}

func NewFCGIServer(cfg Config) *FastCGIServer{
	b := new(FastCGIServer)
	b.dashboard = cfg.Dash
	b.root = path.Clean(cfg.Root) + "/"
	b.docs.data = make([]VPair, len(cfg.Docs))
	i := 0
	for k,v := range cfg.Docs {
		var fname, index, prefix string
		if len(v) == 0 {continue;}
		fname = v[0]
		if len(v) > 1 { index = v[1]}
		if len(v) > 2 { prefix = v[2]}

		b.docs.data[i] = VPair{path.Clean(b.root + k) + "/", fname, index, prefix, nil, nil}	// key is "root/virtual_dir/"
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
	fpath := req.URL.Path		// "Root/virtual_dir/doc_path"

	if fpath == s.root{
		dash, err := os.Open(s.dashboard)
		if err != nil {
			s.E4xx(w, 404)
			return
		}
		defer dash.Close()
		io.Copy(w, dash)
		return;
	}

	// equal to C++ upper_bound
	idx := sort.Search(len(s.docs.data), func(i int) bool { return s.docs.data[i].dir > fpath})
	if idx == 0{
		s.E4xx(w, 404)
		return
	}
	idx--

	item := &s.docs.data[idx]

	if !strings.HasPrefix(fpath, item.dir) {
		s.E4xx(w, 404)
		return;
	}
	head := w.Header()
	if fpath == item.dir {
		fpath += item.index
	}
	fileInZip := fpath[len(item.dir):]

	rc, fsi, err := s.getFile(item, fileInZip)
	if err != nil {
		s.E4xx(w, 404)
		return
	}
	defer rc.Close()
	if fsi.compress == zip.Deflate{
		head["Content-Encoding"] = []string{"deflate"}
	}
	io.Copy(w, rc)

	return
}
func (s FastCGIServer) loadZip(item *VPair)(err error) {
	s.zipMutex.Lock()
	defer s.zipMutex.Unlock()

	if item.rc != nil { return nil}

	item.rc, err = zip.OpenReader(item.zfile)
	if err == nil{
		item.files = make(map[string]FileStoreInfo)
		for _,f := range item.rc.File {
			fstore := FileStoreInfo{f, f.Method}
			f.Method = zip.Store
			item.files[f.Name] = fstore
		}
		fmt.Println("zip loaded<hr/> ", item.zfile)
	} else {
		fmt.Println("open zip faild<hr/> ", item.zfile, err)
	}
	return
}
func (s FastCGIServer) openEntry(entry *zip.File)(io.ReadCloser, error) {
	s.entryMutex.Lock()
	defer s.entryMutex.Unlock()
	return entry.Open()
}

func (s FastCGIServer) getFile(item *VPair, file string)(r io.ReadCloser, rf *FileStoreInfo, err error) {
	if item.rc == nil{
		err = s.loadZip(item)
		if err != nil {return}
	}

	zipFile := item.prefix + file
	fh, ok := item.files[zipFile]
	if !ok {
		fmt.Println("file not found<hr/> ", zipFile)
		return nil, nil, errors.New("zip: file not found")
	}

	rf = &fh
	r, err = s.openEntry(fh.file)
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
