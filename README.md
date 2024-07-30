## Read and download notices from Makaut's website on your terminal

- Written in Go
- With [Charm CLI](https://github.com/charmbracelet)
- And `poppler` for pdf conversion

![preview](./preview.gif)

> This program **requires `poppler` to be installed**.
> Poppler is available on both Windows and Linux.

### Features
- Very fast
- Fully TUI
- No need of a web browser
- Read all the notices on your terminal
- Download PDFs or save them as PNG

### How it works
1. It fetches `makaut1.ucanapply.com/smartexam/public/api/notice-data` which is the API URL for list of notices.
2. It shows the notices in a table format with keyboard controls.
3. On Enter, it fetches the PDF link of the notice and save it in `temp` directory locally.
4. Upon save, it usess `poppler` to copy the PDF into a text file.
5. Then it reads that text file and puts its content to the terminal.

Aditionally,
- On click `s` for save, it moves the PDF file from temp to user's download directory.
- Or on click `p` for save as PNG, it creates a PNG file of the PDF by `poppler` and save it into user's download directory.

### Set up project
- Install [Go](https://go.dev/doc/install)
- Clone this project and from this directory run, `go get .` to install dependencies.
- Run the project by `go run .` or by `just run`.

## Build it
It only has one `main.go` file.

You can use `just` to build the binary file

```bash
just build
```

or by typing `go build .`

*When using `just` the binary will be in `bin/` folder, full path will be `bin/read-notice`.*

## That's it
