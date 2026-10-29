package config

import "testing"

func TestExigirCorreo(t *testing.T) {
	casos := []struct {
		c  Config
		ok bool
	}{
		{Config{ResendClave: "re_x"}, true},
		{Config{SMTP: "mail:465", SMTPUsuario: "u", SMTPClave: "c"}, true},
		// El Mailpit de desarrollo (SMTP sin usuario) no alcanza fuera de desarrollo.
		{Config{SMTP: "127.0.0.1:1026"}, false},
		{Config{}, false},
	}
	for _, c := range casos {
		if err := c.c.ExigirCorreo(); (err == nil) != c.ok {
			t.Errorf("%+v: err = %v", c.c, err)
		}
	}
}
