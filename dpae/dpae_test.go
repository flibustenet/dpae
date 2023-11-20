// Copyright (c) 2023 William Dode. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dpae

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func readDpaeTest() (*Dpae, error) {
	RetryTempo = time.Second
	fname := os.Getenv("DPAE_TEST_JSON")
	if fname == "" {
		fname = "dpae_test.json"
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("erreur ouverture %s %v", fname, err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture %s %v", fname, err)
	}
	d := &Dpae{}
	err = json.Unmarshal(data, d)
	if err != nil {
		return nil, fmt.Errorf("erreur unmarshal %s %v", fname, err)
	}
	d.TestIndicator = 1 // test (prod=120)
	return d, nil
}

func TestReXML(t *testing.T) {
	RetryTempo = time.Second
	for _, d := range []struct {
		orig string
		res  string
	}{
		{"abcd", "abcd"},
		{"ab_cd", "ab cd"},
		{"ab€cd", "ab cd"},
		{"ab,;_cd", "ab   cd"},
		{"éàê", "éàê"},
		{"abcdefghijklmnopqrstuvwxyzazerty012345", "abcdefghijklmnopqrstuvwxyzazerty"},
	} {
		r := fmtXML(d.orig)
		if r != d.res {
			t.Fatalf("%s attend %s reçoit %s", d.orig, d.res, r)
		}
	}
}
func TestAuthOk(t *testing.T) {
	RetryTempo = time.Second
	if testing.Short() {
		t.Skip("skipping test DPAE in short mode.")
	}

	d, err := readDpaeTest()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// test auth ok
	err = d.Auth(d.Identifiants.MotDePasse)
	if err != nil {
		t.Fatalf("auth dpae error : %v", err)
	}
	if len(d.Jeton) != 408 {
		t.Fatalf("Jeton devrait être de 408c :\n%s", d.Jeton)
	}
}

func TestAuth(t *testing.T) {
	RetryTempo = time.Second
	if testing.Short() {
		t.Skip("skipping test DPAE in short mode.")
	}

	d, err := readDpaeTest()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// test auth mauvais password
	err = d.Auth("xxx")
	var AuthError *URSSAFError
	if !errors.As(err, &AuthError) {
		t.Fatalf("erreur devrait être URSSAFError : %v", err)
	}
	if !strings.Contains(AuthError.Message, "Authentification") {
		t.Fatalf("Status code devrait être Authentification : %s", AuthError.Message)
	}
}

func TestDpaeSend(t *testing.T) {
	RetryTempo = time.Second
	if testing.Short() {
		t.Skip("skipping test DPAE in short mode.")
	}
	d, err := readDpaeTest()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// test auth ok
	err = d.Auth(d.Identifiants.MotDePasse)
	if err != nil {
		t.Fatalf("auth dpae error : %v", err)
	}
	err = d.Send()
	if err != nil {
		t.Fatalf("envoi err : %v", err)
	}
	d.Employee.BirthDepartment = "99"
	err = d.Send()
	if err != nil {
		t.Fatalf("envoi err avec 99 : %v", err)
	}
}

func TestDpaeRetour(t *testing.T) {
	RetryTempo = time.Second
	if testing.Short() {
		t.Skip("skipping test DPAE in short mode.")
	}
	d, err := readDpaeTest()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// test auth ok
	err = d.Auth(d.Identifiants.MotDePasse)
	if err != nil {
		t.Fatalf("auth dpae error : %v", err)
	}
	err = d.Send()
	if err != nil {
		t.Fatalf("envoi err : %v", err)
	}

	err = d.Retour()
	if err != nil {
		t.Fatalf("retour err : %v\nEnvoyé : \n%s", err, d.Sended)
	}
	err = d.Retour()
	if err != nil {
		t.Fatalf("retour err : %v\nEnvoyé : \n%s", err, d.Sended)
	}
	err = d.Retour()
	if err != nil {
		t.Fatalf("retour err : %v\nEnvoyé : \n%s", err, d.Sended)
	}
}

func TestDpaeRetourErr(t *testing.T) {
	RetryTempo = time.Second
	if testing.Short() {
		t.Skip("skipping test DPAE in short mode.")
	}
	d, err := readDpaeTest()
	if err != nil {
		t.Fatalf(err.Error())
	}
	// test auth ok
	err = d.Auth(d.Identifiants.MotDePasse)
	if err != nil {
		t.Fatalf("auth dpae error : %v", err)
	}
	numSecuCle := d.Employee.NIR
	d.Employee.NIRKey = "xx"
	err = d.Send()
	if err != nil {
		t.Fatalf("envoi err : %v", err)
	}

	err = d.Retour()
	if !strings.Contains(err.Error(), "Non conforme") {
		t.Fatalf("Devrait recevoir DpaeNonConforme %s", err.Error())
	}
	if !strings.Contains(d.CertifError, "Numero de securite sociale invalide") {
		t.Fatalf("Devrait contenir Numero de securite sociale invalide %s", d.CertifError)
	}

	d.Employee.NIR = numSecuCle
	d.Employer.URSSAFCode = "12345"
	err = d.Send()
	if err != nil {
		t.Fatalf("envoi err : %v", err)
	}

	err = d.Retour()
	if !strings.Contains(err.Error(), "Non conforme") {
		t.Fatalf("Devrait recevoir DpaeNonConforme %s", d.Sended)
	}
	if !strings.Contains(d.CertifError, "Code URSSAF invalide") {
		t.Fatalf("Devrait contenir Code URSSAF invalide %s", d.CertifError)
	}
}
