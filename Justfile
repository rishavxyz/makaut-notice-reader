run: build
    ./bin/read-notice

build:
    go build -o bin/read-notice main.go

clean:
    rm -rv bin/
