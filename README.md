Go-Ecoball
========

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
ecoball  ecoclient ecowallet
```
If build in windows
Run ***build_windows*** in go-ecoball
```bash
%GOPATH%\src\github.com\ecoball\go-ecoball>build_windows
```
Then you will get a directory named **build**:
```bash
%GOPATH%\src\github.com\ecoball\go-ecoball\build>dir
ecoball.exe  ecoclient.exe ecowallet.exe
```

## Notes
This project used CGO, so set the CGO_ENABLED="1"

## ecoclient
### wallet
You must run the program ecowallet before you execute the wallet command
create wallet file
```
$./ecoclient wallet create --name $WALLETFILE --password $PASSWORD
```
open wallet file
```
$./ecoclient wallet open --name $WALLETFILE --password $PASSWORD
```
create keys to wallet
```
$./ecoclient wallet createkey --name $WALLETFILE
```
import privatekey to wallet
```
$./ecoclient wallet import --name $WALLETFILE --private $PRIVATEKEY
```
remove keys from wallet
```
$./ecoclient wallet remove --name $WALLETFILE --password $PASSWORD --public $PUBLICKEY
```
lock wallet
```
$./ecoclient wallet lock --name $WALLETFILE
```
unlock wallet
```
$./ecoclient wallet unlock --name $WALLETFILE --password $PASSWORD
```
list wallets
```
$./ecoclient wallet list
```
list keys
```
$./ecoclient wallet list_keys --name $WALLETFILE --password $PASSWORD
```
### account
create account
```
$./ecoclient wallet createaccount --account $ACCOUNTNAME --password $PASSWORD
```
### transfer
transfer aba  to another person
```
$ ./ecoclient transfer  --from $ADDRESS --to $ADDRESS --value $AMOUNT
```
### query
query account balance
```
$ ./ecoclient query balance --address $ADDRESS
```
### contract
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
### console
There are currently two modes, command line mode and console mode, which by default is command line mode.
If you want to open the console mode, you need to add option --console.
ecoclient console
```
$ ./ecoclient --console
ecoclient: \> $COMMAND
...
```
If you want to quit, please use the command exit
```
ecoclient: \> exit
```

### attach
By default, the command line tool connects to port 20678 and to localhost.
The default listener on node startup is port 20678,The configuration file **ecoball.toml** can change the option **http_port** to change the listening port.
The attach command can change the IP of the connected node and the corresponding port number.
```
$ ./ecoclient attach --ip=127.0.0.1 --port=20789
success!
attach http://127.0.0.1:20789 success!!!
```
## ecowallet
run ecoball

```
$ ./ecowallet
```


## ecoball
run ecoball

```
$ ./ecoball run
```

