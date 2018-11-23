#pragma once 

int ABA_token_Existed(int name, int nameLen);
int ABA_put_token_info(char *name, int nameLen, int maxSupply, char *issuer, int issuerLen);
int ABA_get_token_info(int name, int nameLen, int maxSupply, int maxSupplyLen, int supply, int supplyLen, int issuer, int issuerLen);
int ABA_add_token_balance(int account, int accountLen, int name, int nameLen, int64_t amount);
int ABA_sub_token_balance(int account, int accountLen, int name, int nameLen, int64_t amount);
int64_t  ABA_get_token_balance(int account, int accountLen, int name, int nameLen);
int ABA_get_token_status(int name, int nameLen, int token, int tokenLen);
int ABA_put_token_status(int name, int nameLen, int token, int tokenLen);
