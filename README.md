# livemd

## Install
```bash
go get -u -v github.com/chneau/livemd
```

## Usage
```bash
livemd
C:\Users\c\go\src\github.com\chneau\livemd\README.md
Listening on http://localhost:8888/ 
```

```go
package main;
func main() {
      print("banane")
      b := 6/7
}
```

And visit [http://localhost:8888](http://localhost:8888/) in your browser.  
The command will watch all Markdown file in the directory and sub directories.  
When a file is modified, its content will be sent to the web page.  
You will always see the latest modified Markdown file.
```bash
Usage of livemd: 
  -path string
        dir to watch (and all subdirs ...) (default ".")
  -port string
        port to listen on (default "8888")
``` 

