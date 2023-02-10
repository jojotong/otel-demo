/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func Run(serverAddr string, interval int) error {
	rand.Seed(time.Now().Unix())
	for range time.NewTicker(time.Duration(interval) * time.Second).C {
		url := fmt.Sprintf("%s/users/%d", serverAddr, rand.Intn(10))
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		bts, _ := io.ReadAll(resp.Body)
		log.Printf("Get %s, resp: %s", url, string(bts))
	}
	return nil
}
