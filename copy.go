package main

import (
 "context"
 "errors"
 "fmt"
 "io"
 "io/fs"
 "os"
 "path/filepath"
 "strings"
)

// copyAlbum stages the album beside its final destination. Only a complete
// copy is published; cancellation/errors remove the staging directory.
// Symlinks and special files are rejected rather than followed off the album.
func copyAlbum(ctx context.Context, source, target string, emit func(operationProgress)) (err error) {
 source, err = filepath.EvalSymlinks(source)
 if err != nil { return fmt.Errorf("Quelle prüfen: %w",err) }
 source, err = filepath.Abs(source)
 if err != nil { return err }
 info, err := os.Stat(source)
 if err != nil { return err }
 if !info.IsDir() { return fmt.Errorf("Quelle ist kein Verzeichnis: %s",source) }
 target, err = filepath.Abs(target)
 if err != nil { return err }
 parent, err := filepath.EvalSymlinks(filepath.Dir(target))
 if err != nil { return fmt.Errorf("Zielordner prüfen: %w",err) }
 target = filepath.Join(parent,filepath.Base(target))
 rel, err := filepath.Rel(source,target)
 if err != nil { return err }
 if rel == "." || (rel != ".." && !strings.HasPrefix(rel,".."+string(os.PathSeparator))) {
  return fmt.Errorf("Ziel darf nicht innerhalb der Quelle liegen")
 }
 if _, err := os.Lstat(target); err == nil { return fmt.Errorf("Ziel existiert bereits: %s",target) } else if !os.IsNotExist(err) { return err }
 var total int64
 err = filepath.WalkDir(source,func(path string,d fs.DirEntry,walkErr error) error {
  if walkErr != nil { return walkErr }; if err := ctx.Err(); err != nil { return err }
  info, err := d.Info(); if err != nil { return err }
  if !info.IsDir() && !info.Mode().IsRegular() { return fmt.Errorf("symbolischer Link oder Spezialdatei wird nicht kopiert: %s",path) }
  if info.Mode().IsRegular() { total += info.Size() }
  return nil
 })
 if err != nil { return err }
 p := operationProgress{total:total,status:"Album wird übertragen …"}
 emit(p)
 stage, err := os.MkdirTemp(parent,".audiolab-copy-*")
 if err != nil { return fmt.Errorf("Temporäres Ziel erstellen: %w",err) }
 defer func() {
  if stage != "" {
   // Directory modes may already have been restored to read-only.
   _ = filepath.WalkDir(stage,func(path string,d fs.DirEntry,e error) error { if e == nil && d.IsDir() { _ = os.Chmod(path,0700) }; return nil })
   if e := os.RemoveAll(stage); e != nil { err = errors.Join(err,fmt.Errorf("unvollständige Kopie unter %s konnte nicht entfernt werden: %w",stage,e)) }
  }
 }()
 type directory struct { path string; info fs.FileInfo }
 var dirs []directory
 err = filepath.WalkDir(source,func(path string,d fs.DirEntry,walkErr error) error {
  if walkErr != nil { return walkErr }; if err := ctx.Err(); err != nil { return err }
  rel, err := filepath.Rel(source,path); if err != nil { return err }
  dest := filepath.Join(stage,rel)
  info, err := d.Info(); if err != nil { return err }
  if info.IsDir() {
   if rel != "." { if err := os.Mkdir(dest,0700); err != nil { return err } }
   dirs = append(dirs,directory{dest,info}); return nil
  }
  if !info.Mode().IsRegular() { return fmt.Errorf("Quelle geändert: keine reguläre Datei: %s",path) }
  p.file = rel; emit(p)
  if err := copyRegularFile(ctx,path,dest,info,func(n int64) { p.current += n; emit(p) }); err != nil { return fmt.Errorf("%s: %w",rel,err) }
  return nil
 })
 if err != nil { return err }
 for i:=len(dirs)-1;i>=0;i-- {
  d:=dirs[i]
  if err := os.Chmod(d.path,d.info.Mode().Perm()); err != nil { return err }
  if err := os.Chtimes(d.path,d.info.ModTime(),d.info.ModTime()); err != nil { return err }
 }
 if err := ctx.Err(); err != nil { return err }
 if p.current != total { return fmt.Errorf("Quelle wurde während des Kopierens verändert (%d statt %d Bytes)",p.current,total) }
 // Claim the final path exclusively. Rename cannot overwrite an existing
 // nonempty directory; the empty placeholder is ours and removed on failure.
 if err := os.Mkdir(target,0700); err != nil { return fmt.Errorf("Ziel reservieren: %w",err) }
 if err := os.Rename(stage,target); err != nil {
  cleanupErr := os.Remove(target)
  return errors.Join(fmt.Errorf("Kopie veröffentlichen: %w",err),cleanupErr)
 }
 stage = ""
 p.status = "Kopieren abgeschlossen: "+target; emit(p)
 return nil
}

type countedWriter struct { ctx context.Context; dst io.Writer; count func(int64) }
func (w countedWriter) Write(data []byte) (int,error) {
 if err := w.ctx.Err(); err != nil { return 0,err }
 n,err := w.dst.Write(data); w.count(int64(n)); return n,err
}
func copyRegularFile(ctx context.Context,source,target string,info fs.FileInfo,count func(int64)) (err error) {
 src,err:=os.Open(source); if err!=nil { return err }; defer src.Close()
 opened,err:=src.Stat(); if err!=nil { return err }
 if !opened.Mode().IsRegular() || !os.SameFile(info,opened) { return fmt.Errorf("Quelldatei wurde ausgetauscht") }
 dst,err:=os.OpenFile(target,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600); if err!=nil { return err }
 defer func(){ err=errors.Join(err,dst.Close()) }()
 n,err:=io.CopyBuffer(countedWriter{ctx,dst,count},src,make([]byte,128*1024))
 if err!=nil { return err }; if n!=info.Size() { return fmt.Errorf("Dateigröße während des Kopierens verändert") }
 if err:=dst.Sync();err!=nil{return err}
 if err:=dst.Chmod(info.Mode().Perm());err!=nil{return err}
 return os.Chtimes(target,info.ModTime(),info.ModTime())
}
