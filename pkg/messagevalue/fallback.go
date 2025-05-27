package messagevalue

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// FallbackValue implements MessageValue for fallback values
// TypeScript original code:
//
//	export class FallbackValue implements MessageValue<'fallback'> {
//	  readonly type = 'fallback';
//	  constructor(
//	    public readonly source: string,
//	    public readonly locale?: string,
//	    public readonly dir?: Direction
//	  ) {}
//	  valueOf() { return this.source; }
//	  toString() { return `{${this.source}}`; }
//	  toParts() { return [{ type: 'fallback', value: `{${this.source}}`, source: this.source }]; }
//	  selectKeys() { return []; }
//	}
type FallbackValue struct {
	source string
	locale string
	dir    bidi.Direction
}

// NewFallbackValue creates a new fallback value
func NewFallbackValue(source, locale string) *FallbackValue {
	return &FallbackValue{
		source: source,
		locale: locale,
		dir:    bidi.DirAuto,
	}
}

// NewFallbackValueWithDir creates a new fallback value with explicit direction
func NewFallbackValueWithDir(source, locale string, dir bidi.Direction) *FallbackValue {
	return &FallbackValue{
		source: source,
		locale: locale,
		dir:    dir,
	}
}

func (fv *FallbackValue) Type() string {
	return "fallback"
}

func (fv *FallbackValue) Source() string {
	return fv.source
}

func (fv *FallbackValue) Dir() bidi.Direction {
	return fv.dir
}

func (fv *FallbackValue) Locale() string {
	return fv.locale
}

func (fv *FallbackValue) Options() map[string]interface{} {
	return nil
}

func (fv *FallbackValue) ToString() (string, error) {
	return fmt.Sprintf("{%s}", fv.source), nil
}

func (fv *FallbackValue) ToParts() ([]MessagePart, error) {
	return []MessagePart{
		&FallbackPart{
			source: fv.source,
			locale: fv.locale,
			dir:    fv.dir,
		},
	}, nil
}

func (fv *FallbackValue) ValueOf() (interface{}, error) {
	return fv.source, nil
}

func (fv *FallbackValue) SelectKeys(keys []string) ([]string, error) {
	// Fallback values don't participate in selection
	return []string{}, nil
}
