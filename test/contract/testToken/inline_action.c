#include <stdlib.h>
#include <db.h>
#include <print.h>
#include <runtime.h>
#include <system.h>
#include <types.h>
#include <string.h>
#include <malloc.h>
#include <module.h>

#define ACCOUNT_SIZE    16
#define TOKEN_ID_SIZE   8

struct Stat{
    int  supply;
    int  max_supply;
    char issuer[ACCOUNT_SIZE];
    char token_id[TOKEN_ID_SIZE];
};

struct Account{
    int balance;
    char name[ACCOUNT_SIZE];
    char token_id[TOKEN_ID_SIZE];
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
int create(char *issuer, int max_supply, char *token_id){

    ABA_assert( max_supply < 0, "max_supply can not be negative!" );
    ABA_assert( isTokenString(token_id),  "token id must be all upper character");

    int result;
    struct Account aIssuer;
    struct Stat stat;

    // if token exist
    result = ABA_db_get(token_id, strlen(token_id), &stat, sizeof(stat));
    ABA_assert( strcmp(token_id, stat.token_id) == 0, "The token had existed" );

    // if the creator account exists
    ABA_assert( ABA_is_account(issuer, strlen(issuer)) != 0, "The issuer account does not exist" );

    stat.supply = 0;
    stat.max_supply = max_supply;
    strcpy(stat.issuer, issuer);
    strcpy(stat.token_id, token_id);

    aIssuer.balance = max_supply;
    strcpy(aIssuer.name, issuer);
    strcpy(aIssuer.token_id, token_id);

    // update database
    ABA_db_put(token_id, strlen(token_id), &stat, sizeof(stat));
    ABA_db_put(issuer, strlen(issuer), &aIssuer, sizeof(aIssuer));

    return 0;
}


// issue token
int issue(char *to, int amount, char *token_id){
    struct Stat stat;
    struct Account aIssuer, aTo;
    int result;
  
    ABA_assert( amount < 0, "amount can not be negative!" );
    ABA_assert( isTokenString(token_id),  "token id must be all upper character");

    // if the creator account exists
    ABA_assert( ABA_is_account(to, strlen(to)) != 0, "The receiving account does not exist" );

    // get issuer of token
    result = ABA_db_get(token_id, strlen(token_id), &stat, sizeof(stat));
    ABA_assert( result != 0, "The token does not exist" );

    // can not issue to issuer
    ABA_assert(strcmp(stat.issuer, to) == 0, "can not transfer to self");

    require_auth(stat.issuer, strlen(stat.issuer));

    // get balance of the transfer account
    result = ABA_db_get(stat.issuer, strlen(stat.issuer), &aIssuer, sizeof(aIssuer));
    ABA_assert( result != 0, "The transfer account does not exist" );

    // if the balance of the transfer is greater than amount
    ABA_assert(aIssuer.balance < amount, "The issuer account balance is not enough");

    // get balance of the receiving account
    result = ABA_db_get(to, strlen(to), &aTo, sizeof(aTo));
    ABA_assert( result != 0, "The receiving account does not exist" );

    // sub from account balance, and add to account balance
    aIssuer.balance = aIssuer.balance - amount;
    aTo.balance = aTo.balance + amount;
    stat.supply += amount;

    // update database
    ABA_db_put(token_id, strlen(token_id), &stat, sizeof(stat));
    ABA_db_put(stat.issuer, strlen(stat.issuer), &aIssuer, sizeof(aIssuer));
    ABA_db_put(to, strlen(to), &aTo, sizeof(aTo));

    // inline_action("worker2", "transfer", "[\"worker1\", \"worker2\", \"15\", \"XXX\"]", "worker1", "active");

    return 0;
}

// transfer token
int transfer(char *from, char *to, int amount, char *token_id){
    struct Account aFrom, aTo;
    int result;
  
    ABA_assert( amount < 0, "amount can not be negative!" );
    ABA_assert(strcmp(from, to) == 0, "can not transfer to self");
    ABA_assert( isTokenString(token_id),  "token id must be all upper character");

    // if the creator account exists
    ABA_assert( ABA_is_account(from,strlen(from)) != 0, "The transfer account does not exist" );
    ABA_assert( ABA_is_account(to,strlen(to)) != 0, "The receiving account does not exist" );

    require_auth(from, strlen(from));

    // get balance of the transfer account
    result = ABA_db_get(from, strlen(from), &aFrom, sizeof(aFrom));
    ABA_assert( result != 0, "The transfer account does not exist" );

    // if the balance of the transfer is greater than amount
    ABA_assert(aFrom.balance < amount, "The account balance is not enough");

    // get balance of the receiving account
    result = ABA_db_get(to, strlen(to), &aTo, sizeof(aTo));
    ABA_assert( result != 0, "The receiving account does not exist" );

    // sub from account balance, and add to account balance
    aFrom.balance = aFrom.balance - amount;
    aTo.balance = aTo.balance + amount;

    // update database
    ABA_db_put(from, strlen(from), &aFrom, sizeof(aFrom));
    ABA_db_put(to, strlen(to), &aTo, sizeof(aTo));

    // ABA_transfer(from, to, amount, "active");

    return 0;
}


// contract invoke entry
export int apply(char *method){
    if(strcmp(method,"create") == 0){
        char *token_id,*token_issuer;
        int max_supply;
        // get function parameters
        token_issuer = ABA_read_param(1);
        max_supply = ABA_read_param(2);
        token_id = ABA_read_param(3);
        // call create function
        create(token_issuer, max_supply, token_id);
    }
    if(strcmp(method,"issue") == 0){
        char *to,*token_id;
        int amount;
        // get function parameters
        to = ABA_read_param(1);
        amount = ABA_read_param(2);
        token_id = ABA_read_param(3);
        // call issue function
        issue(to, amount, token_id);
    }
    if(strcmp(method,"transfer") == 0){
        char *from,*to,*token_id;
        int amount;
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