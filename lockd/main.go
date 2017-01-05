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
			key := r.FormValue("key")
			if key == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("empty lock key\n"))
				return
			}

			timeout, _ := strconv.Atoi(r.FormValue("timeout"))

			if timeout <= 0 {
				timeout = 60
			}

			res, err := a.Lock(time.Duration(timeout)*time.Second, key)
			if err != nil && err != errLockTimeout {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else if err == errLockTimeout {
				w.WriteHeader(http.StatusRequestTimeout)
				w.Write([]byte("Lock timeout\n"))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(res))
			}

		case "DELETE":
			key := r.FormValue("key")

			if key == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("The key is empty,pls enter the key\n"))

			} else {

				res, err := a.UnlockKey(key)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(err.Error()))
				} else {
					w.WriteHeader(http.StatusOK)
					//	rs := "Unlock key: " + key + "done"
					w.Write([]byte(res))
				}
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	})
	log.Fatal(http.ListenAndServe(":14000", nil))

}
