package main

import (
	"encoding/json"
	"os"
	"path"
)

func main() {
	Files, err := os.ReadDir(path.Join(".", "files", "meshes"))

	if err != nil {
		panic(err.Error())
	}

	for _, File := range Files {
		Binary, err := os.ReadFile(path.Join(".", "files", "meshes", File.Name()))

		if err != nil {
			panic(err.Error())
		}

		Mesh := ReadMesh(Binary)
		MeshBytes, JSONErr := json.Marshal(Mesh)

		if JSONErr != nil {
			panic(JSONErr.Error())
		}

		WriteErr := os.WriteFile(path.Join(".", "files", "output", File.Name()), MeshBytes, 0677)

		if WriteErr != nil {
			panic(WriteErr.Error())
		}
	}
}
