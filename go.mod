module github.com/john-nguyen09/phpintel

go 1.12

require (
	github.com/GeertJohan/go.rice v1.0.0
	github.com/Masterminds/semver v1.5.0
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/evorts/html-to-markdown v0.0.3
	github.com/jmhodges/levigo v1.0.0
	github.com/john-nguyen09/go-phpparser v0.0.0-20200204091501-d315e8e7d929
	github.com/karrick/godirwalk v1.12.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.8.1
	github.com/smacker/go-tree-sitter v0.0.0-20200219092318-fb146ff28ff0
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
)

replace github.com/smacker/go-tree-sitter => github.com/john-nguyen09/go-tree-sitter v0.0.0-20200223043131-e9dd2bb3a55b
