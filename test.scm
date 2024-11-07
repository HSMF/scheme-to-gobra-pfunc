(module gobra (slice seq len ++ args requires preserves ensures returns =seq && elem)
  (import scheme)

  (define (take seq n)
    (if (or (= n 0) (null? seq))
      '()
      (cons (car seq) (take (cdr seq) (- n 1)))
    )
  )

  (define (drop seq n)
    (cond
      ((= n 0) seq )
      ((null? seq) '())
      (else (drop (cdr seq) (- n 1)))
    )
  )

  (define (slice seq low high)
    (take (drop seq low)  (- high low)
    )
  )

  (define (seq . args-list) (drop args-list 1) )

  (define (len s)  (if (null? s) 0 (+ (len (cdr s)) 1)))

  (define (++ a b)  (if (null? a)
                      b
                      (cons (car a) (++ (cdr a) b) )
                      ))

  (define (elem s n) (cond
                       ((null? s) 0)
                       ((= n 0) (car s))
                       (else (elem (cdr s) (- n 1) ))
                       ))

  (define (requires _) '())
  (define (preserves _) '())
  (define (ensures _) '())
  (define (returns _) '())

  (define-syntax args
    (syntax-rules () ((args . _)  '() ) ))

  (define (=seq a b) (cond
                       ((and (null? a) (null? b)) #t)
                       ((null? a) #f)
                       ((null? b) #f)
                       ((not (= (car a) (car b))) #f)
                       (else (=seq (cdr a) (cdr b)))
   ))


  (define (&& a b) (cond
                     ((not a) #f)
                     ((not b) #f)
                     (else #t)
                     ))
)

(import gobra)

(print (slice (seq "byte" 1 2 3 4 5) 2 4 ))

(print (++ (seq "byte" 1 2 3) (seq "byte" 4 5 6)))

(define (repeat s n)
  (begin
    (args (s "seq[byte]") (n "int") )
    (returns "seq[byte]")
    (requires (>= n 0))
    (print n)
    (if (= n 0)
      (seq "byte")
      (++ (repeat s (- n 1)) s)
    )
  )
)

(define (Split s sep)
  (begin
    (args (s "seq[byte]") (sep "seq[byte]"))
    (returns "seq[seq[byte]]")
    (letrec
      ((aux (lambda (s ac)
              (begin
                (args (s "seq[byte]") (ac "seq[byte]") )
                (cond
                  ((&& (null? s) (null? ac)) (seq "byte"))
                  ((null? s) (seq "seq[byte]" ac))
                  ((=seq s sep) (seq "seq[byte]" ac (seq "byte")))
                  ((=seq (slice s 0 (len sep)) sep) (++ (seq "seq[byte]" ac) (aux (slice s (len sep) (len s)) (seq "byte") )))
                  (else (aux (slice s 1 (len s)) (++ ac (seq "byte" (elem s 0)))))
                )))))
      (aux s (seq "byte")) )
  )
)

(print (Split (seq "byte" 3 4) (seq "byte" 3 4)))
(print (Split (seq "byte" 3 4 5) (seq "byte" 3 4)))
(print (Split (seq "byte" 1 2 3 4 5 6 3 4 7) (seq "byte" 3 4)))
(print (Split (seq "byte" 3 4 1 2 3 4 5 6 3 4 7) (seq "byte" 3 4)))
(print (Split (seq "byte" 3 4 3) (seq "byte" 3 4)))
(print (Split (seq "byte" 1 2 3 5) (seq "byte" 3 4)))
