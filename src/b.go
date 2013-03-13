package main

import(
	"fmt"
	"net/url"
	"path"
	"strings"
)

func main(){
	s := "http://scy.icerote.net/app/"
	s2 := "http://scy.icerote.net/../app"

	u2,_ := url.Parse(s)
	fmt.Println(u2.Scheme, u2.Opaque, u2.Host, u2.Path, u2.RawQuery, u2.Fragment)
	fmt.Println(path.Clean(u2.Path))

	u2,_ = url.Parse(s2)
	fmt.Println(u2.Scheme, u2.Opaque, u2.Host, u2.Path, u2.RawQuery, u2.Fragment)
	fmt.Println(path.Clean(u2.Path))

	r := strings.Split("/doc/aa/bb/cc", "/")
	fmt.Println(len(r))

	m := make(map[string] int)
	m["aa"] = 3;

	for e := range m{
		fmt.Printf("%T, %+v\n", e, e)
	}

}

