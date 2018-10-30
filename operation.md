# 一、基础工作
## 1、创建钱包
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient wallet create --name ubuntu --password ubuntu
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:41:21 | 200 |     313.948µs |       127.0.0.1 | POST     /wallet/create
success
```
## 2、打开钱包
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient wallet open --name ubuntu --password ubuntu
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
exist: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/wallet/ubuntu
[GIN] 2018/10/30 - 10:41:36 | 200 |     972.339µs |       127.0.0.1 | POST     /wallet/openWallet
success

```
## 3、导入私钥
这里导入4个私钥，注意先导入root账户的私钥
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient wallet import key -n ubuntu -k 0x33a0330cd18912c215c9b1125fab59e9a5ebfb62f0223bbea0c6c5f95e30b1c6
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:47:10 | 200 |    3.044481ms |       127.0.0.1 | POST     /wallet/importKey
publickey:0x0463613734b23e5dd247b7147b63369bf8f5332f894e600f7357f3cfd56886f75544fd095eb94dac8401e4986de5ea620f5a774feb71243e95b4dd6b83ca49910c
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient wallet import key -n ubuntu -k 0xc3e2cbed03aacc62d8f32045013364ea493f6d24e84f26bcef4edc2e9d260c0e
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:44:11 | 200 |    2.991919ms |       127.0.0.1 | POST     /wallet/importKey
publickey:0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ 
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient wallet import key -n ubuntu -k 0x5238ede4f91f6c4f5f1f195cbf674e08cb6a18ae351e474b8927db82d3e5ecf5
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:44:33 | 200 |    4.873582ms |       127.0.0.1 | POST     /wallet/importKey
publickey:0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient wallet import key -n ubuntu -k 0x105cb8f936eec87d35e42fc0f656ab4b7fc9a007cbf4554f829c44e528df6ce4
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:44:55 | 200 |    2.379846ms |       127.0.0.1 | POST     /wallet/importKey
publickey:0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient wallet import key -n ubuntu -k 0x68f2dcd39856206fa610546cc4f4611e5d4c3eb5e3f6bae3982348f949810745
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:45:35 | 200 |    2.831333ms |       127.0.0.1 | POST     /wallet/importKey
publickey:0x04b15d8efb9dcf3a086a69a0f6c334ebcb47d21293e36e1f22440185f1b7411a2cb3bcda2a91bf8ddeb71224ebd9233896766b355334b2c98b07f9ce9154c9dec9

```

## 4、创建账户
这里创建4个账户，分别是worker,worker1,worker2,worker3
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient create account -c root -n worker -o 0x04b15d8efb9dcf3a086a69a0f6c334ebcb47d21293e36e1f22440185f1b7411a2cb3bcda2a91bf8ddeb71224ebd9233896766b355334b2c98b07f9ce9154c9dec9
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:47:18 | 200 |      43.909µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 10:47:18 | 200 |     237.783µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient create account -c root -n worker1 -o 0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:48:00 | 200 |      44.242µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 10:48:00 | 200 |     221.309µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient create account -c root -n worker2 -o 0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:48:25 | 200 |      73.584µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 10:48:25 | 200 |     212.123µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient create account -c root -n worker3 -o 0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:48:55 | 200 |      63.897µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 10:48:55 | 200 |     393.202µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
## 5、 抵押ABA，换取CPU和NET资源
使用root账户给新创建的账户抵押
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m pledge -p root,worker,500,500
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:49:28 | 200 |     384.297µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 10:49:28 | 200 |     228.723µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m pledge -p root,worker1,500,500
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 11:12:55 | 200 |     141.683µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 11:12:55 | 200 |      194.47µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
# 二、调用系统合约
创建账号和抵押已经有了，这里主要是修改权限、注册出块节点、投票以及注册新链
## 1、修改权限
worker1给root授予active的权限

小窍门：可以先query account，找到权限信息，再新增或者修改
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424251822899200,"timestamp":1540869188275905399,"token":{"XYX":{"index":"XYX","balance":80}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}}},"contract":{"typeVm":0,"describe":null,"code":null,"abi":null},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":0,"delegated_aba":500,"used_byte":612,"available_byte":27520.421018944293,"limit_byte":28132.421018944293},"Cpu":{"staked_aba":0,"delegated_aba":500,"used_ms":7.407257,"available_ms":5358.426219818904,"limit_ms":5365.833476818904},"Votes":{"staked_aba":0,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}

ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m set_account -p worker1--'{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{"root":{"actor":13630584076887916544,"weight":1,"permission":"active"}}}' -s worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 11:45:24 | 200 |     176.989µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 11:45:24 | 200 |     357.525µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ 
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ 
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424251822899200,"timestamp":1540871125518236206,"token":{"XYX":{"index":"XYX","balance":80}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{"root":{"actor":13630584076887916544,"weight":1,"permission":"active"}}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}}},"contract":{"typeVm":0,"describe":null,"code":null,"abi":null},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":0,"delegated_aba":500,"used_byte":4.786916666666684,"available_byte":28408.958312467068,"limit_byte":28413.745229133736},"Cpu":{"staked_aba":0,"delegated_aba":500,"used_ms":-8.946455468974536,"available_ms":5428.438267056068,"limit_ms":5419.491811587093},"Votes":{"staked_aba":0,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ 

```
# 三、部署合约，调用合约
## 1、部署合约
在worker账户部署token合约
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract deploy -p ../test/contract/testToken/token_api.wasm -n worker --d "token contract" --ap ../test/contract/testToken/simple_token.abi
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:51:18 | 200 |     202.653µs |       127.0.0.1 | GET      /wallet/getPublicKeys
0x04b15d8efb9dcf3a086a69a0f6c334ebcb47d21293e36e1f22440185f1b7411a2cb3bcda2a91bf8ddeb71224ebd9233896766b355334b2c98b07f9ce9154c9dec9,0x0463613734b23e5dd247b7147b63369bf8f5332f894e600f7357f3cfd56886f75544fd095eb94dac8401e4986de5ea620f5a774feb71243e95b4dd6b83ca49910c,0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f,0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee,0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f
[GIN] 2018/10/30 - 10:51:18 | 200 |      399.68µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```

## 2、调用合约
### (1)、创建代币
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n worker -m create -p '["worker", "800", "XYX"]'
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:54:14 | 200 |     138.873µs |       127.0.0.1 | GET      /wallet/getPublicKeys
issuer is  string  worker
max_supply is  int64  800
token_id is  string  XYX
[GIN] 2018/10/30 - 10:54:14 | 200 |     210.527µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
### (2)、分发
给worker1分发100 XYX
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n worker -m issue -p '{"to": "worker1", "amount": "100", "token_id": "XYX"}'
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 10:55:51 | 200 |      78.811µs |       127.0.0.1 | GET      /wallet/getPublicKeys
to is  string  worker1
amount is  int64  100
token_id is  string  XYX
[GIN] 2018/10/30 - 10:55:51 | 200 |     621.761µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
查看分发是否成功
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424251822899200,"timestamp":1540867681307614272,"token":{"XYX":{"index":"XYX","balance":100}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}}},"contract":{"typeVm":0,"describe":null,"code":null,"abi":null},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":0,"delegated_aba":0,"used_byte":0,"available_byte":0,"limit_byte":0},"Cpu":{"staked_aba":0,"delegated_aba":0,"used_ms":0,"available_ms":0,"limit_ms":0},"Votes":{"staked_aba":0,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}

```
worker1的余额为100 XYX，证明分发成功
### (3)、转账
worker1给worker2转账20 XYX
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n worker -m transfer -p '{"from": "worker1", "to": "worker2", "amount": "20", "token_id": "XYX"}' -s worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 11:10:49 | 200 |      51.962µs |       127.0.0.1 | GET      /wallet/getPublicKeys
from is  string  worker1
to is  string  worker2
amount is  int64  20
token_id is  string  XYX
[GIN] 2018/10/30 - 11:10:49 | 200 |     240.575µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
查看转账是否成功
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424251822899200,"timestamp":1540869188275905399,"token":{"XYX":{"index":"XYX","balance":80}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}}},"contract":{"typeVm":0,"describe":null,"code":null,"abi":null},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":0,"delegated_aba":500,"used_byte":612,"available_byte":27520.421018944293,"limit_byte":28132.421018944293},"Cpu":{"staked_aba":0,"delegated_aba":500,"used_ms":7.407257,"available_ms":5358.426219818904,"limit_ms":5365.833476818904},"Votes":{"staked_aba":0,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker2
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424252359770112,"timestamp":1540867707335842061,"token":{"XYX":{"index":"XYX","balance":20}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x013f866d8dfdddf14a6d5ff791766a8bf4ddc270":{"actor":[1,63,134,109,141,253,221,241,74,109,95,247,145,118,106,139,244,221,194,112],"weight":1}},"accounts":{}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x013f866d8dfdddf14a6d5ff791766a8bf4ddc270":{"actor":[1,63,134,109,141,253,221,241,74,109,95,247,145,118,106,139,244,221,194,112],"weight":1}},"accounts":{}}},"contract":{"typeVm":0,"describe":null,"code":null,"abi":null},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":0,"delegated_aba":0,"used_byte":0,"available_byte":0,"limit_byte":0},"Cpu":{"staked_aba":0,"delegated_aba":0,"used_ms":0,"available_ms":0,"limit_ms":0},"Votes":{"staked_aba":0,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}

```
worker1的余额为80 XYX，worker2的余额为20 XYX，证明转账成功