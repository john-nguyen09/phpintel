module github.com/john-nguyen09/phpintel

go 1.12

require (
	github.com/GeertJohan/go.rice v1.0.0
	github.com/JohannesKaufmann/html-to-markdown v0.0.0-20200323205911-a6f44902a8f4
	github.com/Masterminds/semver v1.5.0
	github.com/bep/debounce v1.2.0
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0
	github.com/jmhodges/levigo v1.0.1-0.20191214093932-ed89ec741d96
	github.com/john-nguyen09/go-phpparser v0.0.0-20201019010316-27e48374a0ac
	github.com/junegunn/fzf v0.0.0-20200515062533-d631c76e8d2d
	github.com/karrick/godirwalk v1.12.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.8.1
	github.com/seiflotfy/cuckoofilter v0.0.0-20200511222245-56093a4d3841
	github.com/stretchr/testify v1.5.1
)

replace github.com/hashicorp/go-immutable-radix v1.2.0 => github.com/john-nguyen09/go-immutable-radix v1.2.1-0.20200401082659-e38f7bb2dddd
