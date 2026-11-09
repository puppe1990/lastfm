package i18n

import "testing"

func TestDefaultCatalog_english(t *testing.T) {
	c := DefaultCatalog()
	if got := c.T("auth.welcome"); got != "Welcome back" {
		t.Errorf("T(auth.welcome) = %q", got)
	}
	if got := c.T("auth.signup_prompt"); got != "Don't have an account?" {
		t.Errorf("T(auth.signup_prompt) = %q", got)
	}
}

func TestNewCatalog_portuguese(t *testing.T) {
	c := NewCatalog("pt-BR")
	if got := c.T("auth.welcome"); got != "Bem-vindo de volta" {
		t.Errorf("T(auth.welcome) = %q", got)
	}
	if c.HTMLLang() != "pt-BR" {
		t.Errorf("HTMLLang() = %q, want pt-BR", c.HTMLLang())
	}
	if got := c.T("listen.error"); got != "não deu para falar com o Last.fm" {
		t.Errorf("T(listen.error) = %q", got)
	}
}
