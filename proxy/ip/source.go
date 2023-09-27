package ip

type Source interface {
	Load() ([]string, error)
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
