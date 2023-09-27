package common

type Permission uint32

const (
	PermissionViewHomePage Permission = 1 << iota
	PermissionUseAPI
	PermissionAdmin
)

func (p Permission) Has(perm Permission) bool {
	return (p & perm) == perm
}

func (p Permission) Remove(perm Permission) Permission {
	return p & ^perm
}

func (p Permission) Add(perm Permission) Permission {
	return p | perm
}

func (p Permission) List() []string {
	out := []string{}

	if p.Has(PermissionViewHomePage) {
		out = append(out, "VIEW_HOME_PAGE")
	}

	if p.Has(PermissionUseAPI) {
		out = append(out, "USE_API")
	}

	if p.Has(PermissionAdmin) {
		out = append(out, "ADMIN")
	}

	return out
}
