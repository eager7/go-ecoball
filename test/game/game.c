#include <stdlib.h>
#include <stdio.h>
#include "string.h"
#include <module.h>
#include <malloc.h>
#include <db.h>
#include <print.h>
#include <runtime.h>
#include <system.h>
#include <types.h>
#include <action.h>

const int board_length = 3;

typedef struct{
    int  board[9];
    //participants
    char player1[10];
    char player2[10];
    //the player down on the chessboard this round
    char host[10];
    //the winner of the game, contract will check it when the follow action is asked
    char winner[10];
}game;

//function declaration
int initialize_array(int* arr, int length);
int regame(game* g, char* restarter);
int is_empty_cell(int cell);
int is_valid_movement(int row, int column, int* board);
int get_winner(game* g,char* winner);
int transfertoken(char* account1, char* account2, char* amount);
int is_valid_name(char* account_name);
int create(char* player1, char* player2);
int restart(char* player1, char* player2, char* restarter);
int close(char* player1, char* player2, char* closer);
int follow(char* player1, char* player2, char* host, int row, int column);

//initialize an array
int initialize_array(int* arr, int length){
    for(int i = 0; i < length; i++){
        arr[i] = 0;
    }

	return 0;
}

//reset the game parameters
int regame(game* g, char* restarter){
    initialize_array(g->board, board_length);
	for(int i = 0; i < 10; i++){
        g->host[i] = 0;
		g->winner[i] = 0;
    }
    strcpy(g->host, restarter);
    
	return 0;
}

//judge cell == 0?
int is_empty_cell(int cell){
    if (cell == 0) {
        return 1;
    } else {
        return -1;
    }
}

//judge the movement is valid in the board
int is_valid_movement(int row, int column, int* board){
	if (row < 4 && row > 0 && column < 4 && column > 0) {
		int location = (row - 1) * board_length + column - 1;
		//location < arraylen
    	int is_valid = (location < board_length * board_length) && is_empty_cell(board[location]);
    	if (is_valid) {
        	return 0;
    	}
	}
    return -1;
} 

int get_winner(game* g,char* winner){
    int rowi, columni;
    int vector[8];
    int row[3], column[3];
    int slash = 3, backslash = 3;
	int is_board_full = 0;

    initialize_array(row, board_length);
    initialize_array(column, board_length);

    for(int i = 0; i < board_length * board_length; i++) {
    	//compute row and column
        rowi = i / board_length;
        columni = i % board_length;

        row[rowi] = row[rowi] & g->board[i];
        column[columni] = column[columni] & g->board[i];

        //compute slash value
        if (rowi == columni){
            slash = slash & g->board[i];
        }
        //compute backslash value
        if (rowi + columni == 2) {
            backslash = backslash & g->board[i];
        }
        //compute the number of squares occupied
		if (g->board[i] != 0) {
			is_board_full++;
		}
    }
    for (int i = 0; i < board_length; i++) {
        vector[i] = row[i];
        vector[i + board_length] = column[i];
    }
    vector[6] = slash;
    vector[7] = backslash;

    //if vector[i] == 1, winner is player1; if vector[i] == 2, winner is player2; else winner is inconclusive
    for (int i = 0; i < 8; i++) {
        if(vector[i] == 1) {
			strcpy(winner, g->player1);
            return is_board_full;
        }
        if(vector[i] == 2) {
			strcpy(winner, g->player2);
            return is_board_full;
        }
    }
    return is_board_full;

}

//account1 transfer ABA to account2
int transfertoken(char* account1, char* account2, char* amount){
    char actionData[50];
	//joint actionData
	strcpy(actionData, "[\"");
    strcat(actionData, account1);
    strcat(actionData, "\",\"");
    strcat(actionData, account2);
    strcat(actionData, "\",\"");
    strcat(actionData, amount);
	strcat(actionData, "\",\"ABA\"]");
    ABA_prints(actionData);  
    //call inline_action API
    ABA_inline_action("abatoken", strlen("abatoken"), "transfer", strlen("transfer"), actionData, strlen(actionData), account1, strlen(account1), "active", strlen("active"));
	return 0;
}

//account name length must be greater than 0 and less than or equal to 10
int is_valid_name(char* account_name){
	int len;
	len = strlen(account_name);
	if (len <= 10){
		return 0;
	}
	return -1;
}

//create game : player1 and player2 is the participants and player1 is host
int create(char* player1, char* player2){
    game g;
	char gamekey[20];
    
    //if account name valid 
    ABA_assert(is_valid_name(player1) != 0, "player1 name is too long or is null");
    ABA_assert(is_valid_name(player2) != 0, "player2 name is too long or is null");
    //if player1 、 player2 account exsit
    ABA_assert(ABA_is_account(player1,strlen(player1)) != 0,"player1 is not a account");
    ABA_assert(ABA_is_account(player2,strlen(player2)) != 0,"player2 is not a account");
	//player1 and player1 should be different account
    ABA_assert(strcmp(player1,player2) == 0, "player1 shouldn't be the same as player2");
    
    //initialize game struct
    initialize_array(g.board, 9);
    strcpy(g.player1, player1);
    strcpy(g.player2, player2);
    strcpy(g.host, player1);
    for(int i = 0; i < 10; i++){
		g.winner[i] = 0;
    }

    //store game information
	strcpy(gamekey,player1);
	strcat(gamekey,player2);
    ABA_db_put(gamekey, strlen(gamekey), &g, sizeof(game));

    //player1 and player2 transfer ABA to game contract
	transfertoken(player1, "tictactoe", "2");
	transfertoken(player2, "tictactoe", "2");

    return 0;
}
//restart game : restarter restart the game and will be the host
int restart(char* player1, char* player2, char* restarter){
    int result, len;
    game g;
	char gamekey[20];
	
	//if account name valid 
	ABA_assert(is_valid_name(player1) != 0, "player1 name is too long or is null");
	ABA_assert(is_valid_name(player2) != 0, "player2 name is too long or is null");
	ABA_assert(is_valid_name(restarter) != 0, "restarter name is too long or is null");

	result = (strcmp(player1, restarter) == 0) | (strcmp(player2, restarter) == 0);
    ABA_assert(result != 1, "restarter has insufficient permission");

	strcpy(gamekey,player1);
	strcat(gamekey,player2);
	//if gmae exsit
    result = ABA_db_get(gamekey, strlen(gamekey), &g, sizeof(game));

    ABA_assert(result != 0, "the game does not exsit");

    //reset game and update database
    regame(&g, restarter);
    ABA_db_put(gamekey, strlen(gamekey), &g, sizeof(game));

    //the restarter must transfer 1 ABA to another player to restart game
    if (strcmp(restarter, player1) == 0){
    	transfertoken(player1, "tictactoe", "4");
    } else {
    	transfertoken(player2, "tictactoe", "4");
    }

    return 0;
}
//close game : initialize board and set host and winner null
int close(char* player1, char* player2, char* closer){
    int result;
    game g;
	char gamekey[20];
	
	ABA_assert(is_valid_name(player1) != 0, "player1 name is too long or is null");
	ABA_assert(is_valid_name(player2) != 0, "player2 name is too long or is null");
	ABA_assert(is_valid_name(closer) != 0, "closer name is too long or is null");

	result = (strcmp(player1, closer) == 0) | (strcmp(player2, closer) == 0);
    ABA_assert(result != 1, "closer has insufficient permission");

	strcpy(gamekey, player1);
	strcat(gamekey, player2);

    result = ABA_db_get(gamekey, strlen(gamekey), &g, sizeof(game));

    ABA_assert(result != 0, "the game does not exsit");

    initialize_array(g.board, board_length * board_length); 
	for(int i = 0; i < 10; i++){
        g.host[i] = 0;
		g.winner[i] = 0;
    }   

    ABA_db_put(gamekey, strlen(gamekey), &g, sizeof(game));

    if (strcmp(closer, player1) == 0){
    	//player1 need 1 ABA to close the game
		transfertoken("tictactoe", player1, "1");
		transfertoken("tictactoe", player2, "3");
    } else {
    	//player1 need 1 ABA to close the game
		transfertoken("tictactoe", player1, "3");
		transfertoken("tictactoe", player2, "1");
    }

    return 0;
}
//down on the chessboard and check if one of players win the game
int follow(char* player1, char* player2, char* host, int row, int column){
    int result, location;
    game g;
    int cell_value;
    char turn[10];
	char winner[10];
	char gamekey[20];
	char output[20];

	ABA_assert(is_valid_name(player1) != 0, "player1 name is too long or is null");
	ABA_assert(is_valid_name(player2) != 0, "player2 name is too long or is null");
	ABA_assert(is_valid_name(host) != 0, "host name is too long or is null");

	strcpy(gamekey,player1);
	strcat(gamekey,player2);

    result = ABA_db_get(gamekey, strlen(gamekey), &g, sizeof(game));

    ABA_assert(result != 0, "the game does not exsit");
	ABA_assert(strcmp(g.host, "") == 0, "the game is over");
    ABA_assert(strcmp(g.host, host) != 0, "it is not your turn to move");
    ABA_assert(is_valid_movement(row, column, g.board) != 0, "it is not a valid movement");

    if (strcmp(player1, host) == 0) {
        strcpy(turn, player2);
        cell_value = 1;
    }else {
		strcpy(turn, player1);
    	cell_value = 2;
	}

	location = (row - 1) * board_length + column - 1;
    g.board[location] = cell_value;
    strcpy(g.host, turn);
    result = get_winner(&g, winner);
	if (result == 9 && strcmp(winner,"") == 0) {
		for(int i = 0; i < 10; i++){
        	g.host[i] = 0;
    	}
		ABA_prints("the game is over and none of player win");
		ABA_db_put(gamekey, strlen(gamekey), &g, sizeof(game));
		transfertoken("tictactoe", player1, "2");
		transfertoken("tictactoe", player2, "2");

		return 0;
	}
	
	strcpy(g.winner, winner);
    
    if (strcmp(winner,"") != 0) {
		strcpy(output, "winner is ");
		strcat(output, winner);
        ABA_prints(output);
		for(int i = 0; i < 10; i++){
        	g.host[i] = 0;
    	}
		ABA_db_put(gamekey, strlen(gamekey), &g, sizeof(game));
    	transfertoken("tictactoe", winner, "4");
        return 0;
    }
	ABA_db_put(gamekey, strlen(gamekey), &g, sizeof(game));

    return 0;
}


// contract invoke entry
export int apply(char *method){
    if(strcmp(method, "create") == 0){
        char *player1, *player2;
        // get function parameters
        player1 = ABA_read_param(1);
        player2 = ABA_read_param(2);
        // call create function
        create(player1, player2);
    }
    if(strcmp(method, "close") == 0){
        char *player1, *player2, *closer;
        // get function parameters
        player1 = ABA_read_param(1);
        player2 = ABA_read_param(2);
		closer = ABA_read_param(3);
        // call close function
        close(player1, player2, closer);
    }
    if(strcmp(method,"restart") == 0){
        char *player1, *player2, *restarter;
        // get function parameters
        player1 = ABA_read_param(1);
        player2 = ABA_read_param(2);
    	restarter = ABA_read_param(3);
		
        // call restart function
        restart(player1, player2, restarter);
    }
    if(strcmp(method, "follow") == 0){
        char *player1, *player2, *host;
    	int row, column;
        // get function parameters
        player1 = ABA_read_param(1);
        player2 = ABA_read_param(2);
    	host = ABA_read_param(3);
    	row = ABA_read_param(4);
    	column = ABA_read_param(5);
        // call move function
        follow(player1, player2, host, row, column);
    }
    return 0;
}

