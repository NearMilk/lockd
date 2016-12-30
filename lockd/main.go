package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/teambition/lockd"
)

var errLockTimeout = errors.New("lock timeout")
var httpaddr = flag.String("http_addr", "127.0.0.1:14000", "http listen address")

func main() {

	flag.Parse()
	a := lockd.NewApp()

	http.HandleFunc("/lock", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println(r.Method)
		switch r.Method {
		case "GET":
			buf := a.GetLockInfos()
			w.Header().Set("Content-Type", "text/plain")
			w.Write(buf)
			return
		case "POST", "PUT":
			names := r.FormValue("names")
			if names == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("empty lock names"))
				return
			}

			timeout, _ := strconv.Atoi(r.FormValue("timeout"))

			if timeout <= 0 {
				timeout = 60
			}

			res, err := a.Lock(time.Duration(timeout)*time.Second, names)
			if err != nil && err != errLockTimeout {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else if err == errLockTimeout {
				w.WriteHeader(http.StatusRequestTimeout)
				w.Write([]byte("Lock timeout"))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(res))
			}

		case "DELETE":
			names := r.FormValue("names")

			if names == "" {
				fmt.Println("aas")
			}

			err := a.UnlockKey(names)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else {
				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	})
	log.Fatal(http.ListenAndServe(":14000", nil))

}
