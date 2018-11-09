#!/usr/bin/env python3

import subprocess
import sys
import argparse
import json
import os
import pytoml

# Sharding scheme: initial startup of 5 committee, 3 Shared, each Shared 5 nodes.
# Buy five servers, one server one committee docker instance and three share docker instance


def run(shell_command):
    '''
    Execute shell command.
    If it fails, exit the program with an exit code of 1.
    '''

    print('start.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('start.py: exiting because of error')
        sys.exit(1)


# Command Line Arguments
parser = argparse.ArgumentParser()
parser.add_argument('-i', '--node-ip', metavar='', required=True, help="IP address of node", nargs='+', dest="node_ip")
parser.add_argument('-o', '--host-ip', metavar='', required=True, help="IP address of host node", dest="host_ip")
parser.add_argument('-n', '--number', type=int, required=True, metavar='', help="The index number of container instance", dest="number")
parser.add_argument('-w', '--weight', type=int, default=1, metavar='', help="The number of weights", dest="weight")
parser.add_argument('--http-port', type=int, default=20678, metavar='', help="client http port", dest="http_port")
parser.add_argument('--wallet-http-port', type=int, default=20679, metavar='', help="client wallet http port", dest="wallet_http_port")
parser.add_argument('--version', type=float, default=1.0, metavar='', help="system version")
parser.add_argument('--log-dir', default="/tmp/Log/", metavar='', help="log file location", dest="log_dir")
parser.add_argument('--output-to-terminal', type=bool, default=True, metavar='', help=" debug output type", dest="output_to_terminal")
parser.add_argument('--log-level', type=int, default=1, metavar='', help="debug level", dest="log_level")
parser.add_argument('--consensus-algorithm', default="SHARD", metavar='', help="can set as SOLO, DPOS, ABABFT, SHARD", dest="consensus_algorithm")
parser.add_argument('--time-slot', type=int, default=500, metavar='', help="block interval time, uint ms", dest="time_slot")

#parse Arguments
args = parser.parse_args()

#Input parameter judgment
if args.node_ip is None or args.host_ip is None or args.number is None or args.weight is None:
    print('please input iP address of node and host node and the index number of container instance and weight number. -h shows options.')
    sys.exit(1)

if args.number < 0 or args.number > 4 * args.weight - 1:
    print('The index value must be between 0 and %d' %(4 * args.weight -1))
    sys.exit(1)

#Generate the configuration json files required for sharding
Pubkey = "1109ef616830cd7b8599ae7958fbee56d4c8168ffd5421a16025a398b8a4be"
start_pubkey = 40
start_port = 2000
committee = []
shard = []

ip_index = 0
for ip in args.node_ip:
    port_index = 0
    while port_index < 4 * args.weight:
        node = {
            "Pubkey": Pubkey + str(start_pubkey + 4 * args.weight * ip_index + port_index), 
            "Address": ip, 
            "Port": str(start_port + port_index)
        }
        port_index += 1
        if port_index <= args.weight:
            committee.append(node)
        else:
            shard.append(node)
    ip_index += 1


ip_index = args.node_ip.index(args.host_ip)


data = {
    "Pubkey": Pubkey + str(start_pubkey + 4 * args.weight * ip_index + args.number),
    "Address": args.host_ip,
    "Port": str(start_port + args.number),
    "Committee": committee,
    "Shard": shard
}

root_dir = os.path.split(os.path.realpath(__file__))[0]

with open(os.path.join(root_dir, '../build/sharding.json'), 'w') as f:
    json.dump(data, f)

#Generate the configuration toml files required for ecoball

config = {
    "http_port": str(args.http_port),
    "wallet_http_port": str(args.wallet_http_port),
    "version": str(args.version),
    "log_dir": args.log_dir,
    "output_to_terminal": args.output_to_terminal,
    "log_level": args.log_level,
    "consensus_algorithm": args.consensus_algorithm,
    "time_slot": args.time_slot,
    "start_node": True,
    "root_privkey": "34a44d65ec3f517d6e7550ccb17839d391b69805ddd955e8442c32d38013c54e",
    "root_pubkey": "04de18b1a406bfe6fb95ef37f21c875ffc9f6f59e71fea8efad482b82746da148e0f154d708001810b52fb1762d737fec40508b492628f86c605391a891a61ad0b",
    "aba_token_privkey": "675e6cbc4190bc861a987eec5be717ebdd6ead16cb5f537df00637080f000917",
    "aba_token_pubkey": "040eb444f2962e94722f84d3298b062051b7d488d14c0a8216f730e1f36177fa1e73fdcb16582aaa62efa7a0fa1737f282a276081252cb41429597c8c9159d43ee",
    "dsn_privkey" = "23c45ee08031cfa233a0a1a42df1cd66f73b74ff68645b223e629ac1e0db1374"
    "dsn_pubkey" = "04f7d6be089dc5cfb263c708e123111510ea9d3e29cf8a9b2b3eef35838d6e8d55de92c303024d30b748ec450edc8228ac6d3c6431cacc67b32f9905bb3363cd00"
    "user_privkey": "34a44d65ec3f517d6e7550ccb17839d391b69805ddd955e8442c32d38013c54e",
    "user_pubkey": "04de18b1a406bfe6fb95ef37f21c875ffc9f6f59e71fea8efad482b82746da148e0f154d708001810b52fb1762d737fec40508b492628f86c605391a891a61ad0b",
    "worker1_privkey": "cb0324ee8f7bd11dec57e39c4f560b9343c6c81c71012b96be29f26b92fef6f9",
    "worker1_pubkey": "0425adbea1ddc21124279059057b4c9b5df4d40e49f2625504b45e0d43aea22c25621e42307eb8224f7ea0e65d40c0495d3cbd3f020f801f38b73cec5740bf1ec9",
    "worker2_privkey": "05cac9544f828b570724eb52b5903a68fbe0c8f23a15851cb717a5f7eda801cd",
    "worker2_pubkey": "040cf9d46f4f5945ed7986cb8920feb5ac4eb06bb26cb048ed9dc2de4c54c19914bf4adf5ca0571a6f106bf4542fc7bfcfd164d8065598fc76042c074b24048960",
    "worker3_privkey": "79b99bbd11bd14e8c0da65c20bae059d1eee06f92380fb88ff31a88c84d3fc6e",
    "worker3_pubkey": "04717944fa32da2261eeda1e810c3b3c62216ed486785a9aa78e2cde0f18805882631033aed956d02721e9fae079e600bd512d4feb0375a14d882a63e48971d413",
    "delegate_privkey": "56bd8432606e6e2eb354794d89059f7f9e9a0de62166145b898136b496be6aed",
    "delegate_pubkey": "04070a106e034b11e03bab17aab0d2e75d7795bae8346f6f527f436cd714f7798efdeced276f326ed3406e3baab257487330e61896c838920328a4d745a87f06d1",
    "worker_privkey": "8bbd547fe9d9e867721c6fa643fbe637fc3d955e588358a45c11d63dd5a25016",
    "worker_pubkey": "041a0a2b0bfce1d624c125d2a9fcca16c5b2b96bc78ab827e1c23818df4a70a4441c9665850268d48ab23e102cf1dc6864596a19e748c0867dce400a3f219e3f62",
    "peer_list": [ "120202c924ed1a67fd1719020ce599d723d09d48362376836e04b0be72dfe825e24d810000", "120202935fb8d28b70706de6014a937402a30ae74a56987ed951abbe1ac9eeda56f0160000" ],
    "peer_index": [ "1", "2" ],
    "p2p_listen_address": ["/ip4/0.0.0.0/tcp/4013","/ip6/::/tcp/4013"],
    "announce_address": [],
    "no_announce_address": [],
    "bootstrap_address": [],
    "disable_nat_port_map": False,
    "disable_relay": False,
    "enable_relay_hop": False,
    "conn_mgr_lowwater": 600,
    "conn_mgr_highwater": 900,
    "conn_mgr_graceperiod": 20,
    "enable_local_discovery": False,
    "disable_localdis_log": True,
    "dsn_storage": False,
    "disable_sharding": False
}

with open(os.path.join(root_dir, '../build/ecoball.toml'), 'w') as f:
    pytoml.dump(config, f)

#start ecoball
run("cd " + os.path.join(root_dir, '../build/') + "&& ./ecoball run")

