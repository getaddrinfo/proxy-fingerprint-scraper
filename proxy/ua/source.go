package ua

type Source interface {
	List() ([]string, error)
	Random() (string, error)
}

var validSources = map[string]bool{
	"file": true,
}

func IsValidSource(it string) bool {
	return validSources[it]
}

func ValidSources() []string {
	out := []string{}

	for key := range validSources {
		out = append(out, key)
	}

	return out
}
