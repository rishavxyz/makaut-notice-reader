# makaut-notice-reader

Read and download Makaut's official notices from your terminal, no web browser needed.


### Why did I make this?

I'm learning Go recently and found its very tedious to visit their website to
read the notices often. So why not creating a simple script that does it for me
without leaving my terminal.

### Development

Clone this repo and do whatever changes you like.

> Note: On [https://makautexam.net/](Makaut's website), they have exposed their
API url on the "console tab" and it is a public readonly API so I used it.

**Folder stcucture**

```
makaut-notice-reader
|- ...
|- main.go -- main file
|- utils/ -- helper functions
  |- ...
```

**Run the file**

```bash
go run main.go
```
