package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
	"unicode"

	"git.sr.ht/~mendelmaleh/bin/base26"
	"github.com/pelletier/go-toml"
	"github.com/prologic/bitcask"
)

// This is for the random id generator, it ensures the id is 6 chars long.
// The limitation is that it won't generate more than 297m (diff) ids.
const (
	min  int = 26 * 26 * 26 * 26 * 26 // 11_881_376  // 26 ** 5
	max  int = min * 26               // 308_915_776 // 26 ** 6
	diff int = max - min              // 297_034_400
)

// Config struct
type Config struct {
	Bin struct {
		DB      string
		Addr    string
		Pattern string
	}
}

func main() {
	// get config
	doc, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Panic(err)
	}

	// parse config
	config := Config{}
	err = toml.Unmarshal(doc, &config)
	if err != nil {
		log.Panic(err)
	}

	// setup db
	db, err := bitcask.Open(config.Bin.DB)
	if err != nil {
		log.Fatal(err)
	}

	// setup http server
	http.HandleFunc(config.Bin.Pattern,
		func(w http.ResponseWriter, r *http.Request) {
			Bin(w, r, db)
		},
	)

	// run
	log.Fatal(http.ListenAndServe(config.Bin.Addr, nil))
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// Bin is a pastebin service, it requires a bitcask kv store.
func Bin(w http.ResponseWriter, r *http.Request, b *bitcask.Bitcask) {
	id := r.FormValue("id")

	if r.Method == "GET" {
		// no file was requested
		if id == "" {
			fmt.Fprintln(w, "todo: instructions")
			return
		}

		// a file was requested
		buf, err := b.Get([]byte(id))
		if err != nil {
			if err == bitcask.ErrKeyNotFound {
				fmt.Fprintf(w, "no bin with id %s\n", id)
				return
			}

			log.Print(err)
			fmt.Fprintf(w, "error %T when retrieving bin.\n", err)
			return
		}

		w.Write(buf)
		return
	}

	if r.Method == "POST" {
		// if no id was passed
		if id == "" {
			// generate random int
			rand.Seed(time.Now().UnixNano())
			n := rand.Intn(diff) + min
			id = base26.Itoa(n)

			// ensure id is available
			for b.Has([]byte(id)) {
				id = ""
				n = rand.Intn(diff) + min
				id = base26.Itoa(n)
			}
		} else if b.Has([]byte(id)) {
			fmt.Fprintf(w, "id %s is already taken. try again!\n", id)
			return
		} else if len(id) > 32 || !isASCII(id) {
			fmt.Fprintf(w, "id %s is not valid. ensure it's all ASCII and under 32 chars.\n", id)
			return
		}

		// get file
		file, _, err := r.FormFile("file")
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "error %T when retrieving file, try again.\n", err)
			return
		}
		defer file.Close()

		// copy file
		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "error %T when retrieving file, try again.\n", err)
			return
		}

		// save file
		err = b.Put([]byte(id), buf.Bytes())
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "error %T when saving file, try again.\n", err)
			return
		}

		// return id
		u := url.URL{
			Scheme: "http",
			Host:   r.Host,
			Path:   "/", // todo: use config or dynamic
		}

		q := url.Values{}
		q.Set("id", id)

		u.RawQuery = q.Encode()
		fmt.Fprintf(w, "%s\n", u.String())

		return
	}
}
