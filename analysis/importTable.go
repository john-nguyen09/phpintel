package analysis

type ImportTable map[string]TypeString

func newImportTable() ImportTable {
	return map[string]TypeString{}
}
