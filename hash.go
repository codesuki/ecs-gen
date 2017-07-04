package main

import (
	"encoding/hex"
	"hash/fnv"
	"io"
	"log"
	"os"
	"sort"
)

func getHash(containers []*container) string {
	var defaultHash string
	hasher := fnv.New32a()

	//add template file to hash
	file, err := os.Open(*templateFile)
	if err != nil {
		log.Println(err)
		return defaultHash
	}
	defer file.Close()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Println(err)
		return defaultHash
	}

	//add elements from the Container struct to hash
	//any changes to the Container struct will need to be reflected here
	//if they are to be part of the hash
	for _, container := range containers {
		hasher.Write([]byte(container.Host))
		hasher.Write([]byte(container.Port))
		hasher.Write([]byte(container.Address))

		//Environment vars must be sorted, or subsequent iterations will be out of order
		//and throw off the hash
		keys := make([]string, len(container.Env))
		for k := range container.Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			hasher.Write([]byte(container.Env[k] + k))
		}
	}

	//generate hash
	return hex.EncodeToString(hasher.Sum(nil))
}
