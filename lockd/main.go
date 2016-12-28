package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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
			names := strings.Split(r.FormValue("names"), ",")
			if len(names) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("empty lock names"))
				return
			}

			timeout, _ := strconv.Atoi(r.FormValue("timeout"))

			if timeout <= 0 {
				timeout = 60
			}

			idinfo := make(chan uint64, 1)
			errs := make(chan error, 1)
			chmsg := make(chan string, 1)
			go func() {
				a.LockTimeout(idinfo, errs, time.Duration(timeout)*time.Second, names)
			}()

			select {
			//for broadcast
			case <-chmsg:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("aaaaa"))
				return
			case id := <-idinfo:

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(strconv.FormatUint(id, 10)))
			case err := <-errs:

				if err != nil && err != errLockTimeout {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(err.Error()))
				} else if err == errLockTimeout {
					w.WriteHeader(http.StatusRequestTimeout)
					w.Write([]byte("Lock timeout"))
				}
			}

		case "DELETE":
			id, err := strconv.ParseUint(r.FormValue("id"), 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			err = a.Unlock(id)

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
