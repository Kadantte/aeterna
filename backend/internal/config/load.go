package config

import (
	"fmt"
	"reflect"

	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

func Load() Config {
	var cfg Config
	cfgValue := reflect.ValueOf(&cfg).Elem()
	cfgType := cfgValue.Type()
	sectionFields := make(map[string]reflect.Value, cfgType.NumField())
	for i := 0; i < cfgType.NumField(); i++ {
		fieldType := cfgType.Field(i)
		section := fieldType.Tag.Get("config")
		if section == "" {
			continue
		}
		sectionFields[section] = cfgValue.Field(i)
	}

	loaders := common.RegisteredModules()
	if len(loaders) == 0 {
		panic("no config modules registered")
	}

	loadedSections := make(map[string]string, len(loaders))
	for _, loader := range loaders {
		if previousModule, alreadyLoaded := loadedSections[loader.Section()]; alreadyLoaded {
			panic(fmt.Sprintf("duplicate config section %q registered by %s and %s", loader.Section(), previousModule, loader.Name()))
		}

		loadedSection, err := loader.LoadAndValidateAny()
		if err != nil {
			panic(fmt.Sprintf("config validation failed in %s: %v", loader.Name(), err))
		}

		field, ok := sectionFields[loader.Section()]
		if !ok {
			panic(fmt.Sprintf("config section %q from %s is not mapped in config.Config", loader.Section(), loader.Name()))
		}

		loadedValue := reflect.ValueOf(loadedSection)
		if !loadedValue.IsValid() {
			panic(fmt.Sprintf("config section %q from %s returned an invalid value", loader.Section(), loader.Name()))
		}
		if !loadedValue.Type().AssignableTo(field.Type()) {
			panic(fmt.Sprintf("config section %q type mismatch from %s: got %s, want %s", loader.Section(), loader.Name(), loadedValue.Type(), field.Type()))
		}
		field.Set(loadedValue)
		loadedSections[loader.Section()] = loader.Name()
	}

	for section := range sectionFields {
		if _, ok := loadedSections[section]; !ok {
			panic(fmt.Sprintf("no config module registered for section %q", section))
		}
	}
	return cfg
}
