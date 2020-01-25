# Get rocksdb on WSL Ubuntu
```
$ GOOS=windows GOARCH=amd64 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc CGO_CFLAGS="-I/mnt/c/msys64/home/Thuan/rocksdb-5.18.3/include" CGO_LDFLAGS="-L/mnt/c/msys64/home/Thuan/rocksdb-5.18.3/build -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd" go get github.com/tecbot/gorocksdb
```


```
/mnt/c/msys64/home/Thuan/lz4-1.9.2

-g -O2 -I/mnt/c/msys64/home/Thuan/rocksdb-5.18.3/include -I/mnt/c/msys64/home/Thuan/lz4-1.9.2/lib

-g -O2 -L/mnt/c/msys64/home/Thuan/rocksdb-5.18.3/build -L/mnt/c/msys64/home/Thuan/lz4-1.9.2/lib -static-libstdc++ -lrocksdb -lstdc++ -lm -lz -llz4 -lpthread -lkernel32 -luser32 -lgdi32 -lwinspool -lshell32 -lole32 -loleaut32 -luuid -lcomdlg32 -ladvapi32
```