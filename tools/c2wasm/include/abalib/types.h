#pragma once

#include <stdint.h>
#include <wchar.h>

typedef uint64_t account_name;

typedef uint64_t permission_name;

typedef uint32_t time;

typedef uint64_t action_name;

typedef uint16_t weight_type;

#define ALIGNED(X) __attribute__ ((aligned (16))) X


struct public_key {
   char data[34];
};

struct signature {
   uint8_t data[66];
};


struct ALIGNED(checksum256) {
   uint8_t hash[32];
};


struct ALIGNED(checksum160) {
   uint8_t hash[20];
};


struct ALIGNED(checksum512) {
   uint8_t hash[64];
};

typedef struct checksum256 transaction_id_type;
typedef struct checksum256 block_id_type;

struct account_permission {
   account_name account;
   permission_name permission;
};


