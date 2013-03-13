package main

import(
	"fmt"
	"net/url"
)

func main(){
	s := "http://scy.icerote.net/app/"
	s2 := "http://scy.icerote.net/app"
	us,_ := url.Parse(s)
	fmt.Printf("%+v\n",us)

	u2,_ := url.Parse(s2)
	fmt.Println(u2.Scheme, u2.Opaque, u2.Host, u2.Path, u2.RawQuery, u2.Fragment)

}

