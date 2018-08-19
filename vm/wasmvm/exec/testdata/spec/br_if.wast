(module
  (func $dummy)

  (func (export "as-block-first") (param i32) (result i32)
    (block (br_if 0 (get_local 0)) (return (i32.const 2))) (i32.const 3)
  )
  (func (export "as-block-mid") (param i32) (result i32)
    (block (call $dummy) (br_if 0 (get_local 0)) (return (i32.const 2)))
    (i32.const 3)
  )
  (func (export "as-block-last") (param i32)
    (block (call $dummy) (call $dummy) (br_if 0 (get_local 0)))
  )
  (func (export "as-block-first-value") (param i32) (result i32)
    (block (result i32)
      (drop (br_if 0 (i32.const 10) (get_local 0))) (return (i32.const 11))
    )
  )
  (func (export "as-block-mid-value") (param i32) (result i32)
    (block (result i32)
      (call $dummy)
      (drop (br_if 0 (i32.const 20) (get_local 0)))
      (return (i32.const 21))
    )
  )
  (func (export "as-block-last-value") (param i32) (result i32)
    (block (result i32)
      (call $dummy) (call $dummy) (br_if 0 (i32.const 11) (get_local 0))
    )
  )

  (func (export "as-loop-first") (param i32) (result i32)
    (block (loop (br_if 1 (get_local 0)) (return (i32.const 2)))) (i32.const 3)
  )
  (func (export "as-loop-mid") (param i32) (result i32)
    (block (loop (call $dummy) (br_if 1 (get_local 0)) (return (i32.const 2))))
    (i32.const 4)
  )
  (func (export "as-loop-last") (param i32)
    (loop (call $dummy) (br_if 1 (get_local 0)))
  )

  (func (export "as-if-then") (param i32 i32)
    (block
      (if (get_local 0) (then (br_if 1 (get_local 1))) (else (call $dummy)))
    )
  )
  (func (export "as-if-else") (param i32 i32)
    (block
      (if (get_local 0) (then (call $dummy)) (else (br_if 1 (get_local 1))))
    )
  )

  (func (export "nested-block-value") (param i32) (result i32)
    (i32.add
      (i32.const 1)
      (block (result i32)
        (drop (i32.const 2))
        (i32.add
          (i32.const 4)
          (block (result i32)
            (drop (br_if 1 (i32.const 8) (get_local 0)))
            (i32.const 16)
          )
        )
      )
    )
  )

  (func (export "nested-br-value") (param i32) (result i32)
    (i32.add
      (i32.const 1)
      (block (result i32)
        (drop (i32.const 2))
        (br 0
          (block (result i32)
            (drop (br_if 1 (i32.const 8) (get_local 0))) (i32.const 4)
          )
        )
        (i32.const 16)
      )
    )
  )

  (func (export "nested-br_if-value") (param i32) (result i32)
    (i32.add
      (i32.const 1)
      (block (result i32)
        (drop (i32.const 2))
        (drop (br_if 0
          (block (result i32)
            (drop (br_if 1 (i32.const 8) (get_local 0))) (i32.const 4)
          )
          (i32.const 1)
        ))
        (i32.const 16)
      )
    )
  )

  (func (export "nested-br_if-value-cond") (param i32) (result i32)
    (i32.add
      (i32.const 1)
      (block (result i32)
        (drop (i32.const 2))
        (drop (br_if 0
          (i32.const 4)
          (block (result i32)
            (drop (br_if 1 (i32.const 8) (get_local 0))) (i32.const 1)
          )
        ))
        (i32.const 16)
      )
    )
  )

  (func (export "nested-br_table-value") (param i32) (result i32)
    (i32.add
      (i32.const 1)
      (block (result i32)
        (drop (i32.const 2))
        (br_table 0
          (block (result i32)
            (drop (br_if 1 (i32.const 8) (get_local 0))) (i32.const 4)
          )
          (i32.const 1)
        )
        (i32.const 16)
      )
    )
  )

  (func (export "nested-br_table-value-index") (param i32) (result i32)
    (i32.add
      (i32.const 1)
      (block (result i32)
        (drop (i32.const 2))
        (br_table 0
          (i32.const 4)
          (block (result i32)
            (drop (br_if 1 (i32.const 8) (get_local 0))) (i32.const 1)
          )
        )
        (i32.const 16)
      )
    )
  )
)