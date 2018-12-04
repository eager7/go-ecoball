# 一、基础操作
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
这里创建4个账户，分别是worker, worker1, worker2, worker3
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

## 6、转账
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient transfer --from root --to worker --value 1000
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:33:55 | 200 |      57.749µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:33:55 | 200 |     184.881µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient transfer --from root --to worker1 --value 1000
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:34:22 | 200 |     158.679µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:34:22 | 200 |     246.873µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient transfer --from root --to worker2 --value 1000
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:34:25 | 200 |      39.374µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:34:25 | 200 |     223.414µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ 
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ 
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient transfer --from root --to worker3 --value 1000
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:35:05 | 200 |      67.723µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:35:05 | 200 |     726.774µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
# 二、调用系统合约
所有系统命令都以system开头，可以通过ecoclient system查询子命令
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
NAME:
   ecoclient system - you can pledge, set permission, vote, register to be producer and register a new chain

USAGE:
   ecoclient system command [command options] [args]

COMMANDS:
     set_perm       set_perm
     pledge         pledge
     cancel_pledge  cancel_pledge
     reg_prod       register producer
     vote           vote
     reg_chain      register chain

OPTIONS:
   --help, -h  show help
   
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system pledge
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
NAME:
   ecoclient system pledge - pledge

USAGE:
   ecoclient system pledge [command options] [arguments...]

OPTIONS:
   --payer value, -p value      resource payer
   --user value, -u value       resource user
   --cpu value, -s value        ABA pledged for cpu
   --net value, -n value        ABA pledged for net
   --chainHash value, -c value  chain hash(the default is the main chain hash)
   

```
## 1、抵押
root给abatoken抵押200个ABA，其中100个用于获取CPU，100个用于获取net
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system pledge -p root -u abatoken -s 100 -n 100
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 15:44:25 | 200 |     130.706µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 15:44:25 | 200 |     567.141µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":13630584076887916544,"permission":"owner"},"payload":{"method":"cGxlZGdl","param":["root","abatoken","100","100"]},"console":"pledge success!\n"}]

```
## 2、取消抵押
root取消抵押给abatoken的200个ABA
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system cancel_pledge -p root -u abatoken -s 100 -n 100
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 15:53:33 | 200 |      49.672µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 15:53:33 | 200 |     353.131µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":13630584076887916544,"permission":"owner"},"payload":{"method":"Y2FuY2VsX3BsZWRnZQ==","param":["root","abatoken","100","100"]},"console":"cancel pledge\n"}]

```

## 3、设置权限
abatoken给root授权
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system set_perm -a abatoken -p '{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x01ebfb5f37ff615ac3e506b9692e66bf9c4d00aa":{"actor":[1,235,251,95,55,255,97,90,195,229,6,185,105,46,102,191,156,77,0,170],"weight":1}},"accounts":{"root":{"actor":13630584076887916544,"weight":1,"permission":"active"}}}'
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 17:12:05 | 200 |      45.341µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 17:12:05 | 200 |      457.43µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":3588694083440214016,"permission":"owner"},"payload":{"method":"c2V0X2FjY291bnQ=","param":["abatoken","{\"perm_name\":\"active\",\"parent\":\"owner\",\"threshold\":1,\"keys\":{\"0x01ebfb5f37ff615ac3e506b9692e66bf9c4d00aa\":{\"actor\":[1,235,251,95,55,255,97,90,195,229,6,185,105,46,102,191,156,77,0,170],\"weight\":1}},\"accounts\":{\"root\":{\"actor\":13630584076887916544,\"weight\":1,\"permission\":\"active\"}}}"]},"console":"set account success\n"}]
warning: transaction executed locally, but may not be confirmed by the network yet

```

## 4、注册为出块节点
注意：注册新链需要先抵押200个ABA(暂定)，如果余额不够可以用root转账
下面是抵押ABA，必须是自己抵押给自己
首先使用root给worker和worker1各转1000个ABA
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system pledge -p worker1 -u worker1 -s 500 -n 500
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 17:23:07 | 200 |      45.124µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 17:23:07 | 200 |     298.075µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":16514424251822899200,"permission":"owner"},"payload":{"method":"cGxlZGdl","param":["worker1","worker1","500","500"]},"console":"pledge success!\n"}]
warning: transaction executed locally, but may not be confirmed by the network yet
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system pledge -p worker -u worker -s 500 -n 500
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 17:23:16 | 200 |      77.175µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 17:23:16 | 200 |     878.402µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":16514424251286028288,"permission":"owner"},"payload":{"method":"cGxlZGdl","param":["worker","worker","500","500"]},"console":"pledge success!\n"}]
warning: transaction executed locally, but may not be confirmed by the network yet

```
worker和worker1分别给自己抵押1000个ABA，获得注册出块节点的资格
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system reg_prod -p worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 17:23:25 | 200 |       85.05µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 17:23:25 | 200 |     459.444µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":16514424251822899200,"permission":"owner"},"payload":{"method":"cmVnX3Byb2Q=","param":["worker1"]},"console":"register producer success\n"}]
warning: transaction executed locally, but may not be confirmed by the network yet
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system reg_prod -p worker
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 17:23:27 | 200 |       58.75µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 17:23:27 | 200 |     512.342µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":16514424251286028288,"permission":"owner"},"payload":{"method":"cmVnX3Byb2Q=","param":["worker"]},"console":"register producer success\n"}]
warning: transaction executed locally, but may not be confirmed by the network yet

```

## 5、投票
worker1给worker和worker1投票
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system vote -v worker1 -f worker -s worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 17:30:56 | 200 |      50.087µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 17:30:56 | 200 |     314.967µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":16514424251822899200,"permission":"owner"},"payload":{"method":"dm90ZQ==","param":["worker1","worker","worker1"]},"console":"vote success\n"}]
warning: transaction executed locally, but may not be confirmed by the network yet

```

## 6、注册新链
注意：注册新链需要先抵押200个ABA(暂定)，如果余额不够可以用root转账
下面是抵押ABA，必须是自己抵押给自己
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient system reg_chain -a worker -s solo -p 0x04b15d8efb9dcf3a086a69a0f6c334ebcb47d21293e36e1f22440185f1b7411a2cb3bcda2a91bf8ddeb71224ebd9233896766b355334b2c98b07f9ce9154c9dec9
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/11/09 - 17:41:54 | 200 |       49.31µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/11/09 - 17:41:54 | 200 |      369.66µs |       127.0.0.1 | POST     /wallet/signTransaction
[{"account":13630584076887916544,"permission":{"actor":16514424251286028288,"permission":"owner"},"payload":{"method":"cmVnX2NoYWlu","param":["worker","solo","0x04b15d8efb9dcf3a086a69a0f6c334ebcb47d21293e36e1f22440185f1b7411a2cb3bcda2a91bf8ddeb71224ebd9233896766b355334b2c98b07f9ce9154c9dec9"]},"console":""}]
warning: transaction executed locally, but may not be confirmed by the network yet

```

# 三、用户合约
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
查看worker1的账户，发现worker1的余额为100 XYX，证明分发成功
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424251822899200,"timestamp":1540867681307614272,"token":{"XYX":{"index":"XYX","balance":100}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}}},"contract":{"typeVm":0,"describe":null,"code":null,"abi":null},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":0,"delegated_aba":0,"used_byte":0,"available_byte":0,"limit_byte":0},"Cpu":{"staked_aba":0,"delegated_aba":0,"used_ms":0,"available_ms":0,"limit_ms":0},"Votes":{"staked_aba":0,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}

```
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
查看worker1和worker2的账户，发现worker1的余额为80 XYX，worker2的余额为20 XYX，证明转账成功
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
