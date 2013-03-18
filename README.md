Docbin - export a zip file as virtual web directory
======

Docbin is a Fast CGI app. It intends to response content from zip package directly.

Motivation
=========================
I have many large number of documents, such as boost, html5. They contain lots of small html and image files. It's hard to matain and backup them. Since they are modified rarely, so I think put them into package files are better. I also need to expose them via web, it needs the documents are compressed in gzip or deflate format during transmission. I think gzip is better, because there are many existed gzip tools. But gzip can't contains multiple compressed files, so, zip is my choice.

Zip format contains multiple files, and each of them is compressed with defalte algorithm or store. If the file is deflated, Docbin send the deflated files block directly, without inflate, and set Content-Encoding header with "deflate". This behavior can save the bandwidth and CPU greatly. Thus I can run Docbin on my Raspberry Pi, it's an ARM computer.

Since the Pi has limited CPU resource and all popular web browsers are supput deflate encoding,so Docbin doesn't respect Accept-Encoding request header. It'll reponse "deflate" for always if the reqested file is deflated ( if the file is just "store", the response does not set the Content-Encoding header)

Known Issues
==========================
* Wget doesn't deflate the received html content, and I don't know any tool at hand to deflate the received file.
* I'm not familiar with Golang, Docbin seems use too many memory.

Config
==========================
The config of Docbin is a json file, and specified with "-config" flag. 
in the top level of config:
- ` dash		path to a dashboard static html file. It'll be used when access http://site/$root/
- ` root		root path of documents in some site. if root is "/doc", it means all documents are under http://site/doc/
- ` docs		an object, contains virtual directory to zip package map. Property name is virtual dir name. Property value is an array, it contains at most 3 elements. The first elements is the path to zip file. Second is the path of index page in the zip. the last is the prefix of zipped document. For example

	"boost1.40_cn":["/mnt/d160/scy/www/docs/boost_1_40_cn.zip", "index.html", "boost1.40/"]

for the request url http://scy.icerote.net/doc/boost1.40_cn/index.html, Docbin will open zip file, and get file "boost1.40/boost1.40_cn/index.html" from it.


