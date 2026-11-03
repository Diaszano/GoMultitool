package i18n

import "testing"

func TestDetectLocaleUsesFirstSupportedEnvironmentValue(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Locale
	}{
		{
			name: "LC_ALL wins",
			env: map[string]string{
				"LC_ALL":      "en_US.UTF-8",
				"LC_MESSAGES": "pt_BR.UTF-8",
				"LANG":        "pt_BR",
			},
			want: LocaleEN,
		},
		{
			name: "unsupported LC_ALL is skipped",
			env: map[string]string{
				"LC_ALL":      "C",
				"LC_MESSAGES": "en_US.UTF-8",
				"LANG":        "pt_BR",
			},
			want: LocaleEN,
		},
		{
			name: "language must match exactly",
			env:  map[string]string{"LC_ALL": "ptx_BR.UTF-8", "LANG": "en_US"},
			want: LocaleEN,
		},
		{
			name: "locale codes are case insensitive",
			env:  map[string]string{"LC_ALL": "PT_br.UTF-8", "LANG": "en_US.UTF-8"},
			want: LocalePtBR,
		},
		{
			name: "unsupported LC_MESSAGES is skipped",
			env:  map[string]string{"LC_ALL": "C", "LC_MESSAGES": "fr_FR.UTF-8", "LANG": "pt_BR"},
			want: LocalePtBR,
		},
		{
			name: "empty values are skipped",
			env:  map[string]string{"LC_ALL": "", "LC_MESSAGES": "", "LANG": "en_US"},
			want: LocaleEN,
		},
		{
			name: "unsupported locales default to pt-BR",
			env:  map[string]string{"LC_ALL": "C", "LC_MESSAGES": "fr_FR", "LANG": "de_DE"},
			want: LocalePtBR,
		},
		{name: "unset locales default to pt-BR", env: map[string]string{}, want: LocalePtBR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectLocale(func(key string) string { return tt.env[key] }); got != tt.want {
				t.Fatalf("DetectLocale() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCatalogContainsMenuInBothLocales(t *testing.T) {
	for _, locale := range []Locale{LocalePtBR, LocaleEN} {
		catalog := For(locale)
		if catalog.ProductName != "GoMultitool" {
			t.Errorf("locale %q product name = %q", locale, catalog.ProductName)
		}
		if len(catalog.Menu) != 10 {
			t.Errorf("locale %q menu entries = %d, want 10", locale, len(catalog.Menu))
		}
		if catalog.Menu[0].ID != MenuHardware || catalog.Menu[0].Label == "" {
			t.Errorf("locale %q hardware menu item = %+v", locale, catalog.Menu[0])
		}
		if catalog.Menu[9].ID != MenuExit || catalog.Menu[9].Label == "" {
			t.Errorf("locale %q exit menu item = %+v", locale, catalog.Menu[9])
		}
		for _, item := range catalog.Menu[1:9] {
			if item.ID == "" || item.Label == "" || item.Hint == "" {
				t.Errorf("locale %q has incomplete future menu item: %+v", locale, item)
			}
		}
	}
}
