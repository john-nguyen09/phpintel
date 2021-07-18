module github.com/john-nguyen09/phpintel

go 1.12

require (
	github.com/FastFilter/xorfilter v0.0.0-20210113222943-0cd20fbc5711
	github.com/GeertJohan/go.rice v1.0.0
	github.com/JohannesKaufmann/html-to-markdown v0.0.0-20200323205911-a6f44902a8f4
	github.com/Masterminds/semver v1.5.0
	github.com/akrylysov/pogreb v0.10.1
	github.com/bep/debounce v1.2.0
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/hashicorp/go-immutable-radix v1.2.0
	github.com/jmhodges/levigo v1.0.1-0.20191214093932-ed89ec741d96
	github.com/john-nguyen09/go-phpparser v0.0.0-20210626125202-106d065be921
	github.com/junegunn/fzf v0.0.0-20200515062533-d631c76e8d2d
	github.com/karrick/godirwalk v1.12.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/hashicorp/go-immutable-radix v1.2.0 => github.com/john-nguyen09/go-immutable-radix v1.2.1-0.20200401082659-e38f7bb2dddd
