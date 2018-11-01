#include <stdlib.h>
#include <db.h>
#include <print.h>
#include <runtime.h>
#include <system.h>
#include <types.h>
#include <string.h>
#include <malloc.h>
#include <module.h>

#define ACCOUNT_SIZE    12
#define TOKEN_ID_SIZE   12

struct Stat{
    char token_id[TOKEN_ID_SIZE];
    long long int  max_supply;
    long long int  supply;
    char issuer[ACCOUNT_SIZE];
};

struct Account{
    long long int balance;
};

// token must be all upper because it may be same with account name
int isTokenString(char *token_id) {
    int i = 0;
    for(; i < strlen(token_id); i++) {
        if(token_id[i] >= 'A' && token_id[i] <= 'Z') continue;
        else return -1;
    }

    return 0;
}

// create token
int create(char *issuer, long long int max_supply, char *token_id){
    struct Stat status, stat;
    // long long int balance = 0;

    // balance = ABA_get_token_balance("root", strlen("root"), "ABA", strlen("ABA"));
    // ABA_assert( balance != 67100, "get root's balance is wrong!" );

    ABA_assert( max_supply <= 0, "max_supply must be postive!" );
    ABA_assert( isTokenString(token_id),  "token id must be all upper character");
    ABA_assert( ABA_is_account(issuer, strlen(issuer)) != 0, "The issuer account does not exist" );
    ABA_assert(ABA_token_Existed(token_id, strlen(token_id)) == 1, "The token had existed");

    stat.supply = 0;
    stat.max_supply = max_supply;
    strcpy(stat.issuer, issuer);
    strcpy(stat.token_id, token_id);
    
    ABA_put_token_status(token_id, strlen(token_id), &stat, sizeof(stat));
    // ABA_token_info_put(token_id, strlen(token_id), max_supply, 0, issuer, strlen(issuer));
    // ABA_createToken(token_id, strlen(token_id), max_supply, issuer, strlen(issuer));

    // ABA_get_token_status(token_id, strlen(token_id), &status.max_supply, sizeof(status.max_supply), &status.supply, sizeof(status.supply), status.issuer, sizeof(status.issuer));
    // ABA_token_info_get(token_id, strlen(token_id), &status, sizeof(status));
    // ABA_prints(status.issuer);
    // ABA_assert( max_supply != status.max_supply, "get max_supply is wrong!" );
    // ABA_assert( status.supply != 0, "get supply is wrong!" );

    return 0;
}


// issue token
int issue(char *to, long long int amount, char *token_id){
    struct Stat stat;
    struct Account aIssuer, aTo;
    int result;
  
    // check amount, token_id
    ABA_assert( amount <= 0, "amount must be postive!" );
    ABA_assert( isTokenString(token_id),  "token id must be all upper character");
    // check if to account is existed
    ABA_assert( ABA_is_account(to, strlen(to)) != 0, "The receiving account does not exist" );

    // get issuer of token
    result = ABA_get_token_status(token_id, strlen(token_id), &stat, sizeof(stat));
    ABA_assert( result != 0, "The token does not exist" );

    // check if has issuer's permission
    ABA_require_auth(stat.issuer, strlen(stat.issuer));

    // if unsupplied token is greater than amount
    ABA_assert(stat.max_supply - stat.supply < amount, "The unsupplied token is not enough");

    // add the receiving account's balance
    result = ABA_add_token_balance(to, strlen(to), token_id, strlen(token_id), amount);
    ABA_assert( result != 0, "param is wrong, add balance failed" );

    // issue mean token supply addition 
    stat.supply += amount;

    // update database
    result = ABA_put_token_status(token_id, strlen(token_id), &stat, sizeof(stat));
    ABA_assert( result != 0, "The token does not exist" );

    return 0;
}

// transfer token
int transfer(char *from, char *to, long long int amount, char *token_id){

 struct Account aFrom, aTo;
    int result;
    // check amount and token_id
    ABA_assert( amount <= 0, "amount must be postive!" );
    ABA_assert(strcmp(from, to) == 0, "can not transfer to self");
    ABA_assert( isTokenString(token_id),  "token id must be all upper character");

    // check if from and to account exists
    ABA_assert( ABA_is_account(from,strlen(from)) != 0, "The transfer account does not exist" );
    ABA_assert( ABA_is_account(to,strlen(to)) != 0, "The receiving account does not exist" );

    // check if has from's permission
    ABA_require_auth(from, strlen(from));

    // get balance of the transfer account
    aFrom.balance = ABA_get_token_balance(from, strlen(from), token_id, strlen(token_id));
    ABA_assert( aFrom.balance < 0, "can not get the transfer account" );

    // if the balance of the transfer is greater than amount
    ABA_assert(aFrom.balance < amount, "The account balance is not enough");

    // sub from account balance, and add to account balance
    result = ABA_sub_token_balance(from, strlen(from), token_id, strlen(token_id), amount);
    ABA_assert( result != 0, "param is wrong, add balance failed" );
    result = ABA_add_token_balance(to, strlen(to), token_id, strlen(token_id), amount);
    ABA_assert( result != 0, "param is wrong, add balance failed" );

    // const char *strActionData = "[\"worker1\", \"worker2\", \"15\", \"XXX\"]";
    // ABA_inline_action("worker2", strlen("worker2"), "transfer", strlen("transfer"), strActionData, strlen(strActionData), "worker1", strlen("worker1"), "active", strlen("active"));

    result = ABA_add_token_balance(from, strlen(from), token_id, strlen(token_id), amount);
    ABA_assert( result != 0, "param is wrong, add balance failed" );
    result = ABA_sub_token_balance(to, strlen(to), token_id, strlen(token_id), amount);
    ABA_assert( result != 0, "param is wrong, add balance failed" );

    result = ABA_sub_token_balance(from, strlen(from), token_id, strlen(token_id), amount);
    ABA_assert( result != 0, "param is wrong, add balance failed" );
    result = ABA_add_token_balance(to, strlen(to), token_id, strlen(token_id), amount);
    ABA_assert( result != 0, "param is wrong, add balance failed" );

    return 0;
}


// contract invoke entry
export int apply(char *method){
    if(strcmp(method,"create") == 0){
        char *token_id,*token_issuer;
        long long int max_supply;
        // get function parameters
        token_issuer = ABA_read_param(1);
        max_supply = ABA_read_param(2);
        token_id = ABA_read_param(3);
        // call create function
        create(token_issuer, max_supply, token_id);
    }
    if(strcmp(method,"issue") == 0){
        char *to,*token_id;
        long long int amount;
        // get function parameters
        to = ABA_read_param(1);
        amount = ABA_read_param(2);
        token_id = ABA_read_param(3);
        // call issue function
        issue(to, amount, token_id);
    }
    if(strcmp(method,"transfer") == 0){
        char *from,*to,*token_id;
        long long int amount;
        // get function parameters
        from = ABA_read_param(1);
        to = ABA_read_param(2);
        amount = ABA_read_param(3);
        token_id = ABA_read_param(4);
        // call transfer function
        transfer(from, to, amount, token_id);
    }
    return 0;
}