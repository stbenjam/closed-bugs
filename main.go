package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type result struct {
	Faults []string `json:"faults"`
	Bugs   []bug    `json:"bugs"`
}

type bug struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Resolution string `json:"resolution"`
	Summary    string `json:"summary"`
}

type bugRef struct {
	result  *result
	fileRef []string
}

func main() {
	refs := make(map[string]*bugRef)
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <path>\n", os.Args[0])
		os.Exit(1)
	}

	results, err := exec.Command(
		"grep",
		"--exclude-dir",
		".git",
		"-Pnro",
		"https://bugzilla.redhat.com/show_bug.cgi\\?id=\\K([[:digit:]]+)",
		os.Args[1],
	).Output()
	if err != nil {
		panic(err)
	}

	fmt.Println("Please wait, examining code base for closed bugs...")

	for _, result := range strings.Split(string(results), "\n") {
		ref := strings.Split(result, ":")
		if len(ref) != 3 {
			continue
		}
		file := fmt.Sprintf("%s:%s", ref[0], ref[1])
		bug := ref[2]

		if _, ok := refs[bug]; !ok {
			refs[bug] = &bugRef{
				nil,
				[]string{file},
			}
		} else {
			refs[bug].fileRef = append(refs[bug].fileRef, file)
		}
	}

	var client = &http.Client{}
	for k, _ := range refs {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://bugzilla.redhat.com/rest/bug/%s", k), nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Accept", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			continue
		}
		defer resp.Body.Close()
		res := result{}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			panic(err)
		}
		refs[k].result = &res
	}

	any := false

	for _, v := range refs {
		if v.result != nil && isBugClosed(v.result) {
			any = true
			fmt.Printf("*** Bug %d is %s/%s: %s\n  Link:\n    %s\n  References:\n    %s\n\n",
				v.result.Bugs[0].ID,
				v.result.Bugs[0].Status,
				v.result.Bugs[0].Resolution,
				v.result.Bugs[0].Summary,
				fmt.Sprintf("https://bugzilla.redhat.com/show_bug.cgi?id=%d", v.result.Bugs[0].ID),
				strings.Join(v.fileRef, "\n    "))
		}
	}

	if any {
		os.Exit(1)
	} else {
		fmt.Println("Good job! No closed bugs found.")
	}
}

func isBugClosed(r *result) bool {
	if len(r.Bugs) > 0 {
		status := r.Bugs[0].Status
		return status != "NEW" && status != "ASSIGNED" && status != "POST" && status != "MODIFIED"
	}

	return false
}
