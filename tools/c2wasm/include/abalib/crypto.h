#pragma once
#include <stdint.h>
#include <types.h>

void ABA_sha256( const char* data, uint32_t length, checksum256* hash );

void ABA_sha512( const char* data, uint32_t length, checksum512* hash );


