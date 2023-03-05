package main

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
)

type Vector3 [3]float32
type Vector2 [2]float32

type FaceStruct [3]uint16
type LODStruct []uint

type VerticesStruct struct {
	//VERSION 2.00
	Normal   Vector3
	Position Vector3
	UV       Vector3

	//VERSION 3.00
	Tangent [4]byte
	Color   [4]byte
}

type MeshHeaderStruct struct {
	Version string

	//VERSION 2.00
	Sizeof_MeshHeader uint16
	Sizeof_Vertex     byte
	Sizeof_Face       byte

	NumVerts uint16
	NumFaces uint16

	//VERSION 3.00
	Sizeof_LOD uint16
	NumLODs    uint16

	//VERSION 4.00
	LodType  uint16
	NumBones uint16

	Sizeof_boneNamesBuffer uint16
	NumSubsets             uint16

	NumHighQualityLODs byte
	Unused             byte
}

type MeshStruct struct {
	Valid  bool
	Header MeshHeaderStruct

	Vertices []VerticesStruct
	Faces    []FaceStruct
	LODs     []uint
}

func ReadBytes(Reader *bytes.Reader, BytesToRead int) []byte {
	Bytes := make([]byte, BytesToRead)

	for i := 0; i < BytesToRead; i++ {
		Byte, _ := Reader.ReadByte()
		Bytes[i] = Byte
	}

	return Bytes
}

func GetMeshVersion(Mesh []byte) string {
	return string(Mesh)[8:12]
}

func VersionToFloat(Version string) float64 {
	VersionFloat, _ := strconv.ParseFloat(Version, 64)
	return VersionFloat
}

func ReadASCIIMesh(Mesh []byte) MeshStruct {
	MeshData := string(Mesh)
	Data := strings.Split(MeshData, "\n")

	Version := GetMeshVersion(Mesh)
	NumFaces, _ := strconv.Atoi(strings.ReplaceAll(Data[1], "\r", ""))

	var Vertices VerticesStruct
	Progress := 0

	AllVertices := make([]VerticesStruct, NumFaces*3)

	for i, Vector := range strings.Split(Data[2][1:][:len(Data[2])-1], "][") {
		Coordinates := strings.Split(Vector, ",")

		X, _ := strconv.ParseFloat(Coordinates[0], 32)
		Y, _ := strconv.ParseFloat(Coordinates[1], 32)
		Z, _ := strconv.ParseFloat(Coordinates[2], 32)

		if Version == "1.00" {
			X /= 2
			Y /= 2
			X /= 2
		}

		Vector3 := Vector3{float32(X), float32(Y), float32(Z)}
		Progress++

		if Progress == 3 {
			Vertices.UV = Vector3
			AllVertices[((i+1)/3)-1] = Vertices
			Vertices = VerticesStruct{}

			Progress = 0
		} else if Progress == 2 {
			Vertices.Normal = Vector3
		} else {
			Vertices.Position = Vector3
		}
	}

	return MeshStruct{Header: MeshHeaderStruct{Version: Version, NumFaces: uint16(NumFaces), NumVerts: uint16(len(AllVertices))}, Vertices: AllVertices, Valid: true}
}

func ReadBinaryMesh(MeshBytes []byte) MeshStruct {
	Version := GetMeshVersion(MeshBytes)
	VersionFloat := VersionToFloat(Version)

	Mesh := MeshStruct{Valid: true}
	Reader := bytes.NewReader(MeshBytes)

	ReadBytes(Reader, 13) //VERSION HEADER

	binary.Read(Reader, binary.LittleEndian, &Mesh.Header.Sizeof_MeshHeader)

	if VersionFloat >= 4.00 {
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.LodType)
	} else {
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.Sizeof_Vertex)
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.Sizeof_Face)
	}

	if VersionFloat >= 3.00 {
		if VersionFloat < 4.00 {
			binary.Read(Reader, binary.LittleEndian, &Mesh.Header.Sizeof_LOD)
		}
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.NumLODs)
	}

	binary.Read(Reader, binary.LittleEndian, &Mesh.Header.NumVerts)
	ReadBytes(Reader, 2)
	binary.Read(Reader, binary.LittleEndian, &Mesh.Header.NumFaces)
	ReadBytes(Reader, 2)

	if VersionFloat >= 4.00 {
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.Sizeof_boneNamesBuffer)
		ReadBytes(Reader, 2)
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.NumSubsets)
		ReadBytes(Reader, 2)
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.NumHighQualityLODs)
		binary.Read(Reader, binary.LittleEndian, &Mesh.Header.Unused)
	}

	Vertices := make([]VerticesStruct, int(Mesh.Header.NumVerts))

	for i := 0; i < int(Mesh.Header.NumVerts); i++ {
		Vertex := VerticesStruct{}
		var UV Vector2

		binary.Read(Reader, binary.LittleEndian, &Vertex.Position)
		binary.Read(Reader, binary.LittleEndian, &Vertex.Normal)
		binary.Read(Reader, binary.LittleEndian, &UV)

		Vertex.UV = Vector3{UV[0], UV[1], 0}

		binary.Read(Reader, binary.LittleEndian, &Vertex.Tangent)

		if (Version == "2.00" || Version == "3.00") && Mesh.Header.Sizeof_Vertex == 36 {
			Vertex.Color = [4]byte{255, 255, 255, 255}
		} else {
			binary.Read(Reader, binary.LittleEndian, &Vertex.Color)
		}

		Vertices[i] = Vertex
	}

	Faces := make([]FaceStruct, int(Mesh.Header.NumFaces))

	for i := 0; i < int(Mesh.Header.NumFaces); i++ {
		Face := FaceStruct{}
		binary.Read(Reader, binary.LittleEndian, &Face)
		Faces[i] = Face
	}

	if VersionFloat >= 3.00 {
		LODs := make(LODStruct, int(Mesh.Header.NumLODs))
		binary.Read(Reader, binary.LittleEndian, &LODs)

		Mesh.LODs = LODs
	}

	Mesh.Vertices = Vertices
	Mesh.Faces = Faces

	Mesh.Header.Version = Version

	return Mesh
}

func ReadMesh(Mesh []byte) MeshStruct {
	Version := GetMeshVersion(Mesh)

	if Version == "1.00" || Version == "1.01" {
		return ReadASCIIMesh(Mesh)
	} else {
		return ReadBinaryMesh(Mesh)
	}
}
