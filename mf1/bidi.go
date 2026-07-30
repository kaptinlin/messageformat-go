package mf1

const (
	leftToRightMark = "\u200e"
	rightToLeftMark = "\u200f"
)

func bidiMarkForLocale(localeID string) (string, error) {
	loc, err := parseStrictLocale(localeID)
	if err != nil {
		return "", WrapInvalidLocale(localeID)
	}
	if loc.GetTextInfo().Direction == "rtl" {
		return rightToLeftMark, nil
	}
	return leftToRightMark, nil
}
