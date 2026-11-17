package main

import (
 "context"
 "sync"
 "time"

 tea "charm.land/bubbletea/v2"
)

type operationKind int
const (
 copyOperation operationKind = iota
 coverOperation
 ripOperation
)
type operationProgress struct {
 current, total int64
 file, status string
}
type operationProgressMsg struct {
 id uint64
 operationProgress
}
type operationResultMsg struct {
 id uint64
 operationProgress
 err error
 output string
}
type operationStream struct {
 progress chan operationProgressMsg
 result chan operationResultMsg
}
type operationState struct {
 id uint64
 kind operationKind
 title string
 operationProgress
 running, cancelling bool
 err error
 output string
 stream *operationStream
}
// The application owns workers, including when Bubble Tea exits unexpectedly.
type operationRuntime struct {
 ctx context.Context
 cancel context.CancelFunc
 workers sync.WaitGroup
}
func newOperationRuntime() *operationRuntime {
 ctx, cancel := context.WithCancel(context.Background())
 return &operationRuntime{ctx:ctx, cancel:cancel}
}
func (r *operationRuntime) close() { r.cancel(); r.workers.Wait() }

type operationWorker func(context.Context, func(operationProgress)) (string, error)

func waitOperation(s *operationStream) tea.Cmd {
 return func() tea.Msg {
  select {
  case result := <-s.result: return result
  case progress := <-s.progress: return progress
  }
 }
}

func (m model) startOperation(kind operationKind, title string, worker operationWorker) (tea.Model, tea.Cmd) {
 if m.operation.running { return m, nil }
 m.nextOperationID++
 id := m.nextOperationID
 ctx, cancel := context.WithCancel(m.runtime.ctx)
 m.cancelOperation = cancel
 stream := &operationStream{progress:make(chan operationProgressMsg,1), result:make(chan operationResultMsg,1)}
 m.operation = operationState{id:id,kind:kind,title:title,running:true,stream:stream,
  operationProgress:operationProgress{status:"Vorbereitung läuft …"}}
 m.screen = operationScreen
 m.message = ""
 m.runtime.workers.Add(1)
 go func() {
  defer m.runtime.workers.Done()
  defer cancel()
  var latest operationProgress
  var last time.Time
  emit := func(p operationProgress) {
   latest = p
   if time.Since(last) < 100*time.Millisecond { return }
   last = time.Now()
   // Only intermediate updates are coalesced. The final snapshot and error
   // have a separate, buffered mailbox and can never be dropped.
   select { case <-stream.progress: default: }
   stream.progress <- operationProgressMsg{id:id,operationProgress:p}
  }
  output, err := worker(ctx,emit)
  stream.result <- operationResultMsg{id:id,operationProgress:latest,err:err,output:output}
 }()
 return m, tea.Batch(m.spinner.Tick,waitOperation(stream))
}
