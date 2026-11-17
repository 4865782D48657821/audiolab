package main

import (
 "context"
 "fmt"
 "os"
 "path/filepath"
 "strings"
)

func ripCD(ctx context.Context,emit func(operationProgress)) (string,error) {
 emit(operationProgress{status:"CD-Verarbeitung läuft …"})
 return runProcess(ctx,func(line string){ emit(operationProgress{status:line}) },"abcde","-o","flac")
}

func embedCover(ctx context.Context,albumPath string,emit func(operationProgress)) (string,error) {
 coverPath:=""
 for _,name:=range []string{"cover.jpg","cover.png","folder.jpg"} {
  path:=filepath.Join(albumPath,name)
  if info,err:=os.Lstat(path);err==nil && info.Mode().IsRegular(){ coverPath=path;break }
 }
 if coverPath=="" {return "",fmt.Errorf("kein cover.jpg, cover.png oder folder.jpg in %s gefunden",albumPath)}
 entries,err:=os.ReadDir(albumPath);if err!=nil{return "",err}
 var files []string
 for _,entry:=range entries {
  if strings.EqualFold(filepath.Ext(entry.Name()),".flac") {
   info,err:=entry.Info();if err!=nil{return "",err}
   if !info.Mode().IsRegular(){return "",fmt.Errorf("keine reguläre FLAC-Datei: %s",entry.Name())}
   files=append(files,filepath.Join(albumPath,entry.Name()))
  }
 }
 if len(files)==0 {return "",fmt.Errorf("keine FLAC-Dateien in %s gefunden",albumPath)}
 p:=operationProgress{total:int64(len(files)),status:"Cover wird eingebettet …"}
 emit(p)
 skipped:=0
 for _,file:=range files {
  if err:=ctx.Err();err!=nil{return "",err}
  p.file=filepath.Base(file);emit(p)
  output,err:=runProcess(ctx,nil,"metaflac","--list","--block-type=PICTURE",file)
  if err!=nil{return output,fmt.Errorf("%s: Cover prüfen: %w",p.file,err)}
  if strings.TrimSpace(output)!="" {
   skipped++
   p.status="Vorhandenes Cover beibehalten: "+p.file
  } else {
   output,err=embedCoverFile(ctx,file,coverPath)
   if err!=nil{return output,fmt.Errorf("%s: %w",p.file,err)}
   p.status="Cover eingebettet: "+p.file
  }
  p.current++;emit(p)
 }
 return fmt.Sprintf("%d Dateien abgeschlossen; %d mit vorhandenem Cover unverändert. Bereits abgeschlossene Dateien bleiben bei Abbruch erhalten.",len(files),skipped),nil
}

// Edit a private copy, so a killed metaflac never damages the current track.
func embedCoverFile(ctx context.Context,file,cover string) (string,error) {
 info,err:=os.Lstat(file);if err!=nil{return "",err}
 temp,err:=os.CreateTemp(filepath.Dir(file),".audiolab-cover-*.flac");if err!=nil{return "",err}
 path:=temp.Name()
 if err:=temp.Close();err!=nil{ _=os.Remove(path);return "",err }
 defer os.Remove(path)
 // copyRegularFile creates exclusively; remove only our own empty temp file.
 if err:=os.Remove(path);err!=nil{return "",err}
 if err:=copyRegularFile(ctx,file,path,info,func(int64){});err!=nil{return "",err}
 // Editing a read-only original is safe in our private temporary file.
 if err:=os.Chmod(path,0600);err!=nil{return "",err}
 output,err:=runProcess(ctx,nil,"metaflac","--import-picture-from="+cover,path)
 if err!=nil{return output,err}
 if err:=ctx.Err();err!=nil{return output,err}
 current,err:=os.Lstat(file);if err!=nil{return output,err}
 if !os.SameFile(info,current)||current.Size()!=info.Size()||!current.ModTime().Equal(info.ModTime()){return output,fmt.Errorf("Quelldatei wurde während der Bearbeitung verändert")}
 if err:=os.Chmod(path,info.Mode().Perm());err!=nil{return output,err}
 if err:=os.Chtimes(path,info.ModTime(),info.ModTime());err!=nil{return output,err}
 return output,os.Rename(path,file)
}
