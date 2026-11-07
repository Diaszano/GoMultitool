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
	want := [5]MenuItemID{MenuBurn, MenuBackup, MenuHardware, MenuLegacy, MenuExit}
	for _, locale := range []Locale{LocalePtBR, LocaleEN} {
		catalog := For(locale)
		if catalog.ProductName != "GoMultitool" {
			t.Errorf("locale %q product name = %q", locale, catalog.ProductName)
		}
		for i, item := range catalog.Menu {
			if item.ID != want[i] || item.Label == "" || item.Hint == "" {
				t.Errorf("locale %q menu item %d = %+v, want ID %q", locale, i, item, want[i])
			}
		}
		if catalog.Loading == "" || catalog.Busy == "" {
			t.Errorf("locale %q is missing loading or busy text", locale)
		}
	}
}
