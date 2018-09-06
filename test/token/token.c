void create(char *issue,int max_supply,char *token_id){
    char *supply_str,*max_str;
    char *str = "supply";
    ABA_prints("create func");

    supply_str = ABA_itoa(0);
    max_str = ABA_itoa(max_supply);
    //ABA_prints("db op");    
    ABA_db_put(str,supply_str);
    ABA_db_put(token_id,max_str);
    
    return ;
}

void issue(char *to,int amount,char *token_id){
    int balance = 0,supply;
    char *str = "supply";
    char *supply_str,*balance_str;

    ABA_prints("issue func");

    //ABA_prints("issue db op");

    balance = balance + amount;
    //ABA_prints(balance);
    balance_str = ABA_itoa(balance);
    //ABA_prints(balance_str);
    
    supply_str = ABA_db_get(str);
    supply = ABA_atoi(supply_str);
    supply = supply + amount;
    //ABA_printui(supply);
    supply_str = ABA_itoa(supply);
    ABA_prints(supply_str);

    ABA_db_put(to,balance_str);
    ABA_db_put(str,supply_str);

    return ;
}

void transfer(char *from,char *to,int amount,char *token_id){
    char *balance1_str,*balance2_str;
    int balance1,balance2;

    ABA_prints("transfer func");

    //ABA_prints("transfer db op");
    balance1_str = ABA_db_get(from);
    balance2_str = ABA_db_get(to);
    balance1 = ABA_atoi(balance1_str);
    balance2 = ABA_atoi(balance2_str);
    
    balance1 = balance1 - amount;
    balance2 = balance2 + amount;
    balance1_str = ABA_itoa(balance1);
    balance2_str = ABA_itoa(balance2);

    ABA_db_put(from,balance1_str);
    ABA_db_put(to,balance2_str);

    return ;
}


int apply(char *method){
    //create(char issue_name[],int max_supply,char token_id[])
    if(ABA_strcmp(method,"create") == 0){
        char *token_id,*token_issue;
        int max_supply;
        ABA_prints("create start");

        //ABA_prints("create get param");
        token_issue = ABA_read_param(1);
        max_supply = ABA_read_param(2);
        //ABA_printui(ABA_len(token_issue));
        token_id = ABA_read_param(3);
        //ABA_printui(ABA_len(token_id));
        
        //ABA_prints("create call");
        create(token_issue,max_supply,token_id);
    }
    //issue(char to[],int amount,char token_id[])
    if(ABA_strcmp(method,"issue") == 0){
        char *to,*token_id;
        int amount;
        ABA_prints("issue test");

        //ABA_prints("issue get param");
        to = ABA_read_param(1);
        //ABA_prints(to);
       // ABA_printui(ABA_len(to));
        amount = ABA_read_param(2);
        //ABA_printui(amount);
        token_id = ABA_read_param(3);
        //ABA_prints(token_id);
        //ABA_printui(ABA_len(token_id));

        //ABA_prints("issue call");
        issue(to,amount,token_id);
    }
    //transfer(char from[],char to[],int amount,char token_id[])
    if(ABA_strcmp(method,"transfer") == 0){
        char *from,*to,*token_id;
        int amount;
        ABA_prints("transfer test");

        ABA_prints("transfer get param");
        from = ABA_read_param(1);
        to = ABA_read_param(2);
        amount = ABA_read_param(3);
        token_id = ABA_read_param(4);

        ABA_prints("transfer call");
        transfer(from,to,amount,token_id);
    }
    return 0;
}