package main

import (
 "context"
 "errors"
 "fmt"
 "os/exec"
 "strings"
 "time"
 "unicode"

 "github.com/charmbracelet/x/ansi"
)

const maxProcessOutput = 32*1024

// exec.Cmd serializes writes when stdout and stderr use the same writer.
// Both the pending line and the retained output are bounded, even with no LF.
type processOutput struct {
 tail []byte
 line []byte
 status func(string)
}
func (w *processOutput) Write(data []byte)(int,error){
 n:=len(data)
 w.tail=append(w.tail,data...)
 if len(w.tail)>maxProcessOutput {w.tail=append([]byte(nil),w.tail[len(w.tail)-maxProcessOutput:]...)}
 for _,b:=range data {
  if b=='\n'||b=='\r' {w.flush()} else {
   if len(w.line)>=4096 {w.flush()}
   w.line=append(w.line,b)
  }
 }
 return n,nil
}
func (w *processOutput) flush(){
 line:=safeText(string(w.line));w.line=w.line[:0]
 if line!=""&&w.status!=nil {w.status(line)}
}
func safeText(s string)string {
 return strings.TrimSpace(strings.Map(func(r rune)rune {if unicode.IsControl(r)&&r!='\n'&&r!='\t'{return -1};return r},ansi.Strip(s)))
}
func runProcess(ctx context.Context,status func(string),name string,args ...string)(string,error){
 output:=&processOutput{status:status}
 cmd:=exec.CommandContext(ctx,name,args...)
 configureProcess(cmd)
 cmd.Stdout=output;cmd.Stderr=output
 // No terminal input is forwarded: abcde receives EOF rather than competing
 // with Bubble Tea. Interactive abcde configurations need unattended settings.
 cmd.WaitDelay=2*time.Second
 err:=cmd.Run()
 output.flush()
 if ctx.Err()!=nil {err=errors.Join(ctx.Err(),err)}
 if err!=nil {err=fmt.Errorf("%s: %w",name,err)}
 return safeText(string(output.tail)),err
}
