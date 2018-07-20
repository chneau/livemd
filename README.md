# livemd

## Usage
```bash
go get -u -v github.com/chneau/livemd
livemd
```

And visit [http://localhost:8888](http://localhost:8888/) in your browser. When you save Markdown files in that directory, if you're looking at the file, it will automatically reupdate the content.  
It will always show you the content of the latest modified Markdown file.

```bash
Usage of livemd:
  -path string
        dir to watch (and all subdirs ...) (default ".")
  -port string
        port to listen on (default "8888")
``` 

