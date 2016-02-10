package main

import (
	"log"
	"net"
	"net/http"

	wt "github.com/wellington/wellington"
)

func main() {

	gba := &wt.BuildArgs{
		ImageDir:  "",
		BuildDir:  "",
		Includes:  []string{""},
		Font:      "",
		Gen:       "",
		Style:     0,
		Comments:  false,
		CacheBust: "",
	}
	httpPath := "wellington-io.appspot.com"
	http.Handle("/build/", wt.FileHandler(gba.Gen))
	log.Println("Web server started on :12345")

	lis, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatal("Error listening on :12345", err)
	}

	http.HandleFunc("/", wt.HTTPHandler(gba, httpPath))
	http.Serve(lis, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Println("Server closed")

}
