(module
 (type $0 (func (param i32)))
 (type $1 (func (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32) (result i32)))
 (type $2 (func (param i32 i32)))
 (type $3 (func (result i32)))
 (type $4 (func (param i32 i32 i32 i32) (result i32)))
 (type $5 (func (param i32) (result i32)))
 (type $6 (func (param i32 i32) (result i32)))
 (type $7 (func (param i32 i32 i32) (result i32)))
 (type $8 (func (param i32 i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_assert" (func $import$0 (param i32 i32)))
 (import "env" "ABA_db_get" (func $import$1 (param i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_db_put" (func $import$2 (param i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_inline_action" (func $import$3 (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32) (result i32)))
 (import "env" "ABA_is_account" (func $import$4 (param i32 i32) (result i32)))
 (import "env" "ABA_prints" (func $import$5 (param i32)))
 (import "env" "ABA_read_param" (func $import$6 (param i32) (result i32)))
 (import "env" "memset" (func $import$7 (param i32 i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 4) "\90S\00\00")
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
 (data (i32.const 304) "the game has exit\00")
 (data (i32.const 336) "tictactoe\00")
 (data (i32.const 352) "2\00")
 (data (i32.const 368) "restarter name is too long or is null\00")
 (data (i32.const 416) "restarter has insufficient permission\00")
 (data (i32.const 464) "the game does not exsit\00")
 (data (i32.const 496) "4\00")
 (data (i32.const 512) "closer name is too long or is null\00")
 (data (i32.const 560) "closer has insufficient permission\00")
 (data (i32.const 608) "the game is over\00")
 (data (i32.const 640) "1\00")
 (data (i32.const 656) "3\00")
 (data (i32.const 672) "host name is too long or is null\00")
 (data (i32.const 720) "it is not your turn to move\00")
 (data (i32.const 752) "it is not a valid movement\00")
 (data (i32.const 784) "the game is over and none of player win\00")
 (data (i32.const 832) "winner is \00")
 (data (i32.const 848) "create\00")
 (data (i32.const 864) "close\00")
 (data (i32.const 880) "restart\00")
 (data (i32.const 896) "follow\00")
 (export "memory" (memory $0))
 (export "apply" (func $11))
 (func $0 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (block $label$1
    (br_if $label$1
     (i32.lt_s
      (get_local $var$1)
      (i32.const 1)
     )
    )
    (drop
     (call $import$7
      (get_local $var$0)
      (i32.const 0)
      (i32.shl
       (get_local $var$1)
       (i32.const 2)
      )
     )
    )
   )
   (i32.const 0)
  )
 )
 (func $1 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (i64.store align=4
    (get_local $var$0)
    (i64.const 0)
   )
   (i32.store
    (i32.add
     (get_local $var$0)
     (i32.const 8)
    )
    (i32.const 0)
   )
   (i32.store offset=56 align=1
    (get_local $var$0)
    (i32.const 0)
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
   (drop
    (call $14
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
 (func $2 (type $5) (param $var$0 i32) (result i32)
  (select
   (i32.const -1)
   (i32.const 1)
   (get_local $var$0)
  )
 )
 (func $3 (type $7) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
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
 (func $4 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
  (local $var$2 i32)
  (local $var$3 i32)
  (local $var$4 i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (local $var$7 i32)
  (block $label$0 i32
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
      (call $14
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
    (call $14
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
 (func $5 (type $7) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
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
     (call $13
      (tee_local $var$3
       (call $12
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
     (call $13
      (tee_local $var$1
       (call $12
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
      (call $13
       (tee_local $var$2
        (call $12
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
   (call $import$5
    (get_local $var$2)
   )
   (drop
    (call $import$3
     (i32.const 48)
     (i32.const 8)
     (i32.const 64)
     (i32.const 8)
     (get_local $var$2)
     (call $13
      (get_local $var$2)
     )
     (get_local $var$0)
     (call $13
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
 (func $6 (type $5) (param $var$0 i32) (result i32)
  (select
   (i32.const -1)
   (i32.const 0)
   (i32.gt_s
    (call $13
     (get_local $var$0)
    )
    (i32.const 10)
   )
  )
 )
 (func $7 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
  (local $var$2 i32)
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
      (i32.const 112)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $13
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
      (call $13
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
      (call $13
       (get_local $var$1)
      )
     )
     (i32.const 0)
    )
    (i32.const 224)
   )
   (call $import$0
    (i32.eqz
     (call $16
      (get_local $var$0)
      (get_local $var$1)
     )
    )
    (i32.const 256)
   )
   (set_local $var$3
    (i32.add
     (get_local $var$4)
     (i32.const 88)
    )
   )
   (block $label$1
    (br_if $label$1
     (call $import$1
      (tee_local $var$2
       (call $12
        (call $14
         (get_local $var$4)
         (get_local $var$0)
        )
        (get_local $var$1)
       )
      )
      (call $13
       (get_local $var$2)
      )
      (i32.add
       (get_local $var$4)
       (i32.const 32)
      )
      (i32.const 76)
     )
    )
    (call $import$0
     (i32.ne
      (i32.load8_u
       (get_local $var$3)
      )
      (i32.const 0)
     )
     (i32.const 304)
    )
   )
   (drop
    (call $import$7
     (i32.add
      (get_local $var$4)
      (i32.const 32)
     )
     (i32.const 0)
     (i32.const 36)
    )
   )
   (drop
    (call $14
     (i32.add
      (i32.add
       (get_local $var$4)
       (i32.const 32)
      )
      (i32.const 36)
     )
     (get_local $var$0)
    )
   )
   (drop
    (call $14
     (i32.add
      (get_local $var$4)
      (i32.const 78)
     )
     (get_local $var$1)
    )
   )
   (drop
    (call $14
     (get_local $var$3)
     (get_local $var$0)
    )
   )
   (i32.store16
    (i32.add
     (get_local $var$4)
     (i32.const 106)
    )
    (i32.const 0)
   )
   (i64.store offset=98 align=2
    (get_local $var$4)
    (i64.const 0)
   )
   (drop
    (call $5
     (get_local $var$0)
     (i32.const 336)
     (i32.const 352)
    )
   )
   (drop
    (call $5
     (get_local $var$1)
     (i32.const 336)
     (i32.const 352)
    )
   )
   (drop
    (call $import$2
     (get_local $var$2)
     (call $13
      (get_local $var$2)
     )
     (i32.add
      (get_local $var$4)
      (i32.const 32)
     )
     (i32.const 76)
    )
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$4)
     (i32.const 112)
    )
   )
   (i32.const 0)
  )
 )
 (func $8 (type $7) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
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
      (i32.const 112)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$1)
     )
     (i32.const 10)
    )
    (i32.const 144)
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$2)
     )
     (i32.const 10)
    )
    (i32.const 368)
   )
   (call $import$0
    (i32.and
     (i32.ne
      (call $16
       (get_local $var$0)
       (get_local $var$2)
      )
      (i32.const 0)
     )
     (i32.ne
      (call $16
       (get_local $var$1)
       (get_local $var$2)
      )
      (i32.const 0)
     )
    )
    (i32.const 416)
   )
   (call $import$0
    (i32.ne
     (call $import$1
      (tee_local $var$3
       (call $12
        (call $14
         (get_local $var$4)
         (get_local $var$0)
        )
        (get_local $var$1)
       )
      )
      (call $13
       (get_local $var$3)
      )
      (i32.add
       (get_local $var$4)
       (i32.const 32)
      )
      (i32.const 76)
     )
     (i32.const 0)
    )
    (i32.const 464)
   )
   (i32.store
    (i32.add
     (get_local $var$4)
     (i32.const 40)
    )
    (i32.const 0)
   )
   (i64.store align=4
    (i32.add
     (get_local $var$4)
     (i32.const 100)
    )
    (i64.const 0)
   )
   (i64.store align=4
    (i32.add
     (get_local $var$4)
     (i32.const 92)
    )
    (i64.const 0)
   )
   (i64.store offset=32
    (get_local $var$4)
    (i64.const 0)
   )
   (i32.store offset=88
    (get_local $var$4)
    (i32.const 0)
   )
   (drop
    (call $14
     (i32.add
      (get_local $var$4)
      (i32.const 88)
     )
     (get_local $var$2)
    )
   )
   (drop
    (call $import$2
     (get_local $var$3)
     (call $13
      (get_local $var$3)
     )
     (i32.add
      (get_local $var$4)
      (i32.const 32)
     )
     (i32.const 76)
    )
   )
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (call $16
        (get_local $var$2)
        (get_local $var$0)
       )
      )
     )
     (drop
      (call $5
       (get_local $var$1)
       (i32.const 336)
       (i32.const 496)
      )
     )
     (br $label$1)
    )
    (drop
     (call $5
      (get_local $var$0)
      (i32.const 336)
      (i32.const 496)
     )
    )
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$4)
     (i32.const 112)
    )
   )
   (i32.const 0)
  )
 )
 (func $9 (type $7) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (result i32)
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
      (i32.const 112)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$1)
     )
     (i32.const 10)
    )
    (i32.const 144)
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$2)
     )
     (i32.const 10)
    )
    (i32.const 512)
   )
   (call $import$0
    (i32.and
     (i32.ne
      (call $16
       (get_local $var$0)
       (get_local $var$2)
      )
      (i32.const 0)
     )
     (i32.ne
      (call $16
       (get_local $var$1)
       (get_local $var$2)
      )
      (i32.const 0)
     )
    )
    (i32.const 560)
   )
   (call $import$0
    (i32.ne
     (call $import$1
      (tee_local $var$3
       (call $12
        (call $14
         (get_local $var$4)
         (get_local $var$0)
        )
        (get_local $var$1)
       )
      )
      (call $13
       (get_local $var$3)
      )
      (i32.add
       (get_local $var$4)
       (i32.const 32)
      )
      (i32.const 76)
     )
     (i32.const 0)
    )
    (i32.const 464)
   )
   (call $import$0
    (i32.eqz
     (i32.load8_u offset=88
      (get_local $var$4)
     )
    )
    (i32.const 608)
   )
   (drop
    (call $import$7
     (i32.add
      (get_local $var$4)
      (i32.const 32)
     )
     (i32.const 0)
     (i32.const 36)
    )
   )
   (i32.store
    (i32.add
     (get_local $var$4)
     (i32.const 104)
    )
    (i32.const 0)
   )
   (i64.store
    (i32.add
     (get_local $var$4)
     (i32.const 96)
    )
    (i64.const 0)
   )
   (i64.store offset=88
    (get_local $var$4)
    (i64.const 0)
   )
   (drop
    (call $import$2
     (get_local $var$3)
     (call $13
      (get_local $var$3)
     )
     (i32.add
      (get_local $var$4)
      (i32.const 32)
     )
     (i32.const 76)
    )
   )
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (call $16
        (get_local $var$2)
        (get_local $var$0)
       )
      )
     )
     (drop
      (call $5
       (i32.const 336)
       (get_local $var$0)
       (i32.const 656)
      )
     )
     (drop
      (call $5
       (i32.const 336)
       (get_local $var$1)
       (i32.const 640)
      )
     )
     (br $label$1)
    )
    (drop
     (call $5
      (i32.const 336)
      (get_local $var$0)
      (i32.const 640)
     )
    )
    (drop
     (call $5
      (i32.const 336)
      (get_local $var$1)
      (i32.const 656)
     )
    )
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$4)
     (i32.const 112)
    )
   )
   (i32.const 0)
  )
 )
 (func $10 (type $8) (param $var$0 i32) (param $var$1 i32) (param $var$2 i32) (param $var$3 i32) (param $var$4 i32) (result i32)
  (local $var$5 i32)
  (local $var$6 i32)
  (local $var$7 i32)
  (local $var$8 i32)
  (block $label$0 i32
   (i32.store offset=4
    (i32.const 0)
    (tee_local $var$8
     (i32.sub
      (i32.load offset=4
       (i32.const 0)
      )
      (i32.const 160)
     )
    )
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$0)
     )
     (i32.const 10)
    )
    (i32.const 96)
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$1)
     )
     (i32.const 10)
    )
    (i32.const 144)
   )
   (call $import$0
    (i32.gt_s
     (call $13
      (get_local $var$2)
     )
     (i32.const 10)
    )
    (i32.const 672)
   )
   (call $import$0
    (i32.ne
     (call $import$1
      (tee_local $var$5
       (call $12
        (call $14
         (i32.add
          (get_local $var$8)
          (i32.const 32)
         )
         (get_local $var$0)
        )
        (get_local $var$1)
       )
      )
      (call $13
       (get_local $var$5)
      )
      (i32.add
       (get_local $var$8)
       (i32.const 80)
      )
      (i32.const 76)
     )
     (i32.const 0)
    )
    (i32.const 464)
   )
   (call $import$0
    (i32.eqz
     (i32.load8_u offset=136
      (get_local $var$8)
     )
    )
    (i32.const 608)
   )
   (call $import$0
    (i32.ne
     (call $16
      (tee_local $var$6
       (i32.add
        (get_local $var$8)
        (i32.const 136)
       )
      )
      (get_local $var$2)
     )
     (i32.const 0)
    )
    (i32.const 720)
   )
   (set_local $var$7
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
    (set_local $var$7
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
    (get_local $var$7)
    (i32.const 752)
   )
   (block $label$2
    (block $label$3
     (br_if $label$3
      (i32.eqz
       (call $16
        (get_local $var$0)
        (get_local $var$2)
       )
      )
     )
     (drop
      (call $14
       (i32.add
        (get_local $var$8)
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
     (call $14
      (i32.add
       (get_local $var$8)
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
      (i32.add
       (get_local $var$8)
       (i32.const 80)
      )
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
    (call $14
     (get_local $var$6)
     (i32.add
      (get_local $var$8)
      (i32.const 70)
     )
    )
   )
   (block $label$4
    (block $label$5
     (br_if $label$5
      (i32.ne
       (call $4
        (i32.add
         (get_local $var$8)
         (i32.const 80)
        )
        (i32.add
         (get_local $var$8)
         (i32.const 60)
        )
       )
       (i32.const 9)
      )
     )
     (br_if $label$5
      (i32.and
       (i32.load8_u offset=60
        (get_local $var$8)
       )
       (i32.const 255)
      )
     )
     (i64.store align=4
      (get_local $var$2)
      (i64.const 0)
     )
     (i32.store16
      (i32.add
       (get_local $var$2)
       (i32.const 8)
      )
      (i32.const 0)
     )
     (call $import$5
      (i32.const 784)
     )
     (drop
      (call $import$2
       (get_local $var$5)
       (call $13
        (get_local $var$5)
       )
       (i32.add
        (get_local $var$8)
        (i32.const 80)
       )
       (i32.const 76)
      )
     )
     (drop
      (call $5
       (i32.const 336)
       (get_local $var$0)
       (i32.const 352)
      )
     )
     (drop
      (call $5
       (i32.const 336)
       (get_local $var$1)
       (i32.const 352)
      )
     )
     (br $label$4)
    )
    (drop
     (call $14
      (i32.add
       (get_local $var$8)
       (i32.const 146)
      )
      (i32.add
       (get_local $var$8)
       (i32.const 60)
      )
     )
    )
    (block $label$6
     (br_if $label$6
      (i32.eqz
       (i32.load8_u offset=60
        (get_local $var$8)
       )
      )
     )
     (i32.store8 offset=10
      (get_local $var$8)
      (i32.load8_u offset=842
       (i32.const 0)
      )
     )
     (i32.store16 offset=8
      (get_local $var$8)
      (i32.load16_u offset=840 align=1
       (i32.const 0)
      )
     )
     (i64.store
      (get_local $var$8)
      (i64.load offset=832 align=1
       (i32.const 0)
      )
     )
     (call $import$5
      (call $12
       (get_local $var$8)
       (i32.add
        (get_local $var$8)
        (i32.const 60)
       )
      )
     )
     (i32.store16
      (i32.add
       (get_local $var$2)
       (i32.const 8)
      )
      (i32.const 0)
     )
     (i64.store align=4
      (get_local $var$2)
      (i64.const 0)
     )
     (drop
      (call $import$2
       (get_local $var$5)
       (call $13
        (get_local $var$5)
       )
       (i32.add
        (get_local $var$8)
        (i32.const 80)
       )
       (i32.const 76)
      )
     )
     (drop
      (call $5
       (i32.const 336)
       (i32.add
        (get_local $var$8)
        (i32.const 60)
       )
       (i32.const 496)
      )
     )
     (br $label$4)
    )
    (drop
     (call $import$2
      (get_local $var$5)
      (call $13
       (get_local $var$5)
      )
      (i32.add
       (get_local $var$8)
       (i32.const 80)
      )
      (i32.const 76)
     )
    )
   )
   (i32.store offset=4
    (i32.const 0)
    (i32.add
     (get_local $var$8)
     (i32.const 160)
    )
   )
   (i32.const 0)
  )
 )
 (func $11 (type $5) (param $var$0 i32) (result i32)
  (block $label$0 i32
   (block $label$1
    (br_if $label$1
     (call $16
      (get_local $var$0)
      (i32.const 848)
     )
    )
    (drop
     (call $7
      (call $import$6
       (i32.const 1)
      )
      (call $import$6
       (i32.const 2)
      )
     )
    )
   )
   (block $label$2
    (br_if $label$2
     (call $16
      (get_local $var$0)
      (i32.const 864)
     )
    )
    (drop
     (call $9
      (call $import$6
       (i32.const 1)
      )
      (call $import$6
       (i32.const 2)
      )
      (call $import$6
       (i32.const 3)
      )
     )
    )
   )
   (block $label$3
    (br_if $label$3
     (call $16
      (get_local $var$0)
      (i32.const 880)
     )
    )
    (drop
     (call $8
      (call $import$6
       (i32.const 1)
      )
      (call $import$6
       (i32.const 2)
      )
      (call $import$6
       (i32.const 3)
      )
     )
    )
   )
   (block $label$4
    (br_if $label$4
     (i32.eqz
      (call $16
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
     (call $import$6
      (i32.const 1)
     )
     (call $import$6
      (i32.const 2)
     )
     (call $import$6
      (i32.const 3)
     )
     (call $import$6
      (i32.const 4)
     )
     (call $import$6
      (i32.const 5)
     )
    )
   )
   (i32.const 0)
  )
 )
 (func $12 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (drop
    (call $14
     (i32.add
      (get_local $var$0)
      (call $13
       (get_local $var$0)
      )
     )
     (get_local $var$1)
    )
   )
   (get_local $var$0)
  )
 )
 (func $13 (type $5) (param $var$0 i32) (result i32)
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
 (func $14 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
  (block $label$0 i32
   (drop
    (call $15
     (get_local $var$0)
     (get_local $var$1)
    )
   )
   (get_local $var$0)
  )
 )
 (func $15 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
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
 (func $16 (type $6) (param $var$0 i32) (param $var$1 i32) (result i32)
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

