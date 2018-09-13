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
You must run the program ecowallet before you execute the wallet command.
#### attach wallet server
By default, the command line tool connects to port 20679 and to localhost with the wallet service.
The default listener on wallet node startup is port 20679,The configuration file **ecoball.toml** can change the option **wallet_http_port** to change the listening port.
The attach command can change the IP of the connected wallet node and the corresponding port number.
```
$./ecoclient wallet attach --ip $WALLETSERVERIP --port $WALLETSERVERPORT
```
#### create wallet file
```
$./ecoclient wallet create --name $WALLETFILE --password $PASSWORD
```
#### open wallet file
```
$./ecoclient wallet open --name $WALLETFILE --password $PASSWORD
```
#### create keys to wallet
```
$./ecoclient wallet createkey --name $WALLETFILE
```
#### import privatekey to wallet
```
$./ecoclient wallet import --name $WALLETFILE --private $PRIVATEKEY
```
#### remove keys from wallet
```
$./ecoclient wallet remove --name $WALLETFILE --password $PASSWORD --public $PUBLICKEY
```
#### lock wallet
```
$./ecoclient wallet lock --name $WALLETFILE
```
#### unlock wallet
```
$./ecoclient wallet unlock --name $WALLETFILE --password $PASSWORD
```
#### list wallets
```
$./ecoclient wallet list
```
#### list keys
```
$./ecoclient wallet list_keys --name $WALLETFILE --password $PASSWORD
```
### account
#### create account
```
$./ecoclient create account --creator $CREATORNAME --name $ACCOUNTNAME --owner $OWNERPUBLICKEY --active $ACTIVEPUBLICKEY
```
### transfer
#### transfer aba  to another person
```
$ ./ecoclient transfer  --from $ADDRESS --to $ADDRESS --value $AMOUNT
```
### query
#### query all chainId
```
$ ./ecoclient query listchain
```
#### query account's info
```
$ ./ecoclient query account -n $ACCOUNTNAME
```
### contract
#### deploy contract,you will get contract address
```
$ ./ecoclient contract deploy -p $CONTRACTFILE -n $CONTRACTNAME --d $DESCRIPTION --ap $ABIFILE
success!
0x0133ac14c0633a2a5e09e7109dcb560f6f5270e1
```

#### invoke contract
```
$ ./ecoclient contract invoke -n $CONTRACTNAME -m $METHORD -p $PARA1 $PARA2 $PARA3 ...
```
#### register new chain
```
$ ./ecoclient contract invoke -n $CONTRACTNAME -m reg_chain -p $PARA1,$PARA2,$PARA3
```
#### set account
```
$ ./ecoclient contract invoke -n $CONTRACTNAME -m set_account -p $ACCOUNTNAME--$PERMISSION
```
#### pledge transcation
```
$ ./ecoclient contract invoke -n $CONTRACTNAME -m pledge -p $PARA1,$PARA2,$PARA3 ...
```
#### voting to be producer
```
$ ./ecoclient contract invoke -n $CONTRACTNAME -m vote -p $PARA1,$PARA2,$PARA3 ...
```
#### register to producer
```
$ ./ecoclient contract invoke -n $CONTRACTNAME -m reg_prod -p $PARA1,$PARA2,$PARA3 ...
```

### console
There are currently two modes, command line mode and console mode, which by default is command line mode.
If you want to open the console mode, you need to add option --console.
#### ecoclient console
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
By default, the command line tool connects to port 20678 and to localhost with the node service.
The default listener on node startup is port 20678,The configuration file **ecoball.toml** can change the option **http_port** to change the listening port.
The attach command can change the IP of the connected node and the corresponding port number.
```
$ ./ecoclient attach --ip=NODESERVERIP --port=NODESERVERPORT
success!
attach http://127.0.0.1:20789 success!!!
```
## ecowallet
run ecowallet

```
$ ./ecowallet
```


## ecoball
run ecoball

```
$ ./ecoball run
```

