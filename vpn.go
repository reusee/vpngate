package main

import (
  "encoding/base64"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "os/exec"
  "strconv"
  "strings"
  "time"
)

func main() {
  if len(os.Args) < 2 {
    fmt.Printf("usage: %s [csv file]\n", os.Args[0])
    return
  }
  content, err := ioutil.ReadFile(os.Args[1])
  if err != nil {
    log.Fatal(err)
  }
  n := 0
  configs := make([]string, 0)
  for _, line := range strings.Split(string(content), "\n") {
    if len(line) == 0 || line[0] == '*' || line[0] == '#' {
      continue
    }
    fields := strings.Split(line, ",")
    speed, _ := strconv.Atoi(fields[4])
    if speed < 30000000 {
      continue
    }
    users, _ := strconv.Atoi(fields[9])
    if users < 5000 {
      continue
    }
    traffic, _ := strconv.Atoi(fields[10])
    if traffic < 100000000000 {
      continue
    }
    country := fields[6]
    sessions, _ := strconv.Atoi(fields[7])
    op := fields[12]
    fmt.Printf("%2d %s %3d sessions, %5dG traffic, %7d users, %2dM speed, %s\n", n, country, sessions, traffic/1000000000, users, speed/1000000, op)
    n++
    config := fields[14]
    configs = append(configs, config)
  }
  print("which?\n")
  fmt.Scanf("%d\n", &n)
  data, err := base64.StdEncoding.DecodeString(configs[n])
  if err != nil {
    log.Fatal(err)
  }
  tmpFile, err := ioutil.TempFile("", "")
  if err != nil {
    log.Fatal(err)
  }
  tmpFile.Write(data)
  tmpFile.Close()
  cmd := exec.Command("/usr/bin/env", "sudo", "openvpn", "--config", tmpFile.Name())
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    log.Fatal(err)
  }
  go io.Copy(os.Stdout, stdout)
  go func() {
    for {
      time.Sleep(time.Minute * 1)
      url := "http://www.vpngate.net/api/iphone/"
      resp, err := http.Get(url)
      if err != nil {
        continue
      }
      f, err := os.OpenFile("csv.tmp", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
      if err != nil {
        log.Fatal(err)
      }
      io.Copy(f, resp.Body)
      f.Close()
      resp.Body.Close()
      os.Rename("csv.tmp", "csv")
    }
  }()
  cmd.Run()
}
