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
创建账号和抵押已经有了，这里主要是修改权限、注册出块节点、投票以及注册新链
## 1、修改权限
worker1给root授予active的权限

注意，权限修改是覆盖式的，如果修改active权限就会覆盖整个active权限（不会影响owner权限），这里有个小窍门：可以先query account，找到权限信息，再新增或者修改
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
## 2、注册新链
注意：注册新链需要先抵押200个ABA(暂定)，如果余额不够可以用root转账
下面是抵押ABA，必须是自己抵押给自己
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m pledge -p worker,worker,250,250
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:42:14 | 200 |      99.972µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:42:14 | 200 |     287.065µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m pledge -p worker1,worker1,250,250
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:42:25 | 200 |       60.26µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:42:25 | 200 |     194.487µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
接下来正式注册新链
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m reg_chain -p worker,solo,0x04b15d8efb9dcf3a086a69a0f6c334ebcb47d21293e36e1f22440185f1b7411a2cb3bcda2a91bf8ddeb71224ebd9233896766b355334b2c98b07f9ce9154c9dec9
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:47:17 | 200 |      46.722µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:47:17 | 200 |     175.791µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
## 3、注册为出块节点
注意：注册为出块节点需要先抵押200个ABA(暂定)，如果余额不够可以用root转账
将worker和worker1注册为出块节点
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m reg_prod -p worker -s worker
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 14:58:04 | 200 |      105.79µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 14:58:04 | 200 |     325.226µs |       127.0.0.1 | POST     /wallet/signTransaction
success
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m reg_prod -p worker1 -s worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 15:05:43 | 200 |      45.241µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 15:05:43 | 200 |     260.624µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
## 4、投票
worker1给worker和worker1投票
注意：投票是以自己抵押的ABA为准，前面worker和worker1都抵押了250个CPU和NET，因此他们都有500票
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m vote -p worker1,worker1,worker -s worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 15:06:49 | 200 |      54.054µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 15:06:49 | 200 |      412.86µs |       127.0.0.1 | POST     /wallet/signTransaction
success

```
查看worker1的账户信息，可以知道投票已经成功，worker和worker1各得了500票
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker1
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424251822899200,"timestamp":1540883211135520835,"token":{"ABA":{"index":"ABA","balance":500},"XYX":{"index":"XYX","balance":80}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{"root":{"actor":13630584076887916544,"weight":1,"permission":"active"}}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e":{"actor":[1,29,9,215,248,119,65,73,75,220,105,182,123,139,45,196,236,212,156,126],"weight":1}},"accounts":{}}},"contract":{"typeVm":0,"describe":null,"code":null,"abi":null},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":250,"delegated_aba":500,"used_byte":297.82760943952525,"available_byte":47554.14096894986,"limit_byte":47851.96857838939},"Cpu":{"staked_aba":250,"delegated_aba":500,"used_ms":116.34613231063383,"available_ms":9010.692552227138,"limit_ms":9127.03868453777},"Votes":{"staked_aba":500,"producers":{"16514424251286028288":500,"16514424251822899200":500}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}

```
## 5、取消抵押
root减少给worker的抵押，原先是250个ABA的CPU和NET，现在减少250个
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient contract invoke -n root -m cancel_pledge -p root,worker,250,250
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
[GIN] 2018/10/30 - 15:53:27 | 200 |     146.209µs |       127.0.0.1 | GET      /wallet/getPublicKeys
[GIN] 2018/10/30 - 15:53:27 | 200 |     229.271µs |       127.0.0.1 | POST     /wallet/signTransaction
success
```
查看root的账户，delegate中给worker抵押的CPU和NET都减少了250个ABA
查看worker的账户，CPU和Net的delegated_aba都减少了250个，证明操作成功
```
{"index":13630584076887916544,"timestamp":1540885146345046273,"token":{"ABA":{"index":"ABA","balance":64500}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x0164885481cec31154be43cf86067eccd2dfc088":{"actor":[1,100,136,84,129,206,195,17,84,190,67,207,134,6,126,204,210,223,192,136],"weight":1}},"accounts":{}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x0164885481cec31154be43cf86067eccd2dfc088":{"actor":[1,100,136,84,129,206,195,17,84,190,67,207,134,6,126,204,210,223,192,136],"weight":1}},"accounts":{}}},"contract":{"typeVm":2,"describe":"c3lzdGVtIGNvbnRyYWN0","code":null,"abi":null},"delegate":[{"index":16514424251286028288,"cpu_aba":500,"net_aba":500},{"index":16514424251822899200,"cpu_aba":250,"net_aba":250}],"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":10000,"delegated_aba":0,"used_byte":-478.72833616348885,"available_byte":652466.4915861874,"limit_byte":651987.7632500239},"Cpu":{"staked_aba":10000,"delegated_aba":0,"used_ms":-1.799203398919427,"available_ms":124358.5960879402,"limit_ms":124356.79688454128},"Votes":{"staked_aba":21500,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}

```
```
ubuntu@ubuntu:~/go/src/github.com/ecoball/go-ecoball/build$ ./ecoclient query account -n worker
normal run
Using config file: /home/ubuntu/go/src/github.com/ecoball/go-ecoball/build/ecoball.toml
{"index":16514424251286028288,"timestamp":1540868531900863745,"token":{"ABA":{"index":"ABA","balance":500}},"permissions":{"active":{"perm_name":"active","parent":"owner","threshold":1,"keys":{"0x018802f313931f3710f9173fe1559f3f91c8d98d":{"actor":[1,136,2,243,19,147,31,55,16,249,23,63,225,85,159,63,145,200,217,141],"weight":1}},"accounts":{}},"owner":{"perm_name":"owner","parent":"","threshold":1,"keys":{"0x018802f313931f3710f9173fe1559f3f91c8d98d":{"actor":[1,136,2,243,19,147,31,55,16,249,23,63,225,85,159,63,145,200,217,141],"weight":1}},"accounts":{}}},"contract":{"typeVm":1,"describe":"dG9rZW4gY29udHJhY3Q=","code":"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","abi":"AAAFBmNyZWF0ZQADBmlzc3VlcgZzdHJpbmcKbWF4X3N1cHBseQVpbnQ2NAh0b2tlbl9pZAZzdHJpbmcFaXNzdWUAAwJ0bwZzdHJpbmcGYW1vdW50BWludDY0CHRva2VuX2lkBnN0cmluZwh0cmFuc2ZlcgAEBGZyb20Gc3RyaW5nAnRvBnN0cmluZwZhbW91bnQFaW50NjQIdG9rZW5faWQGc3RyaW5nB0FjY291bnQAAwdiYWxhbmNlBWludDY0BG5hbWUIY2hhclsxNl0IdG9rZW5faWQHY2hhcls4XQRTdGF0AAQGc3VwcGx5BWludDY0Cm1heF9zdXBwbHkFaW50NjQGaXNzdWVyCGNoYXJbMTZdCHRva2VuX2lkB2NoYXJbOF0DAAAAAKhs1EUGY3JlYXRlAAAAAAAApTF2BWlzc3VlAAAAAFctPM3NCHRyYW5zZmVyAAIAAAAgT00RAgNpNjQBCGN1cnJlbmN5AQZ1aW50NjQHQWNjb3VudAAAAAAAkE0GA2k2NAEIY3VycmVuY3kBBnVpbnQ2NARTdGF0AAAA"},"delegate":null,"resource":{"Ram":{"quota":0,"used":0},"Net":{"staked_aba":250,"delegated_aba":250,"used_byte":3244.1974695643567,"available_byte":29681.18457456185,"limit_byte":32925.38204412621},"Cpu":{"staked_aba":250,"delegated_aba":250,"used_ms":10.356275881883157,"available_ms":6269.661966787452,"limit_ms":6280.018242669335},"Votes":{"staked_aba":500,"producers":{}}},"hash":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}

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