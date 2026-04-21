package tests

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
)

func normalizeExternalTestLocales(locales any) []string {
	switch v := locales.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []string:
		return v
	default:
		return nil
	}
}

func buildExternalTestMessageFormat(locales any, source any, options *messageformat.MessageFormatOptions) (*messageformat.MessageFormat, error) {
	opts := []messageformat.Option{}
	if options != nil {
		opts = append(opts, messageformat.Options(*options))
	}

	switch value := source.(type) {
	case string:
		return messageformat.Parse(normalizeExternalTestLocales(locales), value, opts...)
	case datamodel.Message:
		return messageformat.Compile(normalizeExternalTestLocales(locales), value, opts...)
	default:
		return nil, fmt.Errorf("unsupported test source type %T", source)
	}
}
