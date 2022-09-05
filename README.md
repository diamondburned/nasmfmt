# nasmfmt2

Automatically format your assembly sources with a simple command.

nasmfmt2 is a rewrite of [yamnikov-oleg/nasmfmt][nasmfmt]. The rewrite features
an overall better parser that is also reusable as a library.

The main reason for the rewrite is for ease of maintenance and for
style-compatibility with [Professor Floyd Holliday's CPSC-240 Assembly
class][x86-programming].

[nasmfmt]: https://github.com/yamnikov-oleg/nasmfmt
[x86-programming]: https://sites.google.com/a/fullerton.edu/activeprofessor/4-subjects/x86-programming?authuser=0

Inspired by gofmt.

## Example

```asm
lobal _start


section .text

   ;Starting point
_start:
mov rax,1 ;write(fd, buf, len)
mov rdi,1  ; fd
mov rsi, msg   ; buf
mov rdx,  msglen; len
  syscall

mov rax,60 ;exit(status)
mov rdi, 0
  syscall

section .data
msg    db "Hello world!",10
msglen equ $-msg
```

turns into

```asm
global _start

section .text

; Starting point
_start:
        mov     rax, 1                 ; write(fd, buf, len)
        mov     rdi, 1                 ; fd
        mov     rsi, msg               ; buf
        mov     rdx, msglen            ; len
        syscall 

        mov     rax, 60                ; exit(status)
        mov     rdi, 0
        syscall 

section .data

msg db "Hello world!",10

msglen equ $-msg
```

## Installing

Requires Go 1.18+.

```go
go install github.com/diamondburned/nasmfmt/v2@latest
```

## Vim + ALE integration

```vim
autocmd BufRead,BufNewFile *.asm    set filetype=nasm

function! FixNasmfmt(buffer) abort
    return {
    \   'command': 'nasmfmt -'
    \}
endfunction

execute ale#fix#registry#Add('nasmfmt', 'FixNasmfmt', ['nasm'], 'nasmfmt')

let g:ale_fixers = {
	\ 'nasm': [ "nasmfmt" ],
	\ }
```
