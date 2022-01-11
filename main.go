package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	
	"github.com/coreos/etcd/clientv3"
)

/*
查询所有的keys，或者以某个前缀的keys
etcdctl get --prefix ""
etcdctl get --prefix "/my-prefix"
只列出keys，不显示值
etcdctl get --prefix --keys-only ""
etcdctl get --prefix --keys-only "/my-prefix"
*/

// ./etcd_sd -target-file /home/xiewj/container/iotmicro/monitor/prometheus/tgroups.json

type (
	instances map[string]string
	services  map[string]instances
)

type (
	Node struct {
		Id      string `json:"id"`
		Address string `json:"address"`
	}
	
	Service struct {
		Name  string `json:"name"`
		Nodes []Node `json:"nodes"`
	}
)

type TargetGroup struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

func main() {
	var (
		err         error
		path        = "/micro/registry"
		keyPattern  = "-web"
		targetFile  = flag.String("target-file", "tgroups.json", "the file that contains the target groups")
		dialTimeout = 5 * time.Second
		endpoints   = []string{"localhost:2379"}
		cli         *clientv3.Client
		response    *clientv3.GetResponse
		s           = &services{}
		service     = &Service{}
		watchChan   = make(clientv3.WatchChan)
	)
	
	flag.Parse()
	
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.Close()
	
	response, err = cli.Get(context.TODO(), path, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, val := range response.Kvs {
		if err = json.Unmarshal(val.Value, service); err != nil {
			log.Fatal(err)
		}
		
		if strings.Index(service.Name, keyPattern) == -1 {
			continue
		}
		
		// s.handle(service, s.update)
		s.update(service)
	}
	
	if err = s.persist(targetFile); err != nil {
		log.Fatalln(err)
	}
	
	for {
		watchChan = cli.Watch(context.Background(), path, clientv3.WithPrefix())
		for wresp := range watchChan {
			for _, ev := range wresp.Events {
				keys := strings.Split(string(ev.Kv.Key), "/")
				if strings.Index(keys[3], keyPattern) == -1 {
					continue
				}
				
				if fmt.Sprintf("%s", ev.Type) == "DELETE" {
					s.delete(keys[3], keys[4])
				} else if err = json.Unmarshal(ev.Kv.Value, service); err == nil {
					s.update(service)
				}
				
				fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
			
			if err = s.persist(targetFile); err != nil {
				log.Println(err)
			}
		}
	}
}

func (s services) handle(service *Service, h func(*Service)) {
	h(service)
}

func (s services) update(service *Service) {
	for _, node := range service.Nodes {
		insts, ok := s[service.Name]
		if !ok {
			insts = instances{}
		}
		insts[node.Id] = node.Address
		s[service.Name] = insts
	}
}

func (s services) delete(serviceName, nodeID string) {
	delete(s[serviceName], nodeID)
}

func (s services) persist(targetFile *string) error {
	var (
		err          error
		content      []byte
		targets      []string
		targetGroups []TargetGroup
		file         io.WriteCloser
	)
	
	for name, service := range s {
		targets = make([]string,0,len(service))
		for _, node := range service {
			targets = append(targets, node)
		}
		
		if len(targets) > 0 {
			targetGroups = append(targetGroups, TargetGroup{
				Targets: targets,
				Labels:  map[string]string{"project_name": name},
			})
		}
	}
	
	if content, err = json.Marshal(targetGroups); err != nil {
		return err
	}
	
	if file, err = create(*targetFile); err != nil {
		return err
	}
	defer file.Close()
	
	if _, err = file.Write(content); err != nil {
		return err
	}
	
	return nil
}
