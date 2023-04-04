package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func clean(dir string) {
	if err := func() error {
		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		cutoff := time.Now().AddDate(0, 0, -7)
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

func main() {
	const dir = "bins/"
	var counter int

	// clean up
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			clean(dir)
			<-ticker.C
		}
	}()

	// word generator
	words, err := NewWords("/usr/share/dict/usa")
	if err != nil {
		log.Fatal(err)
	}

	// web server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := func() string {
			if r.Method == http.MethodGet {
				if r.URL.Path == "/" {
					return "hi " + words.Code()
				}
				http.ServeFile(w, r, dir+r.URL.Path)
			}

			if r.Method == http.MethodPost {
				counter += 1

				var code string
				for {
					code = words.Code()
					if _, err := os.Stat(dir + code); errors.Is(err, os.ErrNotExist) {
						break
					}

					fmt.Printf("reattempting... (counter=%d)\n", counter)
					time.Sleep(time.Second) // rate limit as it gets crowded
				}

				f, err := os.Create(dir + code)
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

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
