// Copyright (c) 2023 William Dode. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dpae

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

var UrlAuth = "https://mon.urssaf.fr/authentifier_dpae"

type Identifiants struct {
	SIRET      string
	Nom        string
	Prenom     string
	MotDePasse string
	Service    string
}

// Auth appel pour authentification
// enregistre le jeton dans d.Jeton
// passage du mot de passe en param pour éviter
// le risque qu'il n'apparaisse dans un log
func (d *Dpae) Auth(pw string) error {
	if pw == "" || d.Identifiants.SIRET == "" {
		return UErr(nil, "Informations non renseignées")
	}
	d.Identifiants.MotDePasse = pw
	ba := &bytes.Buffer{}
	err := tAuth.Execute(ba, d.Identifiants)
	if err != nil {
		return fmt.Errorf("template auth %s: %v", UrlAuth, err)
	}
	d.Identifiants.MotDePasse = "" // pour éviter log du mdp

	req, err := http.NewRequest("POST", UrlAuth, ba)
	if err != nil {
		return fmt.Errorf("NewRequest %s: %v", UrlAuth, err)
	}
	req.Header.Add("content-type", "application/xml")
	client := &http.Client{Timeout: TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return UErr(err, "Erreur réseau")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return UErr(err, "Erreur réseau")
	}
	if resp.Status == "422 Unprocessable Entity" {
		return UErr(err, "Authentification incorecte")
	}
	if resp.StatusCode != 200 {
		return UErr(err, "Erreur réseau status : "+resp.Status)
	}
	d.Jeton = string(data)
	if len(d.Jeton) < 10 {
		return fmt.Errorf("erreur jeton %s: %v", d.Jeton, err)
	}
	return nil
}
