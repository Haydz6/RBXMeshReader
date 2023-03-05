package main

import (
	"encoding/json"
	"io/ioutil"
	"path"
)

func main() {
	Files, err := ioutil.ReadDir(path.Join(".", "files", "meshes"))

	if err != nil {
		panic(err.Error())
	}

	for _, File := range Files {
		Binary, err := ioutil.ReadFile(path.Join(".", "files", "meshes", File.Name()))

		if err != nil {
			panic(err.Error())
		}

		Mesh := ReadMesh(Binary)
		MeshBytes, JSONErr := json.Marshal(Mesh)

		if JSONErr != nil {
			panic(JSONErr.Error())
		}

		WriteErr := ioutil.WriteFile(path.Join(".", "files", "output", File.Name()), MeshBytes, 0677)

		if WriteErr != nil {
			panic(WriteErr.Error())
		}
	}
}
