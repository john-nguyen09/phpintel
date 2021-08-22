module github.com/john-nguyen09/phpintel

go 1.12

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3 // indirect
	github.com/FastFilter/xorfilter v0.0.0-20210618184958-3504b2eb9fb2
	github.com/GeertJohan/go.rice v1.0.2
	github.com/JohannesKaufmann/html-to-markdown v1.3.0
	github.com/Masterminds/semver v1.5.0
	github.com/PuerkitoBio/goquery v1.7.1 // indirect
	github.com/akrylysov/pogreb v0.10.1
	github.com/bep/debounce v1.2.0
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jmhodges/levigo v1.0.1-0.20191214093932-ed89ec741d96
	github.com/john-nguyen09/go-phpparser v0.0.0-20210822134650-18371c1922eb
	github.com/junegunn/fzf v0.0.0-20210817074024-3f90fb42d887
	github.com/karrick/godirwalk v1.16.1
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/hashicorp/go-immutable-radix v1.2.0 => github.com/john-nguyen09/go-immutable-radix v1.2.1-0.20200401082659-e38f7bb2dddd
