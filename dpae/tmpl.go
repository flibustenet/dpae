// Copyright (c) 2023 William Dode. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dpae

import (
	_ "embed"
	"encoding/xml"
	"regexp"
	"strings"
	"text/template"
)

//go:embed auth.xml
var AuthXml string

//go:embed dpae.xml
var DpaeXml string

var tAuth = template.Must(template.New("auth").Funcs(fmap).Parse(AuthXml)).Option("missingkey=error")
var tDpae = template.Must(template.New("dpae").Funcs(fmap).Parse(DpaeXml)).Option("missingkey=error")

var reChar = regexp.MustCompile(`[^0-9a-zA-ZéèêëàâäùûüîïôöçÉÈÊËÀÂÄÙÛÜÎÏÔÖÇ'\- ]`)

// limite à 32c
// élimine tout ce qui n'est pas dans le regexp
// excessivement radical
func fmtXML(s string) string {
	if len(s) > 32 {
		s = s[:32]
	}
	s = reChar.ReplaceAllString(s, " ")

	out := &strings.Builder{}
	xml.EscapeText(out, []byte(s))
	res := out.String()
	return res
}

var fmap = template.FuncMap{
	"XML": fmtXML,
}
