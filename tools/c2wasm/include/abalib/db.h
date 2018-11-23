#pragma once

int ABA_db_put(char* key, uint32_t k_len, char *value, uint32_t v_len);
int ABA_db_get(char* key, uint32_t k_len, char *value, uint32_t v_len);
