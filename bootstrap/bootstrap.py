#!/usr/bin/env python3

import argparse
import subprocess
import sys
import time
import os
import json

#step 0: Kill all ecowallet ecoball processes
#step 1: Start the wallet, import the private key of root, import the private key of the producers and users
#step 2: Create system accounts
#step 3: Deploy the system contracts

systemAccounts = []
ecoballLogFile = 'ecoball.log'
ecowalletLogFile = 'ecowallet.log'
#execute shell command
def run(shell_command):
    print('bootstrap.py:', shell_command)
    if subprocess.call(shell_command, shell=True):
        print('bootstrap.py: exiting because of error')
        sys.exit(1)

def background(shell_command):
    print('bootstrap.py:', shell_command)
    return subprocess.Popen(shell_command, shell=True)

def sleep(t):
    print('sleep', t, '...')
    time.sleep(t)
    print('resume')

#step kill all
def stepKillAll():
    run("killall ecoball ecowallet || true")
    sleep(1.5)

#step start ecowallet
def runWallet():
    run('rm -rf ' + os.path.abspath(args.wallet_dir))
    background(args.ecowallet + ' > ' + ecowalletLogFile)
    sleep(0.4)
    run(args.ecoclient + ' wallet create --name=default --password=default')

def importKeys():
    run(args.ecoclient + ' wallet import --name=default --private=' + args.private_key)
    keys = {}
    for a in accounts:
        key = a['private']
        if not key in keys:
            if len(keys) >= args.max_user_keys:
                break
            keys[key] = True
            run(args.ecoclient + ' wallet import --name=default --private=' + key)
    for i in range(firstProducer, firstProducer + numProducers):
        a = accounts[i]
        key = a['private']
        if not key in keys:
            keys[key] = True
            run(args.ecoclient + ' wallet import --name=default --private=' + key)

def stepStartEcowallet():
    runWallet()
    importKeys()

#step start ecoball
def runNode():
    background(args.ecoball + " run" + ' > ' + ecoballLogFile)

def stepStartEcoball():
    runNode()
    sleep(1.5)

#step create system account
def stepCreateSystemAccounts():
    for a in systemAccounts:
        run(args.ecoclient + 'create account --creator=root --name=' + a + ' --active=' + args.public_key)

#step install system contract
def stepInstallSystemContracts():
    print("install system contract")

#commands
commands = [
    ('k', 'kill',           stepKillAll,                True,    "Kill all ecoball and ecowallet processes"),
    ('w', 'wallet',         stepStartEcowallet,         True,    "Start ecowallet, create wallet, fill with keys"),
    ('b', 'node',           stepStartEcoball,           True,    "Start ecoball node"),
    ('s', 'account',        stepCreateSystemAccounts,   True,    "Create system accounts"),
    ('c', 'contracts',      stepInstallSystemContracts, True,    "Install system contracts"),
]

# Command Line Arguments
parser = argparse.ArgumentParser()
parser.add_argument('--public-key', metavar='', help="root Public Key", default='0x0463613734b23e5dd247b7147b63369bf8f5332f894e600f7357f3cfd56886f75544fd095eb94dac8401e4986de5ea620f5a774feb71243e95b4dd6b83ca49910c', dest="public_key")
parser.add_argument('--private-Key', metavar='', help="root Private Key", default='0x33a0330cd18912c215c9b1125fab59e9a5ebfb62f0223bbea0c6c5f95e30b1c6', dest="private_key")
parser.add_argument('--wallet-dir', metavar='', help="Path to wallet directory", default='../build/wallet/', dest="wallet_dir")
parser.add_argument('--ecowallet', metavar='', help="Path to ecowallet binary", default='../build/ecowallet')
parser.add_argument('--ecoclient', metavar='', help="ecoclient command", default='../build/ecoclient')
parser.add_argument('--user-limit', metavar='', help="Max number of users. (0 = no limit)", type=int, default=3000, dest='user_limit')
parser.add_argument('--producer-limit', metavar='', help="Maximum number of producers. (0 = no limit)", type=int, default=0, dest='producer_limit')
parser.add_argument('--max-user-keys', metavar='', help="Maximum user keys to import into wallet", type=int, default=10, dest='max_user_keys')
parser.add_argument('--ecoball', metavar='', help="Path to ecoball binary", default='../build/ecoball')
parser.add_argument('-a', '--all', action='store_true', help="Do everything marked with (*)")

for (flag, command, function, inAll, help) in commands:
    prefix = ''
    if inAll: prefix += '*'
    if prefix: help = '(' + prefix + ') ' + help
    if flag:
        parser.add_argument('-' + flag, '--' + command, action='store_true', help=help, dest=command)
    else:
        parser.add_argument('--' + command, action='store_true', help=help, dest=command)

#parse Arguments
args = parser.parse_args()

#load account
with open('accounts.json') as f:
    a = json.load(f)
    if args.user_limit:
        del a['users'][args.user_limit:]
    if args.producer_limit:
        del a['producers'][args.producer_limit:]
    firstProducer = len(a['users'])
    numProducers = len(a['producers'])
    accounts = a['users'] + a['producers']

#run commands
haveCommand = False
for (flag, command, function, inAll, help) in commands:
    if getattr(args, command) or inAll and args.all:
        if function:
            haveCommand = True
            function()
#error           
if not haveCommand:
    print('bootstrap.py: Tell me what to do. -a does almost everything. -h shows options.')
