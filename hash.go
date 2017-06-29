package main

import (
	"encoding/hex"
	"hash/fnv"
	"io"
	"log"
	"os"
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
		for envKey, envVal := range container.Env {
			hasher.Write([]byte(envKey + envVal))
		}
	}

	//generate hash
	return hex.EncodeToString(hasher.Sum(nil))
}
