package main

import (
	"io"
	"text/template"
)

// RenderTerraformConfig renders TeamRoles using team.tf.tpl into target Writer
func RenderTerraformConfig(tr TeamRoles, wr io.Writer) error {
	t := template.Must(
		template.New("team.tf.tpl").
			ParseFiles("templates/team.tf.tpl"))
	return t.Execute(wr, tr)
}
