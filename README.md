## Read and download announcements from Makaut on your terminal

![preview](./preview.gif)


- Written in Go
- With [Charm CLI](https://github.com/charmbracelet)
- And `poppler` for pdf conversion

> That means this program **requires poppler to be installed**.
> Poppler is available on both Windows and Linux.


**Features**
- Very fast
- Fully TUI
- No need a web browser
- Read all the notices on your terminal
- Download PDFs or save them as PNG

**Set up project**
- Install [Go](https://go.dev/doc/install)
- Clone this project and from this directory run, `go get .` to install dependencies.
- Run the project by `go run .` or by `just run`.

**Build**
You can use `just` to build the binary file

```bash
just build
```

or by typing `go build .`

*When using `just` the binary will be in `bin/` folder, full path will be `bin/read-notice`.*

## That's it
