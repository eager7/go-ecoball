(module
 (type $0 (func (param i64)))
 (type $1 (func (param i32)))
 (type $2 (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32) (result i32)))
 (type $3 (func (param i32 i32)))
 (type $4 (func (result i32)))
 (type $5 (func (param i32 i32 i32 i32) (result i32)))
 (type $6 (func (param i32) (result i32)))
 (type $7 (func (param i32 i32) (result i32)))
 (type $8 (func (param i32 i32 i32) (result i32)))
 (type $9 (func (param i32 i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_assert" (func $import$0 (param i32 i32)))
 (import "env" "ABA_db_get" (func $import$1 (param i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_db_put" (func $import$2 (param i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_inline_action" (func $import$3 (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_is_account" (func $import$4 (param i32 i32) (result i32)))
 (import "env" "ABA_printi" (func $import$5 (param i64)))
 (import "env" "ABA_prints" (func $import$6 (param i32)))
 (import "env" "ABA_read_param" (func $import$7 (param i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 4) "\a0V\00\00")
 (data (i32.const 12) "\03\00\00\00")
 (data (i32.const 16) "[\"\00")
 (data (i32.const 32) "\",\"ABA\"]\00")
 (data (i32.const 48) "abatoken\00")
 (data (i32.const 64) "transfer\00")
 (data (i32.const 80) "active\00")
 (data (i32.const 96) "player1 name is too long or is null\00")
 (data (i32.const 144) "player2 name is too long or is null\00")
 (data (i32.const 192) "player1 is not a account\00")
 (data (i32.const 224) "player2 is not a account\00")
 (data (i32.const 256) "player1 shouldn\'t be the same as player2\00")
 (data (i32.const 304) "tictactoe\00")
 (data (i32.const 320) "2\00")
 (data (i32.const 336) "restarter name is too long or is null\00")
 (data (i32.const 384) "restarter has insufficient permission\00")
 (data (i32.const 432) "the game does not exsit\00")
 (data (i32.const 464) "4\00")
 (data (i32.const 480) "closer name is too long or is null\00")
 (data (i32.const 528) "closer has insufficient permission\00")
 (data (i32.const 576) "1\00")
 (data (i32.const 592) "3\00")
 (data (i32.const 608) "host name is too long or is null\00")
 (data (i32.const 656) "the game is over\00")
 (data (i32.const 688) "it is not your turn to move\00")
 (data (i32.const 720) "it is not a valid movement\00")
 (data (i32.const 752) "the game is over and none of player win\00")
 (data (i32.const 800) "winner is \00")
 (data (i32.const 816) "winner is inconclusive\00")
 (data (i32.const 848) "create\00")
 (data (i32.const 864) "close\00")
 (data (i32.const 880) "restart\00")
 (data (i32.const 896) "follow\00")
 (data (i32.const 1688) "\00\00\00\00")
 (export "memory" (memory $0))
 (export "apply" (func $11))
 (func $0 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
  (local $var$2 i64)
  (local $var$3 i64)
  (block $label$0 i32
   (block $label$1
    (br_if $label$1
     (i32.lt_s
      (get_local $var$1)
      (i32.const 1)
     )
    )
    (set_local $var$2
     (i64.extend_u/i32
      (get_local $var$1)
     )
    )
    (set_local $var$3
     (i64.const 0)
    )
    (loop $label$2
     (call $import$5
      (get_local $var$3)
     )
     (i32.store
      (get_local $var$0)
      (i32.const 0)
     )
     (set_local $var$0
      (i32.add
       (get_local $var$0)
       (i32.const 4)
      )
     )
     (br_if $label$2
      (i64.ne
       (get_local $var$2)
       (tee_local $var$3
        (i64.add
         (get_local $var$3)
         (i64.const 1)
        )
       )
      )
     )
    )
   )
   (i32.const 0)
  )
 )
 (func $1 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (call $import$5
    (i64.const 0)
   )
   (i32.store
    (get_local $var$0)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 1)
   )
   (i32.store offset=4
    (get_local $var$0)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 2)
   )
   (i64.store align=1
    (i32.add
     (get_local $var$0)
     (i32.const 68)
    )
    (i64.const 0)
   )
   (i64.store align=1
    (i32.add
     (get_local $var$0)
     (i32.const 60)
    )
    (i64.const 0)
   )
   (i32.store offset=8
    (get_local $var$0)
    (i32.const 0)
   )
   (i32.store offset=56 align=1
    (get_local $var$0)
    (i32.const 0)
   )
   (drop
    (call $17
     (i32.add
      (get_local $var$0)
      (i32.const 56)
     )
     (get_local $var$1)
    )
   )
   (i32.const 0)
  )
 )
 (func $2 (type $6) (param $var$0 i32) (result i32)
  (select
   (i32.const -1)
   (i32.const 1)
   (get_local $var$0)
  )
 )
 (func $3 (type $8) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
  (block $label$0 i32
   (block $label$1
    (br_if $label$1
     (i32.gt_u
      (i32.add
       (get_local $var$0)
       (i32.const -1)
      )
      (i32.const 2)
     )
    )
    (br_if $label$1
     (i32.gt_u
      (i32.add
       (get_local $var$1)
       (i32.const -1)
      )
      (i32.const 2)
     )
    )
    (return
     (select
      (i32.const -1)
      (i32.const 0)
      (i32.gt_s
       (i32.add
        (i32.add
         (i32.mul
          (get_local $var$0)
          (i32.const 3)
         )
         (get_local $var$1)
        )
        (i32.const -3)
       )
       (i32.const 9)
      )
     )
    )
   )
   (i32.const -1)
  )
 )
 (func $4 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
  (local $var$2 i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (local $var$7 i32)
  (block $label$0 i32
   (call $import$5
    (i64.const 0)
   )
   (call $import$5
    (i64.const 1)
   )
   (call $import$5
    (i64.const 2)
   )
   (call $import$5
    (i64.const 0)
   )
   (call $import$5
    (i64.const 1)
   )
   (call $import$5
    (i64.const 2)
   )
   (set_local $var$5
    (i32.add
     (i32.add
      (i32.add
       (i32.add
        (i32.add
         (i32.add
          (i32.add
           (i32.add
            (i32.ne
             (tee_local $var$6
              (i32.load
               (get_local $var$0)
              )
             )
             (i32.const 0)
            )
            (i32.ne
             (i32.load offset=4
              (get_local $var$0)
             )
             (i32.const 0)
            )
           )
           (i32.ne
            (tee_local $var$2
             (i32.load offset=8
              (get_local $var$0)
             )
            )
            (i32.const 0)
           )
          )
          (i32.ne
           (i32.load offset=12
            (get_local $var$0)
           )
           (i32.const 0)
          )
         )
         (i32.ne
          (tee_local $var$3
           (i32.load offset=16
            (get_local $var$0)
           )
          )
          (i32.const 0)
         )
        )
        (i32.ne
         (i32.load offset=20
          (get_local $var$0)
         )
         (i32.const 0)
        )
       )
       (i32.ne
        (tee_local $var$4
         (i32.load offset=24
          (get_local $var$0)
         )
        )
        (i32.const 0)
       )
      )
      (i32.ne
       (i32.load offset=28
        (get_local $var$0)
       )
       (i32.const 0)
      )
     )
     (i32.ne
      (tee_local $var$7
       (i32.load offset=32
        (get_local $var$0)
       )
      )
      (i32.const 0)
     )
    )
   )
   (block $label$1
    (block $label$2
     (block $label$3
      (br_if $label$3
       (i32.eq
        (tee_local $var$6
         (i32.and
          (i32.and
           (get_local $var$7)
           (i32.and
            (get_local $var$3)
            (get_local $var$6)
           )
          )
          (i32.const 3)
         )
        )
        (i32.const 2)
       )
      )
      (br_if $label$1
       (i32.eq
        (get_local $var$6)
        (i32.const 1)
       )
      )
      (br_if $label$1
       (i32.eq
        (tee_local $var$3
         (i32.and
          (i32.and
           (get_local $var$4)
           (i32.and
            (get_local $var$3)
            (get_local $var$2)
           )
          )
          (i32.const 3)
         )
        )
        (i32.const 1)
       )
      )
      (br_if $label$2
       (i32.ne
        (get_local $var$3)
        (i32.const 2)
       )
      )
     )
     (drop
      (call $17
       (get_local $var$1)
       (i32.add
        (get_local $var$0)
        (i32.const 46)
       )
      )
     )
    )
    (return
     (get_local $var$5)
    )
   )
   (drop
    (call $17
     (get_local $var$1)
     (i32.add
      (get_local $var$0)
      (i32.const 36)
     )
    )
   )
   (get_local $var$5)
  )
 )
 (func $5 (type $8) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (block $label$0 i32
   (i32.store offset=4
    (i32.const 0)
    (tee_local $var$4
     (i32.sub
      (i32.load offset=4
       (i32.const 0)
      )
      (i32.const 64)
     )
    )
   )
   (i32.store8 offset=2
    (get_local $var$4)
    (i32.load8_u offset=18
     (i32.const 0)
    )
   )
   (i32.store16
    (get_local $var$4)
    (i32.load16_u offset=16 align=1
     (i32.const 0)
    )
   )
   (i32.store align=1
    (i32.add
     (get_local $var$4)
     (call $16
      (tee_local $var$3
       (call $15
        (get_local $var$4)
        (get_local $var$0)
       )
      )
     )
    )
    (i32.const 2239522)
   )
   (i32.store align=1
    (i32.add
     (get_local $var$4)
     (call $16
      (tee_local $var$1
       (call $15
        (get_local $var$3)
        (get_local $var$1)
       )
      )
     )
    )
    (i32.const 2239522)
   )
   (i64.store align=1
    (tee_local $var$1
     (i32.add
      (get_local $var$4)
      (call $16
       (tee_local $var$2
        (call $15
         (get_local $var$1)
         (get_local $var$2)
        )
       )
      )
     )
    )
    (i64.load offset=32 align=1
     (i32.const 0)
    )
   )
   (i32.store8
    (i32.add
     (get_local $var$1)
     (i32.const 8)
    )
    (i32.load8_u offset=40
     (i32.const 0)
    )
   )
   (call $import$6
    (get_local $var$2)
   )
   (drop
    (call $import$3
     (i32.const 48)
     (i32.const 8)
     (i32.const 64)
     (i32.const 8)
     (get_local $var$2)
     (call $16
      (get_local $var$2)
     )
     (get_local $var$0)
     (call $16
      (get_local $var$0)
     )
     (i32.const 80)
     (i32.const 6)
    )
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$4)
     (i32.const 64)
    )
   )
   (i32.const 0)
  )
 )
 (func $6 (type $6) (param $var$0 i32) (result i32)
  (select
   (i32.const -1)
   (i32.const 0)
   (i32.gt_s
    (call $16
     (get_local $var$0)
    )
    (i32.const 10)
   )
  )
 )
 (func $7 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
  (local $var$2 i32)
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
      (i32.const 32)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$1)
     )
     (i32.const 10)
    )
    (i32.const 144)
   )
   (call $import$0
    (i32.ne
     (call $import$4
      (get_local $var$0)
      (call $16
       (get_local $var$0)
      )
     )
     (i32.const 0)
    )
    (i32.const 192)
   )
   (call $import$0
    (i32.ne
     (call $import$4
      (get_local $var$1)
      (call $16
       (get_local $var$1)
      )
     )
     (i32.const 0)
    )
    (i32.const 224)
   )
   (call $import$0
    (i32.eqz
     (call $19
      (get_local $var$0)
      (get_local $var$1)
     )
    )
    (i32.const 256)
   )
   (drop
    (call $12)
   )
   (set_local $var$2
    (call $14
     (i32.const 76)
    )
   )
   (call $import$5
    (i64.const 0)
   )
   (i32.store
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 1)
   )
   (i32.store offset=4
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 2)
   )
   (i32.store offset=8
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 3)
   )
   (i32.store offset=12
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 4)
   )
   (i32.store offset=16
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 5)
   )
   (i32.store offset=20
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 6)
   )
   (i32.store offset=24
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 7)
   )
   (i32.store offset=28
    (get_local $var$2)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 8)
   )
   (i32.store offset=32
    (get_local $var$2)
    (i32.const 0)
   )
   (set_local $var$3
    (call $17
     (i32.add
      (get_local $var$2)
      (i32.const 36)
     )
     (get_local $var$0)
    )
   )
   (set_local $var$4
    (call $17
     (i32.add
      (get_local $var$2)
      (i32.const 46)
     )
     (get_local $var$1)
    )
   )
   (set_local $var$5
    (call $17
     (i32.add
      (get_local $var$2)
      (i32.const 56)
     )
     (get_local $var$0)
    )
   )
   (i32.store16 align=1
    (i32.add
     (get_local $var$2)
     (i32.const 74)
    )
    (i32.const 0)
   )
   (i64.store offset=66 align=1
    (get_local $var$2)
    (i64.const 0)
   )
   (drop
    (call $import$2
     (tee_local $var$6
      (call $15
       (call $17
        (get_local $var$7)
        (get_local $var$0)
       )
       (get_local $var$1)
      )
     )
     (call $16
      (get_local $var$6)
     )
     (get_local $var$2)
     (i32.const 76)
    )
   )
   (drop
    (call $5
     (get_local $var$0)
     (i32.const 304)
     (i32.const 320)
    )
   )
   (drop
    (call $5
     (get_local $var$1)
     (i32.const 304)
     (i32.const 320)
    )
   )
   (call $import$6
    (get_local $var$3)
   )
   (call $import$6
    (get_local $var$4)
   )
   (call $import$6
    (get_local $var$5)
   )
   (call $import$6
    (i32.add
     (get_local $var$2)
     (i32.const 66)
    )
   )
   (call $import$5
    (i64.load32_s
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=4
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=8
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=12
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=16
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=20
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=24
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=28
     (get_local $var$2)
    )
   )
   (call $import$5
    (i64.load32_s offset=32
     (get_local $var$2)
    )
   )
   (call $13
    (get_local $var$2)
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$7)
     (i32.const 32)
    )
   )
   (i32.const 0)
  )
 )
 (func $8 (type $8) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
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
      (i32.const 32)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$1)
     )
     (i32.const 10)
    )
    (i32.const 144)
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$2)
     )
     (i32.const 10)
    )
    (i32.const 336)
   )
   (drop
    (call $12)
   )
   (set_local $var$3
    (call $14
     (i32.const 76)
    )
   )
   (call $import$0
    (i32.and
     (i32.ne
      (call $19
       (get_local $var$0)
       (get_local $var$2)
      )
      (i32.const 0)
     )
     (i32.ne
      (call $19
       (get_local $var$1)
       (get_local $var$2)
      )
      (i32.const 0)
     )
    )
    (i32.const 384)
   )
   (set_local $var$5
    (call $import$1
     (tee_local $var$4
      (call $15
       (call $17
        (get_local $var$7)
        (get_local $var$0)
       )
       (get_local $var$1)
      )
     )
     (call $16
      (get_local $var$4)
     )
     (get_local $var$3)
     (i32.const 76)
    )
   )
   (call $import$6
    (i32.add
     (get_local $var$3)
     (i32.const 36)
    )
   )
   (call $import$6
    (i32.add
     (get_local $var$3)
     (i32.const 46)
    )
   )
   (call $import$6
    (tee_local $var$6
     (i32.add
      (get_local $var$3)
      (i32.const 56)
     )
    )
   )
   (call $import$6
    (i32.add
     (get_local $var$3)
     (i32.const 66)
    )
   )
   (call $import$5
    (i64.load32_s
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=4
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=8
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=12
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=16
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=20
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=24
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=28
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=32
     (get_local $var$3)
    )
   )
   (call $import$0
    (i32.ne
     (get_local $var$5)
     (i32.const 0)
    )
    (i32.const 432)
   )
   (call $import$5
    (i64.const 0)
   )
   (i32.store
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 1)
   )
   (i32.store offset=4
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 2)
   )
   (i64.store align=1
    (i32.add
     (get_local $var$3)
     (i32.const 68)
    )
    (i64.const 0)
   )
   (i64.store align=1
    (i32.add
     (get_local $var$3)
     (i32.const 60)
    )
    (i64.const 0)
   )
   (i32.store offset=8
    (get_local $var$3)
    (i32.const 0)
   )
   (i32.store offset=56 align=1
    (get_local $var$3)
    (i32.const 0)
   )
   (drop
    (call $17
     (get_local $var$6)
     (get_local $var$2)
    )
   )
   (drop
    (call $import$2
     (get_local $var$4)
     (call $16
      (get_local $var$4)
     )
     (get_local $var$3)
     (i32.const 76)
    )
   )
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (call $19
        (get_local $var$2)
        (get_local $var$0)
       )
      )
     )
     (drop
      (call $5
       (get_local $var$1)
       (i32.const 304)
       (i32.const 464)
      )
     )
     (br $label$1)
    )
    (drop
     (call $5
      (get_local $var$0)
      (i32.const 304)
      (i32.const 464)
     )
    )
   )
   (call $13
    (get_local $var$3)
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$7)
     (i32.const 32)
    )
   )
   (i32.const 0)
  )
 )
 (func $9 (type $8) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (block $label$0 i32
   (i32.store offset=4
    (i32.const 0)
    (tee_local $var$6
     (i32.sub
      (i32.load offset=4
       (i32.const 0)
      )
      (i32.const 32)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$1)
     )
     (i32.const 10)
    )
    (i32.const 144)
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$2)
     )
     (i32.const 10)
    )
    (i32.const 480)
   )
   (call $import$0
    (i32.and
     (i32.ne
      (call $19
       (get_local $var$0)
       (get_local $var$2)
      )
      (i32.const 0)
     )
     (i32.ne
      (call $19
       (get_local $var$1)
       (get_local $var$2)
      )
      (i32.const 0)
     )
    )
    (i32.const 528)
   )
   (drop
    (call $12)
   )
   (set_local $var$3
    (call $14
     (i32.const 76)
    )
   )
   (call $import$6
    (tee_local $var$4
     (call $15
      (call $17
       (get_local $var$6)
       (get_local $var$0)
      )
      (get_local $var$1)
     )
    )
   )
   (set_local $var$5
    (call $import$1
     (get_local $var$4)
     (call $16
      (get_local $var$4)
     )
     (get_local $var$3)
     (i32.const 76)
    )
   )
   (call $import$6
    (i32.add
     (get_local $var$3)
     (i32.const 36)
    )
   )
   (call $import$6
    (i32.add
     (get_local $var$3)
     (i32.const 46)
    )
   )
   (call $import$6
    (i32.add
     (get_local $var$3)
     (i32.const 56)
    )
   )
   (call $import$6
    (i32.add
     (get_local $var$3)
     (i32.const 66)
    )
   )
   (call $import$5
    (i64.load32_s
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=4
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=8
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=12
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=16
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=20
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=24
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=28
     (get_local $var$3)
    )
   )
   (call $import$5
    (i64.load32_s offset=32
     (get_local $var$3)
    )
   )
   (call $import$0
    (i32.ne
     (get_local $var$5)
     (i32.const 0)
    )
    (i32.const 432)
   )
   (call $import$5
    (i64.const 0)
   )
   (i32.store
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 1)
   )
   (i32.store offset=4
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 2)
   )
   (i32.store offset=8
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 3)
   )
   (i32.store offset=12
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 4)
   )
   (i32.store offset=16
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 5)
   )
   (i32.store offset=20
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 6)
   )
   (i32.store offset=24
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 7)
   )
   (i32.store offset=28
    (get_local $var$3)
    (i32.const 0)
   )
   (call $import$5
    (i64.const 8)
   )
   (i32.store offset=32
    (get_local $var$3)
    (i32.const 0)
   )
   (i64.store offset=56 align=1
    (get_local $var$3)
    (i64.const 0)
   )
   (i64.store offset=66 align=1
    (get_local $var$3)
    (i64.const 0)
   )
   (i32.store16 offset=64 align=1
    (get_local $var$3)
    (i32.const 0)
   )
   (i32.store16 offset=74 align=1
    (get_local $var$3)
    (i32.const 0)
   )
   (drop
    (call $import$2
     (get_local $var$4)
     (call $16
      (get_local $var$4)
     )
     (get_local $var$3)
     (i32.const 76)
    )
   )
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (call $19
        (get_local $var$2)
        (get_local $var$0)
       )
      )
     )
     (drop
      (call $5
       (i32.const 304)
       (get_local $var$0)
       (i32.const 592)
      )
     )
     (drop
      (call $5
       (i32.const 304)
       (get_local $var$1)
       (i32.const 576)
      )
     )
     (br $label$1)
    )
    (drop
     (call $5
      (i32.const 304)
      (get_local $var$0)
      (i32.const 576)
     )
    )
    (drop
     (call $5
      (i32.const 304)
      (get_local $var$1)
      (i32.const 592)
     )
    )
   )
   (call $13
    (get_local $var$3)
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$6)
     (i32.const 32)
    )
   )
   (i32.const 0)
  )
 )
 (func $10 (type $9) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (param $var$3 i32) (param $var$4 i32) (result i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (local $var$7 i32)
  (local $var$8 i32)
  (local $var$9 i32)
  (local $var$10 i32)
  (local $var$11 i32)
  (local $var$12 i32)
  (block $label$0 i32
   (i32.store offset=4
    (i32.const 0)
    (tee_local $var$12
     (i32.sub
      (i32.load offset=4
       (i32.const 0)
      )
      (i32.const 80)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$1)
     )
     (i32.const 10)
    )
    (i32.const 144)
   )
   (call $import$0
    (i32.gt_s
     (call $16
      (get_local $var$2)
     )
     (i32.const 10)
    )
    (i32.const 608)
   )
   (drop
    (call $12)
   )
   (set_local $var$5
    (call $14
     (i32.const 76)
    )
   )
   (set_local $var$11
    (call $import$1
     (tee_local $var$6
      (call $15
       (call $17
        (i32.add
         (get_local $var$12)
         (i32.const 32)
        )
        (get_local $var$0)
       )
       (get_local $var$1)
      )
     )
     (call $16
      (get_local $var$6)
     )
     (get_local $var$5)
     (i32.const 76)
    )
   )
   (call $import$6
    (tee_local $var$7
     (i32.add
      (get_local $var$5)
      (i32.const 36)
     )
    )
   )
   (call $import$6
    (tee_local $var$8
     (i32.add
      (get_local $var$5)
      (i32.const 46)
     )
    )
   )
   (call $import$6
    (tee_local $var$9
     (i32.add
      (get_local $var$5)
      (i32.const 56)
     )
    )
   )
   (call $import$6
    (tee_local $var$10
     (i32.add
      (get_local $var$5)
      (i32.const 66)
     )
    )
   )
   (call $import$5
    (i64.load32_s
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=4
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=8
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=12
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=16
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=20
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=24
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=28
     (get_local $var$5)
    )
   )
   (call $import$5
    (i64.load32_s offset=32
     (get_local $var$5)
    )
   )
   (call $import$0
    (i32.ne
     (get_local $var$11)
     (i32.const 0)
    )
    (i32.const 432)
   )
   (call $import$0
    (i32.eqz
     (i32.load8_u offset=56
      (get_local $var$5)
     )
    )
    (i32.const 656)
   )
   (call $import$0
    (i32.ne
     (call $19
      (get_local $var$9)
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 688)
   )
   (set_local $var$11
    (i32.const 1)
   )
   (block $label$1
    (br_if $label$1
     (i32.gt_u
      (i32.add
       (get_local $var$3)
       (i32.const -1)
      )
      (i32.const 2)
     )
    )
    (br_if $label$1
     (i32.gt_u
      (i32.add
       (get_local $var$4)
       (i32.const -1)
      )
      (i32.const 2)
     )
    )
    (set_local $var$11
     (i32.gt_s
      (i32.add
       (i32.add
        (i32.mul
         (get_local $var$3)
         (i32.const 3)
        )
        (get_local $var$4)
       )
       (i32.const -3)
      )
      (i32.const 9)
     )
    )
   )
   (call $import$0
    (get_local $var$11)
    (i32.const 720)
   )
   (block $label$2
    (block $label$3
     (br_if $label$3
      (i32.eqz
       (call $19
        (get_local $var$0)
        (get_local $var$2)
       )
      )
     )
     (drop
      (call $17
       (i32.add
        (get_local $var$12)
        (i32.const 70)
       )
       (get_local $var$0)
      )
     )
     (set_local $var$2
      (i32.const 2)
     )
     (br $label$2)
    )
    (drop
     (call $17
      (i32.add
       (get_local $var$12)
       (i32.const 70)
      )
      (get_local $var$1)
     )
    )
    (set_local $var$2
     (i32.const 1)
    )
   )
   (i32.store
    (i32.add
     (i32.add
      (get_local $var$5)
      (i32.shl
       (i32.add
        (i32.mul
         (get_local $var$3)
         (i32.const 3)
        )
        (get_local $var$4)
       )
       (i32.const 2)
      )
     )
     (i32.const -16)
    )
    (get_local $var$2)
   )
   (set_local $var$2
    (call $17
     (get_local $var$9)
     (i32.add
      (get_local $var$12)
      (i32.const 70)
     )
    )
   )
   (block $label$4
    (block $label$5
     (br_if $label$5
      (i32.ne
       (call $4
        (get_local $var$5)
        (i32.add
         (get_local $var$12)
         (i32.const 60)
        )
       )
       (i32.const 9)
      )
     )
     (br_if $label$5
      (i32.and
       (i32.load8_u offset=60
        (get_local $var$12)
       )
       (i32.const 255)
      )
     )
     (i64.store align=1
      (get_local $var$2)
      (i64.const 0)
     )
     (i32.store16 align=1
      (i32.add
       (get_local $var$2)
       (i32.const 8)
      )
      (i32.const 0)
     )
     (call $import$6
      (i32.const 752)
     )
     (drop
      (call $import$2
       (get_local $var$6)
       (call $16
        (get_local $var$6)
       )
       (get_local $var$5)
       (i32.const 76)
      )
     )
     (drop
      (call $5
       (i32.const 304)
       (get_local $var$0)
       (i32.const 320)
      )
     )
     (drop
      (call $5
       (i32.const 304)
       (get_local $var$1)
       (i32.const 320)
      )
     )
     (br $label$4)
    )
    (set_local $var$0
     (call $17
      (get_local $var$10)
      (i32.add
       (get_local $var$12)
       (i32.const 60)
      )
     )
    )
    (call $import$6
     (i32.add
      (get_local $var$12)
      (i32.const 60)
     )
    )
    (block $label$6
     (br_if $label$6
      (i32.eqz
       (i32.load8_u offset=60
        (get_local $var$12)
       )
      )
     )
     (i32.store8 offset=10
      (get_local $var$12)
      (i32.load8_u offset=810
       (i32.const 0)
      )
     )
     (i32.store16 offset=8
      (get_local $var$12)
      (i32.load16_u offset=808 align=1
       (i32.const 0)
      )
     )
     (i64.store
      (get_local $var$12)
      (i64.load offset=800 align=1
       (i32.const 0)
      )
     )
     (call $import$6
      (call $15
       (get_local $var$12)
       (i32.add
        (get_local $var$12)
        (i32.const 60)
       )
      )
     )
     (i32.store16 align=1
      (i32.add
       (get_local $var$2)
       (i32.const 8)
      )
      (i32.const 0)
     )
     (i64.store align=1
      (get_local $var$2)
      (i64.const 0)
     )
     (drop
      (call $import$2
       (get_local $var$6)
       (call $16
        (get_local $var$6)
       )
       (get_local $var$5)
       (i32.const 76)
      )
     )
     (drop
      (call $5
       (i32.const 304)
       (i32.add
        (get_local $var$12)
        (i32.const 60)
       )
       (i32.const 464)
      )
     )
     (br $label$4)
    )
    (drop
     (call $import$2
      (get_local $var$6)
      (call $16
       (get_local $var$6)
      )
      (get_local $var$5)
      (i32.const 76)
     )
    )
    (call $import$6
     (i32.const 816)
    )
    (call $import$6
     (get_local $var$7)
    )
    (call $import$6
     (get_local $var$8)
    )
    (call $import$6
     (get_local $var$2)
    )
    (call $import$6
     (get_local $var$0)
    )
    (call $import$5
     (i64.load32_s
      (get_local $var$5)
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 4)
      )
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 8)
      )
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 12)
      )
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 16)
      )
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 20)
      )
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 24)
      )
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 28)
      )
     )
    )
    (call $import$5
     (i64.load32_s
      (i32.add
       (get_local $var$5)
       (i32.const 32)
      )
     )
    )
    (call $13
     (get_local $var$5)
    )
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$12)
     (i32.const 80)
    )
   )
   (i32.const 0)
  )
 )
 (func $11 (type $6) (param $var$0 i32) (result i32)
  (block $label$0 i32
   (block $label$1
    (br_if $label$1
     (call $19
      (get_local $var$0)
      (i32.const 848)
     )
    )
    (drop
     (call $7
      (call $import$7
       (i32.const 1)
      )
      (call $import$7
       (i32.const 2)
      )
     )
    )
   )
   (block $label$2
    (br_if $label$2
     (call $19
      (get_local $var$0)
      (i32.const 864)
     )
    )
    (drop
     (call $9
      (call $import$7
       (i32.const 1)
      )
      (call $import$7
       (i32.const 2)
      )
      (call $import$7
       (i32.const 3)
      )
     )
    )
   )
   (block $label$3
    (br_if $label$3
     (call $19
      (get_local $var$0)
      (i32.const 880)
     )
    )
    (drop
     (call $8
      (call $import$7
       (i32.const 1)
      )
      (call $import$7
       (i32.const 2)
      )
      (call $import$7
       (i32.const 3)
      )
     )
    )
   )
   (block $label$4
    (br_if $label$4
     (i32.eqz
      (call $19
       (get_local $var$0)
       (i32.const 896)
      )
     )
    )
    (return
     (i32.const 0)
    )
   )
   (drop
    (call $10
     (call $import$7
      (i32.const 1)
     )
     (call $import$7
      (i32.const 2)
     )
     (call $import$7
      (i32.const 3)
     )
     (call $import$7
      (i32.const 4)
     )
     (call $import$7
      (i32.const 5)
     )
    )
   )
   (i32.const 0)
  )
 )
 (func $12 (type $4) (result i32)
  (local $var$0 i32)
  (block $label$0 i32
   (i32.store offset=912
    (i32.const 0)
    (i32.const 920)
   )
   (i64.store offset=904 align=4
    (i32.const 0)
    (i64.const 0)
   )
   (set_local $var$0
    (current_memory)
   )
   (i32.store offset=924
    (i32.const 0)
    (i32.const 932)
   )
   (i32.store offset=936
    (i32.const 0)
    (i32.const 944)
   )
   (i32.store offset=948
    (i32.const 0)
    (i32.const 956)
   )
   (i32.store offset=960
    (i32.const 0)
    (i32.const 968)
   )
   (i32.store offset=972
    (i32.const 0)
    (i32.const 980)
   )
   (i32.store offset=984
    (i32.const 0)
    (i32.const 992)
   )
   (i32.store offset=996
    (i32.const 0)
    (i32.const 1004)
   )
   (i32.store offset=1008
    (i32.const 0)
    (i32.const 1016)
   )
   (i32.store offset=1020
    (i32.const 0)
    (i32.const 1028)
   )
   (i32.store offset=1032
    (i32.const 0)
    (i32.const 1040)
   )
   (i32.store offset=1044
    (i32.const 0)
    (i32.const 1052)
   )
   (i32.store offset=1056
    (i32.const 0)
    (i32.const 1064)
   )
   (i32.store offset=1068
    (i32.const 0)
    (i32.const 1076)
   )
   (i32.store offset=916
    (i32.const 0)
    (i32.sub
     (i32.shl
      (get_local $var$0)
      (i32.const 16)
     )
     (i32.load offset=1688
      (i32.const 0)
     )
    )
   )
   (i32.store offset=1080
    (i32.const 0)
    (i32.const 1088)
   )
   (i32.store offset=1092
    (i32.const 0)
    (i32.const 1100)
   )
   (i32.store offset=1104
    (i32.const 0)
    (i32.const 1112)
   )
   (i32.store offset=1116
    (i32.const 0)
    (i32.const 1124)
   )
   (i32.store offset=1128
    (i32.const 0)
    (i32.const 1136)
   )
   (i32.store offset=1140
    (i32.const 0)
    (i32.const 1148)
   )
   (i32.store offset=1152
    (i32.const 0)
    (i32.const 1160)
   )
   (i32.store offset=1164
    (i32.const 0)
    (i32.const 1172)
   )
   (i32.store offset=1176
    (i32.const 0)
    (i32.const 1184)
   )
   (i32.store offset=1188
    (i32.const 0)
    (i32.const 1196)
   )
   (i32.store offset=1200
    (i32.const 0)
    (i32.const 1208)
   )
   (i32.store offset=1212
    (i32.const 0)
    (i32.const 1220)
   )
   (i32.store offset=1224
    (i32.const 0)
    (i32.const 1232)
   )
   (i32.store offset=1236
    (i32.const 0)
    (i32.const 1244)
   )
   (i32.store offset=1248
    (i32.const 0)
    (i32.const 1256)
   )
   (i32.store offset=1260
    (i32.const 0)
    (i32.const 1268)
   )
   (i32.store offset=1272
    (i32.const 0)
    (i32.const 1280)
   )
   (i32.store offset=1284
    (i32.const 0)
    (i32.const 1292)
   )
   (i32.store offset=1296
    (i32.const 0)
    (i32.const 1304)
   )
   (i32.store offset=1308
    (i32.const 0)
    (i32.const 1316)
   )
   (i32.store offset=1320
    (i32.const 0)
    (i32.const 1328)
   )
   (i32.store offset=1332
    (i32.const 0)
    (i32.const 1340)
   )
   (i32.store offset=1344
    (i32.const 0)
    (i32.const 1352)
   )
   (i32.store offset=1356
    (i32.const 0)
    (i32.const 1364)
   )
   (i32.store offset=1368
    (i32.const 0)
    (i32.const 1376)
   )
   (i32.store offset=1380
    (i32.const 0)
    (i32.const 1388)
   )
   (i32.store offset=1392
    (i32.const 0)
    (i32.const 1400)
   )
   (i32.store offset=1404
    (i32.const 0)
    (i32.const 1412)
   )
   (i32.store offset=1416
    (i32.const 0)
    (i32.const 1424)
   )
   (i32.store offset=1428
    (i32.const 0)
    (i32.const 1436)
   )
   (i32.store offset=1440
    (i32.const 0)
    (i32.const 1448)
   )
   (i32.store offset=1452
    (i32.const 0)
    (i32.const 1460)
   )
   (i32.store offset=1464
    (i32.const 0)
    (i32.const 1472)
   )
   (i32.store offset=1476
    (i32.const 0)
    (i32.const 1484)
   )
   (i32.store offset=1488
    (i32.const 0)
    (i32.const 1496)
   )
   (i32.store offset=1500
    (i32.const 0)
    (i32.const 1508)
   )
   (i32.store offset=1512
    (i32.const 0)
    (i32.const 1520)
   )
   (i32.store offset=1524
    (i32.const 0)
    (i32.const 1532)
   )
   (i32.store offset=1536
    (i32.const 0)
    (i32.const 1544)
   )
   (i32.store offset=1548
    (i32.const 0)
    (i32.const 1556)
   )
   (i32.store offset=1560
    (i32.const 0)
    (i32.const 1568)
   )
   (i32.store offset=1572
    (i32.const 0)
    (i32.const 1580)
   )
   (i32.store offset=1584
    (i32.const 0)
    (i32.const 1592)
   )
   (i32.store offset=1596
    (i32.const 0)
    (i32.const 1604)
   )
   (i32.store offset=1608
    (i32.const 0)
    (i32.const 1616)
   )
   (i32.store offset=1620
    (i32.const 0)
    (i32.const 1628)
   )
   (i32.store offset=1632
    (i32.const 0)
    (i32.const 1640)
   )
   (i32.store offset=1644
    (i32.const 0)
    (i32.const 1652)
   )
   (i32.store offset=1656
    (i32.const 0)
    (i32.const 1664)
   )
   (i32.store offset=1668
    (i32.const 0)
    (i32.const 1676)
   )
   (i32.const 0)
  )
 )
 (func $13 (type $1) (param $var$0 i32)
  (local $var$1 i32)
  (local $var$2 i32)
  (block $label$0
   (set_local $var$2
    (i32.const 0)
   )
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (tee_local $var$1
        (i32.load offset=908
         (i32.const 0)
        )
       )
      )
     )
     (loop $label$3
      (br_if $label$1
       (i32.eq
        (i32.load
         (tee_local $var$1
          (get_local $var$1)
         )
        )
        (get_local $var$0)
       )
      )
      (set_local $var$2
       (get_local $var$1)
      )
      (br_if $label$3
       (tee_local $var$1
        (i32.load offset=4
         (get_local $var$1)
        )
       )
      )
     )
    )
    (return)
   )
   (i32.store
    (select
     (i32.add
      (get_local $var$2)
      (i32.const 4)
     )
     (i32.const 908)
     (get_local $var$2)
    )
    (i32.load offset=4
     (get_local $var$1)
    )
   )
   (set_local $var$2
    (i32.load offset=904
     (i32.const 0)
    )
   )
   (i32.store offset=904
    (i32.const 0)
    (get_local $var$1)
   )
   (i32.store offset=4
    (get_local $var$1)
    (get_local $var$2)
   )
  )
 )
 (func $14 (type $6) (param $var$0 i32) (result i32)
  (local $var$1 i32)
  (local $var$2 i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (block $label$0 i32
   (set_local $var$2
    (i32.and
     (i32.add
      (get_local $var$0)
      (i32.const 7)
     )
     (i32.const -8)
    )
   )
   (set_local $var$4
    (i32.const 0)
   )
   (set_local $var$1
    (i32.load offset=916
     (i32.const 0)
    )
   )
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (tee_local $var$0
        (i32.load offset=904
         (i32.const 0)
        )
       )
      )
     )
     (loop $label$3
      (br_if $label$1
       (i32.ge_u
        (i32.load offset=8
         (tee_local $var$0
          (get_local $var$0)
         )
        )
        (get_local $var$2)
       )
      )
      (set_local $var$4
       (get_local $var$0)
      )
      (br_if $label$3
       (tee_local $var$0
        (i32.load offset=4
         (get_local $var$0)
        )
       )
      )
     )
    )
    (set_local $var$4
     (i32.const 0)
    )
    (block $label$4
     (br_if $label$4
      (i32.lt_s
       (get_local $var$2)
       (i32.const 0)
      )
     )
     (set_local $var$0
      (i32.load offset=1688
       (i32.const 0)
      )
     )
     (set_local $var$3
      (current_memory)
     )
     (block $label$5
      (br_if $label$5
       (i32.ge_u
        (get_local $var$0)
        (get_local $var$2)
       )
      )
      (br_if $label$4
       (i32.eqz
        (grow_memory
         (i32.add
          (i32.shr_u
           (i32.sub
            (i32.add
             (get_local $var$2)
             (i32.const -1)
            )
            (get_local $var$0)
           )
           (i32.const 16)
          )
          (i32.const 1)
         )
        )
       )
      )
      (i32.store offset=1688
       (i32.const 0)
       (tee_local $var$0
        (i32.add
         (i32.shl
          (i32.sub
           (current_memory)
           (get_local $var$3)
          )
          (i32.const 16)
         )
         (i32.load offset=1688
          (i32.const 0)
         )
        )
       )
      )
     )
     (i32.store offset=1688
      (i32.const 0)
      (i32.sub
       (get_local $var$0)
       (get_local $var$2)
      )
     )
    )
    (block $label$6
     (br_if $label$6
      (i32.gt_u
       (tee_local $var$3
        (i32.add
         (get_local $var$1)
         (get_local $var$2)
        )
       )
       (i32.const 16777216)
      )
     )
     (br_if $label$6
      (i32.eqz
       (tee_local $var$0
        (i32.load offset=912
         (i32.const 0)
        )
       )
      )
     )
     (i32.store
      (get_local $var$0)
      (get_local $var$1)
     )
     (i32.store offset=8
      (get_local $var$0)
      (get_local $var$2)
     )
     (i32.store offset=916
      (i32.const 0)
      (get_local $var$3)
     )
     (i32.store offset=912
      (i32.const 0)
      (i32.load offset=4
       (get_local $var$0)
      )
     )
     (i32.store offset=4
      (get_local $var$0)
      (i32.load offset=908
       (i32.const 0)
      )
     )
     (i32.store offset=908
      (i32.const 0)
      (get_local $var$0)
     )
     (set_local $var$4
      (get_local $var$1)
     )
    )
    (return
     (get_local $var$4)
    )
   )
   (i32.store
    (select
     (i32.add
      (get_local $var$4)
      (i32.const 4)
     )
     (i32.const 904)
     (get_local $var$4)
    )
    (i32.load offset=4
     (get_local $var$0)
    )
   )
   (i32.store offset=4
    (get_local $var$0)
    (i32.load offset=908
     (i32.const 0)
    )
   )
   (i32.store offset=908
    (i32.const 0)
    (get_local $var$0)
   )
   (i32.load
    (get_local $var$0)
   )
  )
 )
 (func $15 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (drop
    (call $17
     (i32.add
      (get_local $var$0)
      (call $16
       (get_local $var$0)
      )
     )
     (get_local $var$1)
    )
   )
   (get_local $var$0)
  )
 )
 (func $16 (type $6) (param $var$0 i32) (result i32)
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
 (func $17 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (drop
    (call $18
     (get_local $var$0)
     (get_local $var$1)
    )
   )
   (get_local $var$0)
  )
 )
 (func $18 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
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
 (func $19 (type $7) (param $var$0 i32) (param $var$1 i32) (result i32)
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
)

