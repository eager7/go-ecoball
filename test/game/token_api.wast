(module
 (type $0 (func (param i32 i32)))
 (type $1 (func (result i32)))
 (type $2 (func (param i32) (result i32)))
 (type $3 (func (param i32 i32 i32 i32 i64) (result i32)))
 (type $4 (func (param i32 i32 i32 i32) (result i32)))
 (type $5 (func (param i32 i32) (result i32)))
 (type $6 (func (param i32 i64 i32) (result i32)))
 (type $7 (func (param i32 i32 i64 i32) (result i32)))
 (import "env" "ABA_add_token_balance" (func $import$0 (param i32 i32 i32 i32 i64) (result i32)))
 (import "env" "ABA_assert" (func $import$1 (param i32 i32)))
 (import "env" "ABA_get_token_balance" (func $import$2 (param i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_get_token_status" (func $import$3 (param i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_is_account" (func $import$4 (param i32 i32) (result i32)))
 (import "env" "ABA_put_token_status" (func $import$5 (param i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_read_param" (func $import$6 (param i32) (result i32)))
 (import "env" "ABA_require_auth" (func $import$7 (param i32 i32) (result i32)))
 (import "env" "ABA_sub_token_balance" (func $import$8 (param i32 i32 i32 i32 i64) (result i32)))
 (import "env" "ABA_token_Existed" (func $import$9 (param i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 4) "`R\00\00")
 (data (i32.const 16) "max_supply must be postive!\00")
 (data (i32.const 48) "token id must be all upper character\00")
 (data (i32.const 96) "The issuer account does not exist\00")
 (data (i32.const 144) "The token had existed\00")
 (data (i32.const 176) "amount must be postive!\00")
 (data (i32.const 208) "The receiving account does not exist\00")
 (data (i32.const 256) "The token does not exist\00")
 (data (i32.const 288) "The unsupplied token is not enough\00")
 (data (i32.const 336) "param is wrong, add balance failed\00")
 (data (i32.const 384) "can not transfer to self\00")
 (data (i32.const 416) "The transfer account does not exist\00")
 (data (i32.const 464) "can not get the transfer account\00")
 (data (i32.const 512) "The account balance is not enough\00")
 (data (i32.const 560) "create\00")
 (data (i32.const 576) "issue\00")
 (data (i32.const 592) "transfer\00")
 (export "memory" (memory $0))
 (export "apply" (func $4))
 (func $0 (type $2) (param $var$0 i32) (result i32)
  (local $var$1 i32)
  (local $var$2 i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (block $label$0 i32
   (set_local $var$4
    (i32.const 0)
   )
   (block $label$1
    (br_if $label$1
     (i32.eqz
      (tee_local $var$3
       (i32.load8_u
        (get_local $var$0)
       )
      )
     )
    )
    (block $label$2
     (br_if $label$2
      (i32.gt_u
       (i32.and
        (i32.add
         (get_local $var$3)
         (i32.const -65)
        )
        (i32.const 255)
       )
       (i32.const 25)
      )
     )
     (set_local $var$1
      (call $8
       (get_local $var$0)
      )
     )
     (set_local $var$3
      (i32.const 1)
     )
     (loop $label$3
      (br_if $label$1
       (i32.ge_u
        (get_local $var$3)
        (get_local $var$1)
       )
      )
      (set_local $var$2
       (i32.add
        (get_local $var$0)
        (get_local $var$3)
       )
      )
      (set_local $var$3
       (i32.add
        (get_local $var$3)
        (i32.const 1)
       )
      )
      (br_if $label$3
       (i32.lt_u
        (i32.and
         (i32.add
          (i32.load8_u
           (get_local $var$2)
          )
          (i32.const -65)
         )
         (i32.const 255)
        )
        (i32.const 26)
       )
      )
     )
    )
    (set_local $var$4
     (i32.const -1)
    )
   )
   (get_local $var$4)
  )
 )
 (func $1 (type $6) (param $var$0 i32) (param $var$1 i64) (param $var$2 i32) (result i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (local $var$7 i32)
  (block $label$0 i32
   (i32.store offset=4
    (i32.const 0)
    (tee_local $var$7
     (i32.sub
      (i32.load offset=4
       (i32.const 0)
      )
      (i32.const 48)
     )
    )
   )
   (call $import$1
    (i64.lt_s
     (get_local $var$1)
     (i64.const 1)
    )
    (i32.const 16)
   )
   (set_local $var$6
    (i32.const 0)
   )
   (block $label$1
    (br_if $label$1
     (i32.eqz
      (tee_local $var$5
       (i32.load8_u
        (get_local $var$2)
       )
      )
     )
    )
    (set_local $var$6
     (i32.const -1)
    )
    (br_if $label$1
     (i32.gt_u
      (i32.and
       (i32.add
        (get_local $var$5)
        (i32.const -65)
       )
       (i32.const 255)
      )
      (i32.const 25)
     )
    )
    (set_local $var$3
     (call $8
      (get_local $var$2)
     )
    )
    (set_local $var$5
     (i32.const 1)
    )
    (block $label$2
     (loop $label$3
      (br_if $label$2
       (i32.ge_u
        (get_local $var$5)
        (get_local $var$3)
       )
      )
      (set_local $var$4
       (i32.add
        (get_local $var$2)
        (get_local $var$5)
       )
      )
      (set_local $var$5
       (i32.add
        (get_local $var$5)
        (i32.const 1)
       )
      )
      (br_if $label$3
       (i32.lt_u
        (i32.and
         (i32.add
          (i32.load8_u
           (get_local $var$4)
          )
          (i32.const -65)
         )
         (i32.const 255)
        )
        (i32.const 26)
       )
      )
      (br $label$1)
     )
     (unreachable)
    )
    (set_local $var$6
     (i32.const 0)
    )
   )
   (call $import$1
    (get_local $var$6)
    (i32.const 48)
   )
   (call $import$1
    (i32.ne
     (call $import$4
      (get_local $var$0)
      (call $8
       (get_local $var$0)
      )
     )
     (i32.const 0)
    )
    (i32.const 96)
   )
   (call $import$1
    (i32.eq
     (call $import$9
      (get_local $var$2)
      (call $8
       (get_local $var$2)
      )
     )
     (i32.const 1)
    )
    (i32.const 144)
   )
   (i64.store offset=16
    (get_local $var$7)
    (get_local $var$1)
   )
   (i64.store offset=24
    (get_local $var$7)
    (i64.const 0)
   )
   (drop
    (call $6
     (i32.add
      (get_local $var$7)
      (i32.const 32)
     )
     (get_local $var$0)
    )
   )
   (drop
    (call $6
     (get_local $var$7)
     (get_local $var$2)
    )
   )
   (drop
    (call $import$5
     (get_local $var$2)
     (call $8
      (get_local $var$2)
     )
     (get_local $var$7)
     (i32.const 48)
    )
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$7)
     (i32.const 48)
    )
   )
   (i32.const 0)
  )
 )
 (func $2 (type $6) (param $var$0 i32) (param $var$1 i64) (param $var$2 i32) (result i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (local $var$7 i32)
  (block $label$0 i32
   (i32.store offset=4
    (i32.const 0)
    (tee_local $var$7
     (i32.sub
      (i32.load offset=4
       (i32.const 0)
      )
      (i32.const 48)
     )
    )
   )
   (call $import$1
    (i64.lt_s
     (get_local $var$1)
     (i64.const 1)
    )
    (i32.const 176)
   )
   (set_local $var$6
    (i32.const 0)
   )
   (block $label$1
    (br_if $label$1
     (i32.eqz
      (tee_local $var$5
       (i32.load8_u
        (get_local $var$2)
       )
      )
     )
    )
    (set_local $var$6
     (i32.const -1)
    )
    (br_if $label$1
     (i32.gt_u
      (i32.and
       (i32.add
        (get_local $var$5)
        (i32.const -65)
       )
       (i32.const 255)
      )
      (i32.const 25)
     )
    )
    (set_local $var$3
     (call $8
      (get_local $var$2)
     )
    )
    (set_local $var$5
     (i32.const 1)
    )
    (block $label$2
     (loop $label$3
      (br_if $label$2
       (i32.ge_u
        (get_local $var$5)
        (get_local $var$3)
       )
      )
      (set_local $var$4
       (i32.add
        (get_local $var$2)
        (get_local $var$5)
       )
      )
      (set_local $var$5
       (i32.add
        (get_local $var$5)
        (i32.const 1)
       )
      )
      (br_if $label$3
       (i32.lt_u
        (i32.and
         (i32.add
          (i32.load8_u
           (get_local $var$4)
          )
          (i32.const -65)
         )
         (i32.const 255)
        )
        (i32.const 26)
       )
      )
      (br $label$1)
     )
     (unreachable)
    )
    (set_local $var$6
     (i32.const 0)
    )
   )
   (call $import$1
    (get_local $var$6)
    (i32.const 48)
   )
   (call $import$1
    (i32.ne
     (call $import$4
      (get_local $var$0)
      (call $8
       (get_local $var$0)
      )
     )
     (i32.const 0)
    )
    (i32.const 208)
   )
   (call $import$1
    (i32.ne
     (call $import$3
      (get_local $var$2)
      (call $8
       (get_local $var$2)
      )
      (get_local $var$7)
      (i32.const 48)
     )
     (i32.const 0)
    )
    (i32.const 256)
   )
   (drop
    (call $import$7
     (tee_local $var$5
      (i32.add
       (get_local $var$7)
       (i32.const 32)
      )
     )
     (call $8
      (get_local $var$5)
     )
    )
   )
   (call $import$1
    (i64.lt_s
     (i64.sub
      (i64.load offset=16
       (get_local $var$7)
      )
      (i64.load offset=24
       (get_local $var$7)
      )
     )
     (get_local $var$1)
    )
    (i32.const 288)
   )
   (call $import$1
    (i32.ne
     (call $import$0
      (get_local $var$0)
      (call $8
       (get_local $var$0)
      )
      (get_local $var$2)
      (call $8
       (get_local $var$2)
      )
      (get_local $var$1)
     )
     (i32.const 0)
    )
    (i32.const 336)
   )
   (i64.store offset=24
    (get_local $var$7)
    (i64.add
     (i64.load offset=24
      (get_local $var$7)
     )
     (get_local $var$1)
    )
   )
   (call $import$1
    (i32.ne
     (call $import$5
      (get_local $var$2)
      (call $8
       (get_local $var$2)
      )
      (get_local $var$7)
      (i32.const 48)
     )
     (i32.const 0)
    )
    (i32.const 256)
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$7)
     (i32.const 48)
    )
   )
   (i32.const 0)
  )
 )
 (func $3 (type $7) (param $var$0 i32) (param $var$1 i32) (param $var$2 i64) (param $var$3 i32) (result i32)
  (local $var$4 i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (local $var$7 i32)
  (block $label$0 i32
   (call $import$1
    (i64.lt_s
     (get_local $var$2)
     (i64.const 1)
    )
    (i32.const 176)
   )
   (call $import$1
    (i32.eqz
     (call $5
      (get_local $var$0)
      (get_local $var$1)
     )
    )
    (i32.const 384)
   )
   (set_local $var$7
    (i32.const 0)
   )
   (block $label$1
    (br_if $label$1
     (i32.eqz
      (tee_local $var$6
       (i32.load8_u
        (get_local $var$3)
       )
      )
     )
    )
    (set_local $var$7
     (i32.const -1)
    )
    (br_if $label$1
     (i32.gt_u
      (i32.and
       (i32.add
        (get_local $var$6)
        (i32.const -65)
       )
       (i32.const 255)
      )
      (i32.const 25)
     )
    )
    (set_local $var$4
     (call $8
      (get_local $var$3)
     )
    )
    (set_local $var$6
     (i32.const 1)
    )
    (block $label$2
     (loop $label$3
      (br_if $label$2
       (i32.ge_u
        (get_local $var$6)
        (get_local $var$4)
       )
      )
      (set_local $var$5
       (i32.add
        (get_local $var$3)
        (get_local $var$6)
       )
      )
      (set_local $var$6
       (i32.add
        (get_local $var$6)
        (i32.const 1)
       )
      )
      (br_if $label$3
       (i32.lt_u
        (i32.and
         (i32.add
          (i32.load8_u
           (get_local $var$5)
          )
          (i32.const -65)
         )
         (i32.const 255)
        )
        (i32.const 26)
       )
      )
      (br $label$1)
     )
     (unreachable)
    )
    (set_local $var$7
     (i32.const 0)
    )
   )
   (call $import$1
    (get_local $var$7)
    (i32.const 48)
   )
   (call $import$1
    (i32.ne
     (call $import$4
      (get_local $var$0)
      (call $8
       (get_local $var$0)
      )
     )
     (i32.const 0)
    )
    (i32.const 416)
   )
   (call $import$1
    (i32.ne
     (call $import$4
      (get_local $var$1)
      (call $8
       (get_local $var$1)
      )
     )
     (i32.const 0)
    )
    (i32.const 208)
   )
   (drop
    (call $import$7
     (get_local $var$0)
     (call $8
      (get_local $var$0)
     )
    )
   )
   (call $import$1
    (i32.shr_u
     (tee_local $var$6
      (call $import$2
       (get_local $var$0)
       (call $8
        (get_local $var$0)
       )
       (get_local $var$3)
       (call $8
        (get_local $var$3)
       )
      )
     )
     (i32.const 31)
    )
    (i32.const 464)
   )
   (call $import$1
    (i64.lt_s
     (i64.extend_s/i32
      (get_local $var$6)
     )
     (get_local $var$2)
    )
    (i32.const 512)
   )
   (call $import$1
    (i32.ne
     (call $import$8
      (get_local $var$0)
      (call $8
       (get_local $var$0)
      )
      (get_local $var$3)
      (call $8
       (get_local $var$3)
      )
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 336)
   )
   (call $import$1
    (i32.ne
     (call $import$0
      (get_local $var$1)
      (call $8
       (get_local $var$1)
      )
      (get_local $var$3)
      (call $8
       (get_local $var$3)
      )
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 336)
   )
   (call $import$1
    (i32.ne
     (call $import$0
      (get_local $var$0)
      (call $8
       (get_local $var$0)
      )
      (get_local $var$3)
      (call $8
       (get_local $var$3)
      )
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 336)
   )
   (call $import$1
    (i32.ne
     (call $import$8
      (get_local $var$1)
      (call $8
       (get_local $var$1)
      )
      (get_local $var$3)
      (call $8
       (get_local $var$3)
      )
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 336)
   )
   (call $import$1
    (i32.ne
     (call $import$8
      (get_local $var$0)
      (call $8
       (get_local $var$0)
      )
      (get_local $var$3)
      (call $8
       (get_local $var$3)
      )
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 336)
   )
   (call $import$1
    (i32.ne
     (call $import$0
      (get_local $var$1)
      (call $8
       (get_local $var$1)
      )
      (get_local $var$3)
      (call $8
       (get_local $var$3)
      )
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 336)
   )
   (i32.const 0)
  )
 )
 (func $4 (type $2) (param $var$0 i32) (result i32)
  (block $label$0 i32
   (block $label$1
    (br_if $label$1
     (call $5
      (get_local $var$0)
      (i32.const 560)
     )
    )
    (drop
     (call $1
      (call $import$6
       (i32.const 1)
      )
      (i64.extend_u/i32
       (call $import$6
        (i32.const 2)
       )
      )
      (call $import$6
       (i32.const 3)
      )
     )
    )
   )
   (block $label$2
    (br_if $label$2
     (call $5
      (get_local $var$0)
      (i32.const 576)
     )
    )
    (drop
     (call $2
      (call $import$6
       (i32.const 1)
      )
      (i64.extend_u/i32
       (call $import$6
        (i32.const 2)
       )
      )
      (call $import$6
       (i32.const 3)
      )
     )
    )
   )
   (block $label$3
    (br_if $label$3
     (i32.eqz
      (call $5
       (get_local $var$0)
       (i32.const 592)
      )
     )
    )
    (return
     (i32.const 0)
    )
   )
   (drop
    (call $3
     (call $import$6
      (i32.const 1)
     )
     (call $import$6
      (i32.const 2)
     )
     (i64.extend_u/i32
      (call $import$6
       (i32.const 3)
      )
     )
     (call $import$6
      (i32.const 4)
     )
    )
   )
   (i32.const 0)
  )
 )
 (func $5 (type $5) (param $var$0 i32) (param $var$1 i32) (result i32)
  (local $var$2 i32)
  (local $var$3 i32)
  (block $label$0 i32
   (set_local $var$3
    (i32.load8_u
     (get_local $var$1)
    )
   )
   (block $label$1
    (br_if $label$1
     (i32.eqz
      (tee_local $var$2
       (i32.load8_u
        (get_local $var$0)
       )
      )
     )
    )
    (br_if $label$1
     (i32.ne
      (get_local $var$2)
      (i32.and
       (get_local $var$3)
       (i32.const 255)
      )
     )
    )
    (set_local $var$0
     (i32.add
      (get_local $var$0)
      (i32.const 1)
     )
    )
    (set_local $var$1
     (i32.add
      (get_local $var$1)
      (i32.const 1)
     )
    )
    (loop $label$2
     (set_local $var$3
      (i32.load8_u
       (get_local $var$1)
      )
     )
     (br_if $label$1
      (i32.eqz
       (tee_local $var$2
        (i32.load8_u
         (get_local $var$0)
        )
       )
      )
     )
     (set_local $var$0
      (i32.add
       (get_local $var$0)
       (i32.const 1)
      )
     )
     (set_local $var$1
      (i32.add
       (get_local $var$1)
       (i32.const 1)
      )
     )
     (br_if $label$2
      (i32.eq
       (get_local $var$2)
       (i32.and
        (get_local $var$3)
        (i32.const 255)
       )
      )
     )
    )
   )
   (i32.sub
    (get_local $var$2)
    (i32.and
     (get_local $var$3)
     (i32.const 255)
    )
   )
  )
 )
 (func $6 (type $5) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (drop
    (call $7
     (get_local $var$0)
     (get_local $var$1)
    )
   )
   (get_local $var$0)
  )
 )
 (func $7 (type $5) (param $var$0 i32) (param $var$1 i32) (result i32)
  (local $var$2 i32)
  (block $label$0 i32
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.and
       (i32.xor
        (get_local $var$1)
        (get_local $var$0)
       )
       (i32.const 3)
      )
     )
     (block $label$3
      (br_if $label$3
       (i32.eqz
        (i32.and
         (get_local $var$1)
         (i32.const 3)
        )
       )
      )
      (loop $label$4
       (i32.store8
        (get_local $var$0)
        (tee_local $var$2
         (i32.load8_u
          (get_local $var$1)
         )
        )
       )
       (br_if $label$1
        (i32.eqz
         (get_local $var$2)
        )
       )
       (set_local $var$0
        (i32.add
         (get_local $var$0)
         (i32.const 1)
        )
       )
       (br_if $label$4
        (i32.and
         (tee_local $var$1
          (i32.add
           (get_local $var$1)
           (i32.const 1)
          )
         )
         (i32.const 3)
        )
       )
      )
     )
     (br_if $label$2
      (i32.and
       (i32.and
        (i32.xor
         (tee_local $var$2
          (i32.load
           (get_local $var$1)
          )
         )
         (i32.const -1)
        )
        (i32.add
         (get_local $var$2)
         (i32.const -16843009)
        )
       )
       (i32.const -2139062144)
      )
     )
     (loop $label$5
      (i32.store
       (get_local $var$0)
       (get_local $var$2)
      )
      (set_local $var$2
       (i32.load offset=4
        (get_local $var$1)
       )
      )
      (set_local $var$0
       (i32.add
        (get_local $var$0)
        (i32.const 4)
       )
      )
      (set_local $var$1
       (i32.add
        (get_local $var$1)
        (i32.const 4)
       )
      )
      (br_if $label$5
       (i32.eqz
        (i32.and
         (i32.and
          (i32.xor
           (get_local $var$2)
           (i32.const -1)
          )
          (i32.add
           (get_local $var$2)
           (i32.const -16843009)
          )
         )
         (i32.const -2139062144)
        )
       )
      )
     )
    )
    (i32.store8
     (get_local $var$0)
     (tee_local $var$2
      (i32.load8_u
       (get_local $var$1)
      )
     )
    )
    (br_if $label$1
     (i32.eqz
      (get_local $var$2)
     )
    )
    (set_local $var$1
     (i32.add
      (get_local $var$1)
      (i32.const 1)
     )
    )
    (loop $label$6
     (i32.store8 offset=1
      (get_local $var$0)
      (tee_local $var$2
       (i32.load8_u
        (get_local $var$1)
       )
      )
     )
     (set_local $var$1
      (i32.add
       (get_local $var$1)
       (i32.const 1)
      )
     )
     (set_local $var$0
      (i32.add
       (get_local $var$0)
       (i32.const 1)
      )
     )
     (br_if $label$6
      (get_local $var$2)
     )
    )
   )
   (get_local $var$0)
  )
 )
 (func $8 (type $2) (param $var$0 i32) (result i32)
  (local $var$1 i32)
  (local $var$2 i32)
  (local $var$3 i32)
  (block $label$0 i32
   (set_local $var$3
    (get_local $var$0)
   )
   (block $label$1
    (block $label$2
     (block $label$3
      (br_if $label$3
       (i32.eqz
        (i32.and
         (get_local $var$0)
         (i32.const 3)
        )
       )
      )
      (set_local $var$3
       (get_local $var$0)
      )
      (loop $label$4
       (br_if $label$2
        (i32.eqz
         (i32.load8_u
          (get_local $var$3)
         )
        )
       )
       (br_if $label$4
        (i32.and
         (tee_local $var$3
          (i32.add
           (get_local $var$3)
           (i32.const 1)
          )
         )
         (i32.const 3)
        )
       )
      )
     )
     (set_local $var$2
      (i32.add
       (get_local $var$3)
       (i32.const -4)
      )
     )
     (loop $label$5
      (br_if $label$5
       (i32.eqz
        (i32.and
         (i32.and
          (i32.xor
           (tee_local $var$3
            (i32.load
             (tee_local $var$2
              (i32.add
               (get_local $var$2)
               (i32.const 4)
              )
             )
            )
           )
           (i32.const -1)
          )
          (i32.add
           (get_local $var$3)
           (i32.const -16843009)
          )
         )
         (i32.const -2139062144)
        )
       )
      )
     )
     (br_if $label$1
      (i32.eqz
       (i32.and
        (get_local $var$3)
        (i32.const 255)
       )
      )
     )
     (loop $label$6
      (set_local $var$1
       (i32.load8_u offset=1
        (get_local $var$2)
       )
      )
      (set_local $var$2
       (tee_local $var$3
        (i32.add
         (get_local $var$2)
         (i32.const 1)
        )
       )
      )
      (br_if $label$6
       (get_local $var$1)
      )
     )
    )
    (return
     (i32.sub
      (get_local $var$3)
      (get_local $var$0)
     )
    )
   )
   (i32.sub
    (get_local $var$2)
    (get_local $var$0)
   )
  )
 )
)

