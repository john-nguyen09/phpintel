module github.com/john-nguyen09/phpintel

go 1.12

require (
	github.com/GeertJohan/go.rice v1.0.0
	github.com/Masterminds/semver v1.5.0
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/evorts/html-to-markdown v0.0.3
	github.com/hashicorp/go-immutable-radix v1.2.0
	github.com/jmhodges/levigo v1.0.0
	github.com/john-nguyen09/go-phpparser v0.0.0-20200507091038-45e522afc830
	github.com/karrick/godirwalk v1.12.0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.8.1
	github.com/sahilm/fuzzy v0.1.0
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
)

replace github.com/hashicorp/go-immutable-radix v1.2.0 => github.com/john-nguyen09/go-immutable-radix v1.2.1-0.20200401082659-e38f7bb2dddd

replace github.com/sahilm/fuzzy v0.1.0 => github.com/john-nguyen09/fuzzy v0.1.1-0.20200503031448-486a244c615b
