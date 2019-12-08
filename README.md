## usage
#### post

```sh
# pass a file to the file form field, returns a url with a random id
$ curl -F file=@/path/to/file http://localhost:8080
http://localhost:8080/?id=abcdef

# pass an id field if you want to choose it
$ curl -F file=@/path/to/file -F id=bin http://localhost:8080
http://localhost:8080/?id=bin
```
#### get

```sh
# get the bin's content with curl -G and the id
$ curl -Gd id=bin http://localhost:8080
some text here
```
