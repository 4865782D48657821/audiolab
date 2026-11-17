package main

import (
 "bytes"
 "context"
 "errors"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func writeTestFile(t *testing.T,path string,data []byte) {
 t.Helper()
 if err:=os.MkdirAll(filepath.Dir(path),0755);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(path,data,0640);err!=nil{t.Fatal(err)}
}
func assertNoCopyResidue(t *testing.T,target string) {
 t.Helper()
 if _,err:=os.Lstat(target);!os.IsNotExist(err){t.Fatalf("target still exists: %v",err)}
 entries,err:=filepath.Glob(filepath.Join(filepath.Dir(target),".audiolab-copy-*"));if err!=nil{t.Fatal(err)}
 if len(entries)!=0{t.Fatalf("staging directories remain: %v",entries)}
}
func TestCopyEmptyAndNestedDirectories(t *testing.T) {
 for _,nested:=range []bool{false,true} {
  t.Run(map[bool]string{false:"empty",true:"nested"}[nested],func(t *testing.T){
   source:=t.TempDir(); target:=filepath.Join(t.TempDir(),"album")
   if nested {writeTestFile(t,filepath.Join(source,"disc","track.flac"),[]byte("music"));if err:=os.Mkdir(filepath.Join(source,"empty"),0750);err!=nil{t.Fatal(err)}}
   var last operationProgress
   if err:=copyAlbum(context.Background(),source,target,func(p operationProgress){last=p});err!=nil{t.Fatal(err)}
   if nested {
    data,err:=os.ReadFile(filepath.Join(target,"disc","track.flac"));if err!=nil||string(data)!="music"{t.Fatalf("data %q, error %v",data,err)}
    if info,err:=os.Stat(filepath.Join(target,"empty"));err!=nil||!info.IsDir(){t.Fatalf("empty directory missing: %v",err)}
    if last.current!=5||last.total!=5{t.Fatalf("progress: %+v",last)}
   } else if last.current!=0||last.total!=0{t.Fatalf("empty progress: %+v",last)}
  })
 }
}
func TestCopyLargeFileProgressAndMetadata(t *testing.T) {
 source:=t.TempDir();target:=filepath.Join(t.TempDir(),"album")
 data:=bytes.Repeat([]byte("FLAC"),2*1024*1024)
 path:=filepath.Join(source,"large.flac");writeTestFile(t,path,data)
 stamp:=time.Unix(1700000000,0)
 if err:=os.Chtimes(path,stamp,stamp);err!=nil{t.Fatal(err)}
 updates:=0;var last int64
 err:=copyAlbum(context.Background(),source,target,func(p operationProgress){
  if p.current<last{t.Errorf("progress decreased")};last=p.current
  if p.current>0&&p.current<p.total{updates++}
 })
 if err!=nil{t.Fatal(err)}
 if updates<2{t.Fatalf("no intra-file updates: %d",updates)}
 got,err:=os.ReadFile(filepath.Join(target,"large.flac"));if err!=nil||!bytes.Equal(data,got){t.Fatalf("data mismatch: %v",err)}
 info,err:=os.Stat(filepath.Join(target,"large.flac"));if err!=nil{t.Fatal(err)}
 if info.Mode().Perm()!=0640||!info.ModTime().Equal(stamp){t.Fatalf("metadata: %v %v",info.Mode(),info.ModTime())}
}
func TestCopyRejectsInvalidPaths(t *testing.T) {
 source:=t.TempDir();writeTestFile(t,filepath.Join(source,"track"),[]byte("safe"))
 dest:=t.TempDir()
 tests:=[]struct{name,source,target string}{
  {"existing",source,dest},
  {"inside",source,filepath.Join(source,"copy")},
  {"self",source,source},
  {"missing",filepath.Join(source,"missing"),filepath.Join(dest,"album")},
  {"file source",filepath.Join(source,"track"),filepath.Join(dest,"album")},
  {"missing parent",source,filepath.Join(dest,"missing","album")},
 }
 for _,tt:=range tests {t.Run(tt.name,func(t *testing.T){if err:=copyAlbum(context.Background(),tt.source,tt.target,func(operationProgress){});err==nil{t.Fatal("expected error")}})}
 data,err:=os.ReadFile(filepath.Join(source,"track"));if err!=nil||string(data)!="safe"{t.Fatal("source modified")}
}
func TestCopyRejectsSymlinks(t *testing.T) {
 for _,kind:=range []string{"file","directory","dangling"} {
  t.Run(kind,func(t *testing.T){
   source:=t.TempDir();external:=t.TempDir();target:=filepath.Join(t.TempDir(),"album")
   linkTo:=external
   if kind=="file"{linkTo=filepath.Join(external,"track");writeTestFile(t,linkTo,[]byte("outside"))}
   if kind=="dangling"{linkTo=filepath.Join(external,"missing")}
   if err:=os.Symlink(linkTo,filepath.Join(source,"link"));err!=nil{t.Skip(err)}
   if err:=copyAlbum(context.Background(),source,target,func(operationProgress){});err==nil{t.Fatal("expected link rejection")}
   assertNoCopyResidue(t,target)
  })
 }
}
func TestCopyRejectsTargetInsideSourceThroughSymlink(t *testing.T) {
 source:=t.TempDir();parent:=t.TempDir();alias:=filepath.Join(parent,"alias")
 if err:=os.Symlink(source,alias);err!=nil{t.Skip(err)}
 target:=filepath.Join(alias,"album")
 if err:=copyAlbum(context.Background(),source,target,func(operationProgress){});err==nil{t.Fatal("expected rejection")}
 assertNoCopyResidue(t,filepath.Join(source,"album"))
}
func TestCopyCancellationCleansPartialData(t *testing.T) {
 source:=t.TempDir();parent:=t.TempDir();target:=filepath.Join(parent,"album")
 writeTestFile(t,filepath.Join(source,"track.flac"),bytes.Repeat([]byte("x"),1024*1024))
 ctx,cancel:=context.WithCancel(context.Background());defer cancel()
 err:=copyAlbum(ctx,source,target,func(p operationProgress){if p.current>0{cancel()}})
 if !errors.Is(err,context.Canceled){t.Fatalf("error = %v",err)}
 assertNoCopyResidue(t,target)
}
func TestCopyWriteFailureCleansPartialData(t *testing.T) {
 source:=t.TempDir();parent:=t.TempDir();target:=filepath.Join(parent,"album")
 writeTestFile(t,filepath.Join(source,"track.flac"),[]byte("data"))
 injected:=false
 err:=copyAlbum(context.Background(),source,target,func(p operationProgress){
  if p.file==""||injected{return};injected=true
  stages,e:=filepath.Glob(filepath.Join(parent,".audiolab-copy-*"));if e!=nil||len(stages)!=1{t.Fatalf("stages %v, %v",stages,e)}
  // A conflicting entry deterministically makes the exclusive file open fail,
  // including when tests are run as root or on a writable CI filesystem.
  if e:=os.Mkdir(filepath.Join(stages[0],p.file),0700);e!=nil{t.Fatal(e)}
 })
 if err==nil||!strings.Contains(err.Error(),"track.flac"){t.Fatalf("error = %v",err)}
 assertNoCopyResidue(t,target)
}
func TestCopyDestinationAppearingDuringCopyIsPreserved(t *testing.T) {
 source:=t.TempDir();target:=filepath.Join(t.TempDir(),"album")
 writeTestFile(t,filepath.Join(source,"track"),[]byte("music"))
 created:=false
 err:=copyAlbum(context.Background(),source,target,func(p operationProgress){
  if p.current>0&&!created{created=true;writeTestFile(t,filepath.Join(target,"existing"),[]byte("keep"))}
 })
 if err==nil{t.Fatal("expected existing-target error")}
 data,e:=os.ReadFile(filepath.Join(target,"existing"));if e!=nil||string(data)!="keep"{t.Fatalf("existing data changed: %v",e)}
}
