#include <module.h>
#include <print.h>
#include <store.h>

//create token
int create(char *creator,int max_supply,char *token_id){
    char *supply_str,*max_supply_str;
    char *str_supply = "supply";
	char *str_max_supply = "max_supply";
	char *str_token = "token";
    int result;

    //if the creator account exists
    result = ABA_account_contain(creator);
    if(result != 1) {
      ABA_prints("The creator account does not exist");
      return -1;
    }
	
	result = ABA_db_get(str_token);
    if(result != -1) {
      ABA_prints("Token had been created! Only support one token!");
      return -2;
    }
	
    //if the token has created
	/*
    max_supply_str = ABA_db_get(token_id);
    result = (int)max_supply_str;
    if((int)max_supply_str != -1){
        ABA_prints("token has been created");
        return;
    }
	*/
    supply_str = ABA_itoa(0);
    max_supply_str = ABA_itoa(max_supply);
	
    //update database
	ABA_db_put(str_token,token_id);
    ABA_db_put(str_max_supply,max_supply_str);
	ABA_db_put(str_supply,supply_str);
	
    return 1;
}
//issue token
void issue(char *to,int amount,char *token_id){
    int result,balance,supply,max_supply;
    char *str = "supply";
	char *str_max_supply = "max_supply";
	char *str_token = "token";
	char *str_issuer = "issuer";
    char *supply_str,*balance_str,*max_supply_str;

    //if the token exsit
	result = ABA_db_get(str_token);
    if(result == -1) {
      ABA_prints("token don't exsit!");
      return;
    }
	
	max_supply_str = ABA_db_get(str_max_supply);
    if((int)max_supply_str == -1){
        ABA_prints("can not find max supply");
        return;
    }	
	
    //if the receiving account exists
    result = ABA_account_contain(to);
    if(result != 1) {
      ABA_prints("The receiving account does not exist");
      return;
    }
    //get the receiver balance
    balance_str = ABA_db_get(to);
    if((int)balance_str == -1){
        balance = amount;
        balance_str = ABA_itoa(balance);
    } else {
      balance = ABA_atoi(balance_str);
      balance = balance + amount;
      balance_str = ABA_itoa(balance); 
    }
    
    //update token circulation
    supply_str = ABA_db_get(str);
    //max_supply_str = ABA_db_get(str_max_supply);
    supply = ABA_atoi(supply_str);
    max_supply = ABA_atoi(max_supply_str);
    if(supply + amount > max_supply) {
        ABA_prints("token issued too much");
        return;
    }
    supply = supply + amount;
    supply_str = ABA_itoa(supply);

    //update database
	ABA_db_put(str_issuer, to);
    ABA_db_put(to,balance_str);
    ABA_db_put(str,supply_str);
	
	inline_action("worker2", "transfer", "[\"worker1\", \"worker\", \"15\", \"xxx\"]", "worker1", "active");
	
    return ;
}
// transfer token
void transfer(char *from,char *to,int amount,char *token_id){
    char *balance1_str,*balance2_str;
    int balance1,balance2,result;
  
    //if the creator account exists
    result = ABA_account_contain(from);
    if(result != 1) {
      ABA_prints("The transfer account does not exist");
      return;
    }
    result = ABA_account_contain(to);
    if(result != 1) {
      ABA_prints("The receiving account does not exist");
      return;
    }

    require_auth(from);

    //get the account balance
    balance1_str = ABA_db_get(from);
    if((int)balance1_str == -1){
        ABA_prints("The transfer account does not have token");
        return;
    }
    balance2_str = ABA_db_get(to);
    if((int)balance2_str == -1){
        balance1 = ABA_atoi(balance1_str);
        balance2 = 0;
    } else {
        balance1 = ABA_atoi(balance1_str);
        balance2 = ABA_atoi(balance2_str);
    }

    // if the balance of the transfer is greater than amount
    if(balance1 < amount){
        ABA_prints("The account balance is not enough");
        return;
    }
    balance1 = balance1 - amount;
    balance2 = balance2 + amount;
    
    balance1_str = ABA_itoa(balance1);
    balance2_str = ABA_itoa(balance2);

    //update database
    ABA_db_put(from,balance1_str);
    ABA_db_put(to,balance2_str);

    return ;
}

// contract invoke entry
export int apply(char *method){
    //create(char issue_name[],int max_supply,char token_id[])
    if(ABA_strcmp(method,"create") == 0){
        char *token_id,*token_issue;
        int max_supply;
        //get function parameters
        token_issue = ABA_read_param(1);
        max_supply = ABA_read_param(2);
        token_id = ABA_read_param(3);
        //call create function
        create(token_issue,max_supply,token_id);
    }
    //issue(char to[],int amount,char token_id[])
    if(ABA_strcmp(method,"issue") == 0){
        char *to,*token_id;
        int amount;
        //get function parameters
        to = ABA_read_param(1);
        amount = ABA_read_param(2);
        token_id = ABA_read_param(3);
        //call issue function
        issue(to,amount,token_id);
    }
    //transfer(char from[],char to[],int amount,char token_id[])
    if(ABA_strcmp(method,"transfer") == 0){
        char *from,*to,*token_id;
        int amount;
        //get function parameters
        from = ABA_read_param(1);
        to = ABA_read_param(2);
        amount = ABA_read_param(3);
        token_id = ABA_read_param(4);
        //call transfer function
        transfer(from,to,amount,token_id);
    }
    return 0;
}
