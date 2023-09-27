package api

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/common"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/database"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy"
)

type RenderHomeTemplateParams struct {
	Count uint64

	CanUseApi bool
	Admin     bool

	RenderProxies bool
	Proxies       []string
}

type RenderUserData struct {
	Id         uint64
	PermString string
}

type RenderAdminTemplateParams struct {
	Users []RenderUserData
}

func RenderHomeTemplate(count uint64, user database.GetAuthResult, manager proxy.Manager) (string, error) {
	str := `fingerprint generation service
count: {{.Count}}

auth:
Your token can be provided either via the token query parameter, or via the Authorization header. Either is accepted.

routes:
GET /
Shows this page

{{- if .Admin }}

GET /admin
Shows all the users registered with this app, and their permissions
{{- end}}

{{- if .CanUseApi }}

GET /api/fingerprints{/raw}
Lists all the fingerprints in the database

GET /api/fingerprints/random
Returns a random fingerprint from the database (json) 

GET /api/fingerprints/{fingerprint.id}
Returns a specific fingerprint from the database (json)
{{- end}}


{{ if .RenderProxies -}}
proxies:
{{ range .Proxies -}}
- {{.}}
{{ end }}
{{ end }}
`
	tmpl := template.Must(template.New("index").Parse(str))
	canUseApi := user.Permissions.Has(common.PermissionUseAPI)
	isAdmin := user.Permissions.Has(common.PermissionAdmin)

	var proxies []string
	renderProxies := manager != nil

	if manager != nil {
		proxies = manager.IPs()
	}

	data := RenderHomeTemplateParams{
		Count:         count,
		Proxies:       proxies,
		RenderProxies: renderProxies,
		CanUseApi:     canUseApi,
		Admin:         isAdmin,
	}

	var out bytes.Buffer

	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}

	return out.String(), nil
}

func RenderAdminTemplate(users []database.GetUserResult) (string, error) {
	str := `fingerprint generation service

users:
{{ range .Users -}}
id={{ .Id }} permissions={{.PermString}}
{{ end }}
`
	tmpl := template.Must(template.New("index").Parse(str))
	var renderUsers []RenderUserData

	for _, user := range users {
		var permString = strings.Join(user.Permissions.List(), ", ")

		if permString == "" {
			permString = "None"
		} else {
			permString = fmt.Sprintf("(%s)", permString)
		}

		renderUsers = append(renderUsers, RenderUserData{
			Id:         user.UserId,
			PermString: permString,
		})
	}

	data := RenderAdminTemplateParams{
		Users: renderUsers,
	}

	var out bytes.Buffer

	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}

	return out.String(), nil
}
