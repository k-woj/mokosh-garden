//go:build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	if err := build(); err != nil {
		log.Fatal(err)
	}
	log.Println("serving http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("web"))))
}

func build() error {
	cmd := exec.Command("go", "build", "-o", "web/game.wasm", ".")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return err
	}
	src := string(goroot[:len(goroot)-1]) + "/lib/wasm/wasm_exec.js"
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile("web/wasm_exec.js", data, 0644)
}
