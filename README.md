Go-Ecoball
-------

## Depends
1. Firstly,you need install [protoc](https://github.com/google/protobuf/blob/master/src/README.md) 
2. Golang version >= 1.9
3. Then you need install golang proto tools:
```bash
go get github.com/gogo/protobuf/protoc-gen-gofast
```
4. If build in windows, you must install [mingw](http://www.mingw.org/)

## Build
Run ***make all*** in go-ecoball
```bash
$:~/go/src/github.com/ecoball/go-ecoball$ make
```
Then you will get a directory named **build**:
```bash
~/go/src/github.com/ecoball/go-ecoball$ ls build/
ecoball  ecoclient
```
If build in windows
Run ***build_windows*** in go-ecoball
```bash
%GOPATH%\src\github.com\ecoball\go-ecoball>build_windows
```
Then you will get a directory named **build**:
```bash
%GOPATH%\src\github.com\ecoball\go-ecoball\build>dir
ecoball.exe  ecoclient.exe
```

## Notes
This project used CGO, so set the CGO_ENABLED="1"

## ecoclient
create wallet file
```
$./ecoclient wallet create --name $WALLETFILE --password $PASSWORD
```
create account
```
$./ecoclient wallet createaccount --account $ACCOUNTNAME --password $PASSWORD
```
list account
```
$./ecoclient wallet list --password $PASSWORD
```
transfer aba  to another person
```
$ ./ecoclient transfer  --from $ADDRESS --to $ADDRESS --value $AMOUNT
```

query account balance
```
$ ./ecoclient query balance --address $ADDRESS
```

deploy contract,you will get contract address
```
$ ./ecoclient contract deploy -p $CONTRACTFILE -n $CONTRACTNAME --d $DESCRIPTION
success!
0x0133ac14c0633a2a5e09e7109dcb560f6f5270e1
```

invoke contract
```
$ ./ecoclient contract invoke -n $CONTRACTNAME -m $METHORD -p $PARA1 $PARA2 $PARA3 ...
```

ecoclient console
```
$ ./ecoclient $COMMAND
ecoclient: \> $COMMAND
...
```
## ecoball
run ecoball

```
$ ./ecoball --name=$WALLETFILE --password=$PASSWORD run
```

