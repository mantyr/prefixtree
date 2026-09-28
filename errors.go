package prefixtree

// Готовые ошибки без подстановок, создаются один раз и не аллоцируются при поиске
// Экземпляры общие: менять поля *codes.Error после получения нельзя
var (
	// PathNotFound
	errNotFound = newError(PathNotFound, "not found")

	// PathAlreadyExists
	errPathInUse      = newError(PathAlreadyExists, "path already in use")
	errCatchAllExists = newError(PathAlreadyExists, "CatchAll already exists")

	// InvalidPath
	errEmptyPath           = newError(InvalidPath, "empty path")
	errExpectedEOF         = newError(InvalidPath, "expected EOF")
	errExpectedStaticToken = newError(InvalidPath, "expected Static token")
	errEmptyTokenValue     = newError(InvalidPath, "empty token value")

	// EmptyValue
	errEmptyValue = newError(EmptyValue, "empty value")

	// NotChanged
	errNotChanged = newError(NotChanged, "not changed")
)
