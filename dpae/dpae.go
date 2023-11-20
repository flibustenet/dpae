// Copyright (c) 2023 William Dode. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dpae

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

var TimeOut = time.Second * 60

var RetryTempoFirst = time.Second
var RetryTempo = 10 * time.Second
var RetryNb = 60

var UrlDepot = "https://depot.dpae-edi.urssaf.fr/deposer-dsn/1.0/"
var UrlConsultation = "https://consultation.dpae-edi.urssaf.fr/lister-retours-flux/2.0/"

type Employer struct {
	Designation   string
	SIRET         string
	APE           string
	URSSAFCode    string
	Adress        string
	Town          string
	Postal        string
	Phone         string
	HealthService string
}

type Employee struct {
	Surname         string
	ChristianName   string
	Sex             int
	NIR             string
	NIRKey          string
	BirthDate       Date
	BirthTown       string
	BirthDepartment string
}

type Contract struct {
	StartContractDate Date
	StartContractTime Time
	EndContractDate   Date
	NatureCode        string
}

type Dpae struct {
	TestIndicator int // test:1 prod:120
	Identifiants  Identifiants
	Employer      Employer
	Employee      Employee
	Contract      Contract
	// answer
	Jeton  string
	IdFlux string
	Sended string // xml sended
	// retour
	Certificat  string
	CertifError string // message si certif error
}

func (d *Dpae) String() string {
	return fmt.Sprintf("dpae employer=%s flux=%s", d.Employer.Designation, d.IdFlux)
}

type BilanError struct {
	Bilan string
}

func (a BilanError) Error() string {
	return a.Bilan
}

// SendDpae send the DPAE and receive the IdFlux
// doIt
// false : test = 1
// true : prod = 120
// record Sended (the template sended)
// record IdFlux (to get the retour)
func (d *Dpae) Send() error {
	if d.Jeton == "" {
		return UErr(nil, "Jeton vide")
	}
	d.Employer.HealthService = "01"
	if d.Employee.BirthDepartment == "00" { // num provisoire
		d.Employee.BirthDepartment = "99"
	}
	if len(d.Employee.BirthDepartment) > 2 { // si dom-tom (9xx) ne prendre que 9x
		d.Employee.BirthDepartment = d.Employee.BirthDepartment[0:2]
	}
	bUTF := &bytes.Buffer{}
	err := tDpae.Execute(bUTF, d)
	if err != nil {
		return fmt.Errorf("Error in template dpae : %v\n%v", err, d)
	}
	// remember the template sended
	d.Sended = bUTF.String()

	// iso8859 encoding in bISO
	bISO := &bytes.Buffer{}
	toIso := charmap.ISO8859_1.NewEncoder().Writer(bISO)
	io.Copy(toIso, bUTF)

	// zip in bufgz
	bufgz := &bytes.Buffer{}
	g := gzip.NewWriter(bufgz)
	io.Copy(g, bISO)
	g.Close()

	// send ziped bufgz
	req, err := http.NewRequest("POST", UrlDepot, bufgz)
	if err != nil {
		return fmt.Errorf("build request POST %s: %v", UrlDepot, err)
	}
	req.Header.Add("Content-Type", "application/xml")
	req.Header.Add("Authorization", "DSNLogin jeton="+d.Jeton)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{Timeout: TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return UErr(nil, "Network error at sending "+err.Error())
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return UErr(nil, "Network error at reading "+err.Error())
	}
	// recup idflux
	reIdFlux := regexp.MustCompile(`idflux>(.*)</idflux`)
	fd := reIdFlux.FindStringSubmatch(string(data))
	if len(fd) != 2 {
		return fmt.Errorf("DPAE: idflux not found in %s", data)
	}
	d.IdFlux = fd[1]
	if len(d.IdFlux) != 23 {
		return fmt.Errorf("idflux length should be 23 : %s", d.IdFlux)
	}
	return nil
}

type Consultation struct {
	Retours Retours `json:"retours"`
}
type Retour struct {
	Publication string `json:"publication"`
	Production  string `json:"production"`
	Nature      string `json:"nature"`
	Statut      string `json:"statut"`
	ID          string `json:"id"`
	URL         string `json:"url"`
}

type Flux struct {
	ID     string   `json:"id"`
	Retour []Retour `json:"retour"`
}

type Retours struct {
	Flux []Flux `json:"flux"`
}

// URSSAFError
// Erreur provenant de l'URSSAF de type accès réseau impossible
// pour différencier des erreurs de debug
type URSSAFError struct {
	Message string
	Err     error
}

func (a URSSAFError) Error() string {
	return a.Message
}
func UErr(err error, s string) *URSSAFError {
	return &URSSAFError{
		Message: fmt.Sprintf("URSSAF: %s", s),
		Err:     err,
	}
}

// Retour get answers from URSSAF starting at 0 to RETRY_NB
func (d *Dpae) Retour() error {
	return d.retour(0)
}

// retour get answers from URSSAF, trying to RETRY_NB
func (d *Dpae) retour(retry int) error {
	if retry == 0 {
		time.Sleep(RetryTempoFirst)
	} else {
		time.Sleep(RetryTempo)
	}
	if retry > RetryNb {
		return UErr(nil, fmt.Sprintf("No answer with idflux %s after %d tries", d.IdFlux, RetryNb))
	}
	retry = retry + 1

	if d.IdFlux == "" {
		return fmt.Errorf("no IdFlux")
	}
	req, err := http.NewRequest("GET", UrlConsultation+d.IdFlux, nil)
	if err != nil {
		return fmt.Errorf("build request GET %s: %v", UrlConsultation+d.IdFlux, err)
	}
	req.Header.Add("Authorization", "DSNLogin jeton="+d.Jeton)

	client := &http.Client{Timeout: TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		// network error, try again
		return d.retour(retry)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		// network error try again
		return d.retour(retry)
	}
	consultation := &Consultation{}
	err = json.Unmarshal(data, &consultation)
	if err != nil {
		return fmt.Errorf("unmarshal retours on %v : %s : %v", d, data, err)
	}
	urls := []string{}
	for _, flux := range consultation.Retours.Flux {
		for _, retour := range flux.Retour {
			urls = append(urls, retour.URL)
		}
	}
	if len(urls) == 0 {
		return d.retour(retry)
	}

	reCertificatConformite := regexp.MustCompile(`<certificat_conformite>(.*)</certificat_conformite>`)
	reCertificatNonConformite := regexp.MustCompile(`(?s)<message>(.*)</message>`)

	for _, url := range urls {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("build request GET %s: %v", url, err)
		}
		req.Header.Add("Authorization", "DSNLogin jeton="+d.Jeton)
		client := &http.Client{Timeout: TimeOut}
		resp, err = client.Do(req)
		if err != nil {
			// network error, try again
			return d.retour(retry)
		}

		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			// network error, try again
			return d.retour(retry)
		}
		str := string(data)

		if !strings.Contains(str, `profil="DPAE"`) { // not the good profil, ignore it
			continue
		}

		if strings.Contains(str, `<etat_conformite>KO</etat_conformite>`) {
			bilan := reCertificatNonConformite.FindStringSubmatch(str)
			if len(bilan) != 2 {
				return fmt.Errorf("DPAE pas de description pour non conformité %v\n%s", d, str)
			}
			d.CertifError = bilan[1]
			return UErr(nil, "Non conforme : "+d.CertifError)
		}
		if !strings.Contains(str, `<etat_conformite>OK</etat_conformite>`) {
			return fmt.Errorf("devrait contenir conformite %v\n%s", d, str)
		}
		certif := reCertificatConformite.FindStringSubmatch(str)
		if len(certif) != 2 {
			return fmt.Errorf("ne trouve pas de certificat %v : %s\n%s", d, certif, str)
		}
		d.Certificat = certif[1]
		if len(d.Certificat) < 10 {
			return fmt.Errorf("certificat incorrect %v \n%s", d, str)
		}
		break
	}
	if len(d.Certificat) == 0 {
		return d.retour(retry)
	}
	if len(d.Certificat) < 10 {
		return fmt.Errorf("pas de certificat sur %v", d)
	}
	return nil
}
