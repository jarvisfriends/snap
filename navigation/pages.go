package navigation

const (
	// PageIDHome and PageIDSettings are the stable page ID strings
	// used when registering pages with any Navigator. Router and tests should use
	// these rather than bare string literals so renames stay in one place.
	PageIDHome     = "home"
	PageIDSettings = "settings"

	pageIDHome     = PageIDHome
	pageIDSettings = PageIDSettings

	pageHome     = "Home"
	pageSettings = "Settings"
)
