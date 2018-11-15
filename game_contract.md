# 一、game contract action
the length of user name should less than 10
```
//create action is used to create a new game
int create(char* player1, char* player2);
//restart action is used to restart a exsit game
int restart(char* player1, char* player2, char* restarter);
//close action is used to close a exsit game
int close(char* player1, char* player2, char* closer);
//follow action is used to move in chess
int follow(char* player1, char* player2, char* host, int row, int column);
```

# 二、contract deploy and invoke
## 1、deploy contract
```
./ecoclient contract deploy -p game.wasm -n tictactoe -d game -i game.abi
```
## 2、invoke create action
```
./ecoclient contract invoke -n tictactoe -m create -p {"player1":"user1","player2":"user2"} -i user1
```
## 3、invoke close action
user2 close the game
```
./ecoclient contract invoke -m close -p {"player1":"user1","player2":"user2","closer":"player2"} -n tictactoe -i user2
```
## 4、invoke restart action
user2 restart the game
```
./ecoclient contract invoke -m restart -p {"player1":"user1","player2":"user2","restart":"player2"} -n tictactoe -i user2
```
## 5、invoke follow action
user1 win the game
```
./ecoclient contract invoke -n tictactoe -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"2","column":"2"}  -i user1
./ecoclient contract invoke -n tictactoe -m follow -p {"player1":"user1","player2":"user2","host":"user2","row":"1","column":"2"} -i user2
./ecoclient contract invoke -n tictactoe -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"3","column":"1"} -i user1
./ecoclient contract invoke -n tictactoe -m follow -p {"player1":"user1","player2":"user2","host":"user2","row":"1","column":"3"} -i user2
./ecoclient contract invoke -n tictactoe -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"1","column":"3"} -i user1
```
user1 and user2 don't win the game，the chess is full
```
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"2","column":"2"} -n tictactoe -i user1
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user2","row":"1","column":"2"} -n tictactoe -i user2
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"1","column":"3"} -n tictactoe -i user1
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user2","row":"3","column":"1"} -n tictactoe -i user2
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"2","column":"1"} -n tictactoe -i user1
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user2","row":"2","column":"3"} -n tictactoe -i user2
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"3","column":"2"} -n tictactoe -i user1
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user2","row":"1","column":"1"} -n tictactoe -i user2
./ecoclient contract invoke -m follow -p {"player1":"user1","player2":"user2","host":"user1","row":"3","column":"3"} -n tictactoe -i user1
```
