// Copyright (c) 2023 William Dode. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"os"

	"github.com/flibustenet/dpae/dpae"
)

func main() {
	verb := flag.Bool("v", false, "show the xml")
	flag.Parse()

	if flag.Arg(0) == "" {
		fmt.Fprintf(os.Stderr, "dpae dpae.json\n")
		flag.PrintDefaults()
		os.Exit(-1)
	}
	d, err := start(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur : %v\n", err)
		os.Exit(-1)
	}
	fmt.Println(d.IdFlux, d.Certificat, d.CertifError)
	if *verb {
		fmt.Println(d.Sended)
	}
}

func start(fname string) (*dpae.Dpae, error) {

	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("erreur ouverture %s %v", fname, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture %s %v", fname, err)
	}

	d := &dpae.Dpae{}
	err = json.Unmarshal(data, d)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshal %s %v", fname, err)
	}

	err = d.Auth(d.Identifiants.MotDePasse)
	if err != nil {
		return nil, err
	}

	err = d.Send()
	if err != nil {
		return nil, err
	}

	err = d.Retour()
	if err != nil {
		return d, err
	}

	return d, nil
}
