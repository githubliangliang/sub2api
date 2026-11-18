package main
import (
 "context"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "net/http/httptest"
 "os"
 "os/exec"
 "strings"
 containerapi "github.com/docker/docker/api/types/container"
 imageapi "github.com/docker/docker/api/types/image"
 "github.com/docker/docker/client"
)
func main() {
 server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  w.Header().Set("Content-Type","application/json")
  switch {
  case strings.HasSuffix(r.URL.Path,"/_ping"):
   w.Header().Set("API-Version","1.51"); w.Header().Set("OSType","linux"); fmt.Fprint(w,"OK")
  case strings.HasSuffix(r.URL.Path,"/images/create"):
   fmt.Fprintln(w,`{"errorDetail":{"message":"fixture registry pull failed"},"error":"fixture registry pull failed"}`)
  case strings.HasSuffix(r.URL.Path,"/containers/create"):
   w.WriteHeader(404);fmt.Fprintln(w,`{"message":"No such image: redis:8.4-alpine"}`)
  default:
   w.WriteHeader(404);fmt.Fprintln(w,`{"message":"fixture endpoint not found"}`)
  }
 })); defer server.Close()
 host:="tcp://"+strings.TrimPrefix(server.URL,"http://")
 cli,err:=client.NewClientWithOpts(client.WithHost(host),client.WithVersion("1.51"));if err!=nil {panic(err)};defer cli.Close()
 stream,err:=cli.ImagePull(context.Background(),"redis:8.4-alpine",imageapi.PullOptions{});if err!=nil {panic(err)}
 _,copyErr:=io.Copy(io.Discard,stream);stream.Close()
 _,createErr:=cli.ContainerCreate(context.Background(),&containerapi.Config{Image:"redis:8.4-alpine"},nil,nil,nil,"probe")
 if copyErr!=nil || createErr==nil || !strings.Contains(createErr.Error(),"No such image") {panic("SDK failure mechanism did not reproduce")}
 command:=exec.Command("docker","pull","redis:8.4-alpine")
 for _,item:=range os.Environ() {
  name:=strings.SplitN(item,"=",2)[0]
  switch name {case "DOCKER_HOST","DOCKER_API_VERSION","DOCKER_TLS_VERIFY","DOCKER_CERT_PATH","HTTP_PROXY","HTTPS_PROXY","ALL_PROXY","http_proxy","https_proxy","all_proxy":continue}
  command.Env=append(command.Env,item)
 }
 command.Env=append(command.Env,"DOCKER_HOST="+host,"DOCKER_API_VERSION=1.51")
 output,commandErr:=command.CombinedOutput();if commandErr==nil {panic("docker CLI silently accepted pull error")}
 record:=map[string]any{"sdk_pull_stream_copy_error":copyErr,"container_create_error":createErr.Error(),"docker_cli_rejected_pull":true,"docker_cli_output":string(output)}
 encoded,_:=json.MarshalIndent(record,"","  ");fmt.Println(string(encoded))
}
