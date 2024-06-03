// SPDX-License-Identifier: MIT

package i18n

import (
	"embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudfoundry-attic/jibber_jabber"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var bundle *goi18n.Bundle

func InitI18n(fsys embed.FS, folder string) error {
	bundle = goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	languages := []string{}

	files, err := fsys.ReadDir(folder)
	if err != nil {
		err = fmt.Errorf("unable to read i18n data: %w", err)
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fname := file.Name()
		lc := strings.Split(fname, ".")[0]

		lt, err := language.Parse(lc)
		if err != nil {
			log.Printf(
				"Unable to parse language code '%s': %v", lc, err,
			)
			continue
		}

		if _, err := bundle.LoadMessageFileFS(
			fsys, folder+"/"+file.Name(),
		); err != nil {
			err = fmt.Errorf("unable to load i18n data '%s': %w", file.Name(), err)
			return err
		}
		languages = append(languages, lt.String())
	}
	goi18n.NewLocalizer(bundle, languages...)
	return nil
}

func GetOsLocalizer() *goi18n.Localizer {
	langs := []string{}

	if l := os.Getenv("VSCORE_LANGUAGE"); l != "" {
		langs = append(langs, l)
	}

	if l, err := jibber_jabber.DetectLanguage(); err == nil {
		langs = append(langs, l)
	}
	langs = append(langs, "en")
	localizer := GetLocalizer(langs...)
	return localizer
}

func GetLocalizer(langs ...string) *goi18n.Localizer {
	if bundle == nil {
		return nil
	}
	return goi18n.NewLocalizer(bundle, langs...)
}

func T(
	localizer *goi18n.Localizer,
	key string,
) string {
	return TF(localizer, key, nil)
}

func TF(
	localizer *goi18n.Localizer,
	key string,
	data map[string]interface{},
	pluralCount ...int,
) string {
	if localizer == nil {
		return key
	}

	var pc interface{}
	if len(pluralCount) > 0 {
		pc = pluralCount[0]
	}
	cfg := goi18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
		PluralCount:  pc,
	}
	msg, err := localizer.Localize(&cfg)
	if err != nil {
		log.Printf("Failed to localize message with key: '%s': %s", key, err)
		if msg != "" {
			return msg
		}
		return key
	}
	return msg
}

func Tx(
	condition bool,
	localizer *goi18n.Localizer,
	key string,
) string {
	if !condition {
		l := goi18n.NewLocalizer(bundle, "en")
		return T(l, key)
	}
	return T(localizer, key)
}

func TFx(
	condition bool,
	localizer *goi18n.Localizer,
	key string,
	data map[string]interface{},
	pluralCount ...int,
) string {
	if !condition {
		l := goi18n.NewLocalizer(bundle, "en")
		return TF(l, key, data, pluralCount...)
	}
	return TF(localizer, key, data, pluralCount...)
}
