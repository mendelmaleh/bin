package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Bin struct {
		Addr string // address for web the webserver

		Dir  string // directory to store the pastebins in
		Days int    // expiration for files
	}

	Words WordsConfig
}

//go:embed index.html
var index []byte

func main() {
	var config Config
	var words *Words
	var counter int

	if err := func() error {
		// config
		data, err := os.ReadFile("config.toml")
		if err != nil {
			return err
		}

		if err = toml.Unmarshal(data, &config); err != nil {
			return err
		}

		// word generator
		words, err = NewWords(config.Words)
		return err
	}(); err != nil {
		log.Fatal(err)
	}

	// clean up
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			clean(config.Bin.Dir, config.Bin.Days)
			<-ticker.C
		}
	}()

	// web server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := func() string {
			if r.Method == http.MethodGet {
				if r.URL.Path == "/" {
					w.Write(index)
				} else {
					// serve pastebin from bins dir
					http.ServeFile(w, r, config.Bin.Dir+r.URL.Path)
				}
			}

			if r.Method == http.MethodPost {
				counter += 1

				var code string
				for {
					code = words.Code()
					if _, err := os.Stat(config.Bin.Dir + code); errors.Is(err, os.ErrNotExist) {
						break
					}

					fmt.Printf("reattempting... (counter=%d)\n", counter)
					time.Sleep(time.Second) // rate limit as it gets crowded
				}

				f, err := os.Create(config.Bin.Dir + code)
				if err != nil {
					return err.Error()
				}

				if _, err := io.Copy(f, r.Body); err != nil {
					return err.Error()
				}

				return code
			}

			return ""
		}()

		fmt.Fprintln(w, resp)
	})

	if err := http.ListenAndServe(config.Bin.Addr, nil); err != nil {
		log.Fatal(err)
	}
}

func clean(dir string, days int) {
	if err := func() error {
		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		cutoff := time.Now().AddDate(0, 0, -days)
		for _, f := range files {
			info, err := f.Info()
			if err != nil {
				return err
			}

			if info.ModTime().Before(cutoff) {
				if err := os.Remove(dir + f.Name()); err != nil {
					return err
				}
			}
		}

		return nil
	}(); err != nil {
		log.Println(err)
	}
}
